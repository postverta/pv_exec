package worktree

import (
	"bufio"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/postverta/lazytree/fuse"
	"github.com/postverta/lazytree/tree/mem"
	"github.com/satori/go.uuid"
	"io"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type WorktreeManager struct {
	// Worktree to be loaded from. Useful for cloning. If empty, create an
	// empty worktree
	sourceWorktreeId string

	// Worktree to be written to
	worktreeId string

	// Mount point
	mountPoint string

	// Autosave interval
	autoSaveInterval time.Duration

	blobContainer *storage.Container

	rootDir         *mem.Dir
	context         *mem.Context
	generation      uint64
	savedGeneration uint64
	saveMutex       sync.Mutex
	autoSaveTimer   *time.Timer

	fuseFs *fuse.FuseFileSystem
}

func NewWorktreeManager(accountName string, accountKey string, container string, sourceWorktreeId string, worktreeId string, mountPoint string, autoSaveInterval time.Duration) (*WorktreeManager, error) {
	// Don't use HTTPS for lower overhead
	client, err := storage.NewClient(accountName, accountKey, storage.DefaultBaseURL, storage.DefaultAPIVersion, false)
	if err != nil {
		return nil, err
	}
	blobCli := client.GetBlobService()
	blobContainer := blobCli.GetContainerReference(container)

	manager := &WorktreeManager{
		blobContainer:    blobContainer,
		sourceWorktreeId: sourceWorktreeId,
		worktreeId:       worktreeId,
		mountPoint:       mountPoint,
		autoSaveInterval: autoSaveInterval,
	}
	manager.context = &mem.Context{
		Notifier: manager,
	}
	return manager, nil
}

func getLatestSnapshot(container *storage.Container, blobName string) (*storage.Blob, error) {
	listBlobResponse, err := container.ListBlobs(storage.ListBlobsParameters{
		Prefix: blobName,
		Include: &storage.IncludeBlobDataset{
			Snapshots: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(listBlobResponse.Blobs) == 0 {
		return nil, fmt.Errorf("Cannot find the blob %s", blobName)
	}

	var latest time.Time
	var blobRef *storage.Blob
	for _, blob := range listBlobResponse.Blobs {
		if blob.Snapshot.IsZero() {
			// not a snapshot
			continue
		}
		if blob.Snapshot.After(latest) {
			latest = blob.Snapshot
			blobRef = &blob
		}
	}

	if blobRef == nil {
		return nil, fmt.Errorf("No snapshot is available for blob %s", blobName)
	}

	return blobRef, nil
}

type bySnapshotTime []storage.Blob

func (bst bySnapshotTime) Len() int {
	return len(bst)
}
func (bst bySnapshotTime) Swap(i, j int) {
	bst[i], bst[j] = bst[j], bst[i]
}
func (bst bySnapshotTime) Less(i, j int) bool {
	return bst[i].Snapshot.Before(bst[j].Snapshot)
}

func updateSnapshots(container *storage.Container, blobName string) error {
	listBlobResponse, err := container.ListBlobs(storage.ListBlobsParameters{
		Prefix: blobName,
		Include: &storage.IncludeBlobDataset{
			Snapshots: true,
		},
	})
	if err != nil {
		return err
	}

	snapshots := make([]storage.Blob, 0)

	for _, blob := range listBlobResponse.Blobs {
		if blob.Snapshot.IsZero() {
			// Not a snapshot
			continue
		}
		snapshots = append(snapshots, blob)
	}

	if len(snapshots) == 0 {
		return nil
	}

	sort.Sort(bySnapshotTime(snapshots))

	// Always keep the newest one
	snapshots = snapshots[:len(snapshots)-1]

	nowTime := time.Now()
	buckets := []time.Time{
		nowTime.Add(-time.Hour * 24 * 7),
		nowTime.Add(-time.Hour * 24),
		nowTime.Add(-time.Hour),
	}
	bucketFilled := make([]bool, len(buckets))
	for _, blob := range snapshots {
		keep := false
		for bi, bucket := range buckets {
			if bucketFilled[bi] {
				continue
			}
			if blob.Snapshot.After(bucket) {
				bucketFilled[bi] = true
				keep = true
				break
			}
		}
		if !keep {
			err := blob.Delete(&storage.DeleteBlobOptions{
				Snapshot: &blob.Snapshot,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (wt *WorktreeManager) Load() error {
	if wt.rootDir != nil {
		log.Panic("WorktreeManager cannot be loaded twice!")
	}

	if wt.sourceWorktreeId == "" {
		wt.rootDir = mem.NewDir(wt.context, uuid.NewV4().String())
	} else {
		blobRef, err := getLatestSnapshot(wt.blobContainer, wt.sourceWorktreeId)
		if err != nil {
			return err
		}

		br, err := blobRef.Get(nil)
		if err != nil {
			return err
		}

		startTime := time.Now()
		wt.rootDir, err = mem.Deserialize(wt.context, br, false)
		if err != nil {
			return err
		}
		log.Printf("[INFO] time to deserialize: %f\n", time.Since(startTime).Seconds())
	}

	if wt.sourceWorktreeId != wt.worktreeId {
		wt.generation = 1
	}

	return nil
}

func (wt *WorktreeManager) Save() error {
	// There can only be one save running at any time
	wt.saveMutex.Lock()
	defer wt.saveMutex.Unlock()

	if wt.savedGeneration == wt.generation {
		// No new changes
		return nil
	}

	currentGeneration := wt.generation
	blobRef := wt.blobContainer.GetBlobReference(wt.worktreeId)

	pr, pw := io.Pipe()
	// Buffer the pipe
	writer := bufio.NewWriterSize(pw, 8*1024*1024)

	errChan := make(chan error, 1)
	go func() {
		defer pw.Close()
		startTime := time.Now()
		err := mem.Serialize(wt.rootDir, writer)
		log.Printf("[INFO] time to serialize: %f\n", time.Since(startTime).Seconds())
		if err != nil {
			errChan <- err
			return
		}
		err = writer.Flush()
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	err := blobRef.CreateBlockBlobFromReader(pr, nil)
	if err != nil {
		return err
	}

	// Check for any write error. It could mean interrupted writes and internal bug,
	// which could cause corrupted data.
	err = <-errChan
	if err != nil {
		return err
	}

	// Take a new snapshot of the blob, which is the new source of reading
	_, err = blobRef.CreateSnapshot(nil)
	if err != nil {
		return err
	}

	// Clean up old snapshots
	err = updateSnapshots(wt.blobContainer, wt.worktreeId)
	if err != nil {
		return err
	}

	// Mark the generation ID
	wt.savedGeneration = currentGeneration
	return nil
}

func (wt *WorktreeManager) StartAutoSave() {
	if wt.autoSaveTimer != nil {
		return
	}

	wt.autoSaveTimer = time.AfterFunc(wt.autoSaveInterval, func() {
		err := wt.Save()
		if err != nil {
			log.Println("[ERROR] auto-save failed:", err)
		}
		wt.autoSaveTimer.Reset(wt.autoSaveInterval)
	})
}

func (wt *WorktreeManager) StopAutoSave() {
	if wt.autoSaveTimer == nil {
		return
	}

	wt.autoSaveTimer.Stop()
	wt.autoSaveTimer = nil
}

func (wt *WorktreeManager) MountFuse() error {
	fuseFs := fuse.NewFuseFileSystem(wt.rootDir)
	err := fuseFs.Mount(wt.mountPoint, false)
	if err != nil {
		return err
	}
	wt.fuseFs = fuseFs

	return nil
}

func (wt *WorktreeManager) ServeFuse() {
	// Will not return
	wt.fuseFs.Serve()
}

func (wt *WorktreeManager) WaitFuse() error {
	return wt.fuseFs.WaitMount()
}

func (wt *WorktreeManager) StopFuse() error {
	return wt.fuseFs.Unmount()
}

// Implement the Notifier interface
func (wt *WorktreeManager) NodeChanged(id string) {
	atomic.AddUint64(&(wt.generation), 1)
}
