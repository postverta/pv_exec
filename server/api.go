package server

import (
	"fmt"
	pbe "github.com/postverta/pv_exec/proto/exec"
	pbp "github.com/postverta/pv_exec/proto/process"
	pbw "github.com/postverta/pv_exec/proto/worktree"
	"github.com/postverta/pv_exec/worktree"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start(config *Config) error {
	worktreeManager, err := worktree.NewWorktreeManager(
		config.AccountName,
		config.AccountKey,
		config.Container,
		config.SourceWorktreeId,
		config.WorktreeId,
		config.MountPoint,
		time.Duration(config.AutoSaveIntervalSeconds)*time.Second)
	if err != nil {
		return err
	}

	err = worktreeManager.Load()
	if err != nil {
		return err
	}

	worktreeManager.StartAutoSave()
	err = worktreeManager.MountFuse()
	if err != nil {
		return err
	}

	go func() {
		worktreeManager.ServeFuse()
	}()

	err = worktreeManager.WaitFuse()
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.MaxSendMsgSize(10*1024*1024),
		grpc.MaxRecvMsgSize(10*1024*1024),
	)
	execServer := NewExecServer(config.ExecConfigRoots)
	processServer := NewProcessServer()
	worktreeServer := NewWorktreeServer(worktreeManager)

	// Set up the signal handler for SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		grpcServer.GracefulStop()
	}()

	// Start the exec server processing loop
	go execServer.Run()

	log.Printf("Server listening on %s:%d\n", config.Host, config.Port)
	pbe.RegisterExecServiceServer(grpcServer, execServer)
	pbp.RegisterProcessServiceServer(grpcServer, processServer)
	pbw.RegisterWorktreeServiceServer(grpcServer, worktreeServer)
	grpcServer.Serve(lis)

	// If the code reaches this point, the GRPC server is shut down, and no
	// new request can be received. Wait for all existing process/exec
	// tasks to finish.
	processServer.Shutdown()
	execServer.Shutdown()

	// Unmount file system
	err = worktreeManager.StopFuse()
	if err != nil {
		// This is likely a bug, but anyway we will clean up the mountpoints
		// when the container is shut down. Log and move on
		log.Println("[ERROR] Cannot stop fuse:", err)
	}

	worktreeManager.StopAutoSave()

	// Just in case, issue a final save
	startTime := time.Now()
	err = worktreeManager.Save()
	log.Printf("[INFO] Time to save: %f\n", time.Since(startTime).Seconds())
	if err != nil {
		log.Println("[ERROR] Cannot save data:", err)
	}
	return nil
}
