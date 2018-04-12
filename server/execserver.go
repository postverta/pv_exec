package server

import (
	"bytes"
	"fmt"
	pb "github.com/postverta/pv_exec/proto/exec"
	context "golang.org/x/net/context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

type Task struct {
	Id    uint64
	Name  string
	Input []byte
	Env   []string

	InScopes  map[string]bool
	OutScopes map[string]bool
	RunPath   string
	Server    *ExecServer

	Started  bool
	Output   []byte
	Finished bool
}

func (t *Task) Run() {
	runScript := path.Join(t.RunPath, "run")
	cmd := exec.Command(runScript)
	cmd.Dir = t.RunPath

	// Add the env vars on top of system's env var
	cmd.Env = make([]string, len(os.Environ()))
	copy(cmd.Env, os.Environ())
	cmd.Env = append(cmd.Env, t.Env...)

	cmd.Stdin = bytes.NewBuffer(t.Input)
	outputBuf := new(bytes.Buffer)

	var logCmd *exec.Cmd
	var pr *io.PipeReader
	var pw *io.PipeWriter
	if _, err := os.Stat(path.Join(t.RunPath, "log")); err == nil {
		logCmd = exec.Command(path.Join(t.RunPath, "log"))
		logCmd.Env = make([]string, len(cmd.Env))
		copy(logCmd.Env, cmd.Env)

		pr, pw = io.Pipe()
		logCmd.Stdin = pr
		logCmd.Stdout = outputBuf
		cmd.Stdout = pw
		logCmd.Start()
	} else {
		cmd.Stdout = outputBuf
	}

	cmd.Stderr = nil
	log.Println("Start task " + t.Name)
	err := cmd.Run()
	if err != nil {
		log.Println("Task "+t.Name+" failed:", err)
	} else {
		log.Println("Task " + t.Name + " finished")
	}

	if logCmd != nil {
		// Close the writer half of the pipe
		pw.Close()
		// Wait for the log script to quit
		err = logCmd.Wait()
		if err != nil {
			log.Println("Log for task "+t.Name+" failed:", err)
		}
		pr.Close()
	}

	t.Output = outputBuf.Bytes()
	t.Finished = true

	// kick the server
	t.Server.kick <- true
}

type ExecServer struct {
	ConfigRootDirs []string

	mutex sync.Mutex

	nextTaskId uint64

	tasks map[uint64]*Task

	taskFinishChan map[uint64]chan bool

	scopeReadQueue        map[string][]uint64
	scopeReadRunningTasks map[string]map[uint64]bool
	scopeWriteQueue       map[string][]uint64
	scopeWriteRunningTask map[string]uint64

	kick    chan bool
	stop    chan bool
	stopped chan bool
}

func NewExecServer(configRootDirs []string) *ExecServer {
	return &ExecServer{
		ConfigRootDirs:        configRootDirs,
		tasks:                 make(map[uint64]*Task),
		scopeReadQueue:        make(map[string][]uint64),
		scopeReadRunningTasks: make(map[string]map[uint64]bool),
		scopeWriteQueue:       make(map[string][]uint64),
		scopeWriteRunningTask: make(map[string]uint64),
		taskFinishChan:        make(map[uint64]chan bool),
		kick:                  make(chan bool, 256),
		stop:                  make(chan bool, 1),
		stopped:               make(chan bool, 1),
	}
}

func (s *ExecServer) Run() {
	var stop bool = false
	for {
		select {
		case <-s.stop:
			s.stopped <- true
			stop = true
			break
		case <-s.kick:
			break
		}

		s.mutex.Lock()

		// Clean up finished tasks
		for id, task := range s.tasks {
			if task.Finished {
				for scope, runningTasks := range s.scopeReadRunningTasks {
					if !task.InScopes[scope] {
						continue
					}

					delete(runningTasks, id)

					if len(runningTasks) == 0 {
						delete(s.scopeReadRunningTasks, scope)
					}
				}
				for scope, runningTaskId := range s.scopeWriteRunningTask {
					if !task.OutScopes[scope] {
						continue
					}

					if runningTaskId != id {
						log.Fatal("BUG: running task is inconsistent")
					}

					delete(s.scopeWriteRunningTask, scope)
				}

				if _, found := s.taskFinishChan[id]; found {
					s.taskFinishChan[id] <- true
				}

				delete(s.taskFinishChan, id)
				delete(s.tasks, id)
			}
		}

		// If we don't have any task, and we have been told to stop, exit
		if len(s.tasks) == 0 && stop {
			s.stopped <- true
			s.mutex.Unlock()
			return
		}

		// Try to start new tasks
		for _, task := range s.tasks {
			if !task.Started {
				// Schedule new task if possible
				executable := true
				for scope, _ := range task.InScopes {
					if _, found := s.scopeWriteRunningTask[scope]; found {
						// Someone is still writing to this scope
						executable = false
						break
					}

					if s.scopeReadQueue[scope][0] != task.Id {
						// We are not the first in the queue
						executable = false
						break
					}
				}

				for scope, _ := range task.OutScopes {
					if _, found := s.scopeReadRunningTasks[scope]; found {
						// Someone is still reading from this scope
						executable = false
						break
					}

					if _, found := s.scopeWriteRunningTask[scope]; found {
						// Someone is still writing to this scope
						executable = false
						break
					}

					if s.scopeWriteQueue[scope][0] != task.Id {
						// We are not the first in the queue
						executable = false
						break
					}
				}

				if executable {
					for scope, _ := range task.InScopes {
						if _, found := s.scopeReadRunningTasks[scope]; !found {
							s.scopeReadRunningTasks[scope] = make(map[uint64]bool)
						}

						s.scopeReadRunningTasks[scope][task.Id] = true
						s.scopeReadQueue[scope] = s.scopeReadQueue[scope][1:]
					}

					for scope, _ := range task.OutScopes {
						s.scopeWriteRunningTask[scope] = task.Id
						s.scopeWriteQueue[scope] = s.scopeWriteQueue[scope][1:]
					}

					task.Started = true
					go task.Run()
				}
			}
		}

		s.mutex.Unlock()
	}
}

func (s *ExecServer) NewTask(taskName string) *Task {
	found := false
	task := &Task{
		Name:      taskName,
		InScopes:  make(map[string]bool),
		OutScopes: make(map[string]bool),
		Server:    s,
	}

	for _, rootDir := range s.ConfigRootDirs {
		if stat, err := os.Stat(path.Join(rootDir, taskName)); err == nil && stat.IsDir() {
			if f, err := os.Open(path.Join(rootDir, taskName, "input_scopes")); err == nil {
				data, _ := ioutil.ReadAll(f)
				for _, scope := range strings.Split(string(data), "\n") {
					if scope != "" {
						task.InScopes[scope] = true
					}
				}
				f.Close()
			}
			if f, err := os.Open(path.Join(rootDir, taskName, "output_scopes")); err == nil {
				data, _ := ioutil.ReadAll(f)
				for _, scope := range strings.Split(string(data), "\n") {
					if scope != "" {
						task.OutScopes[scope] = true
					}
				}
				f.Close()
			}
			task.RunPath = path.Join(rootDir, taskName)
			found = true
			break
		}
	}

	if found {
		return task
	} else {
		return nil
	}
}

func (s *ExecServer) Exec(c context.Context, req *pb.ExecReq) (*pb.ExecResp, error) {
	task := s.NewTask(req.TaskName)
	if task == nil {
		log.Println("Ignore unknown task " + req.TaskName)
		return &pb.ExecResp{
			Completed: true,
			Data:      []byte{},
		}, nil
	}

	task.Input = make([]byte, len(req.Data))
	copy(task.Input, req.Data)

	task.Env = make([]string, len(req.KeyValues))
	for i, kv := range req.KeyValues {
		task.Env[i] = fmt.Sprintf("%s=%s", kv.Key, kv.Value)
	}

	var finishChan chan bool
	if req.WaitForCompletion {
		finishChan = make(chan bool, 1)
	}

	s.EnqueueTask(task, finishChan)

	resp := &pb.ExecResp{}
	if req.WaitForCompletion {
		<-finishChan
		resp.Data = task.Output
		resp.Completed = true
	}

	return resp, nil
}

func (s *ExecServer) EnqueueTask(task *Task, finishChan chan bool) {
	s.mutex.Lock()
	task.Id = s.nextTaskId
	s.nextTaskId += 1

	s.tasks[task.Id] = task
	if finishChan != nil {
		s.taskFinishChan[task.Id] = finishChan
	}

	for scope, _ := range task.InScopes {
		if _, found := s.scopeReadQueue[scope]; !found {
			s.scopeReadQueue[scope] = make([]uint64, 0)
		}
		s.scopeReadQueue[scope] = append(s.scopeReadQueue[scope], task.Id)
	}

	for scope, _ := range task.OutScopes {
		if _, found := s.scopeWriteQueue[scope]; !found {
			s.scopeWriteQueue[scope] = make([]uint64, 0)
		}
		s.scopeWriteQueue[scope] = append(s.scopeWriteQueue[scope], task.Id)
	}

	s.mutex.Unlock()

	// kick the main processing loop
	s.kick <- true
}

func (s *ExecServer) Shutdown() {
	s.stop <- true
	<-s.stopped
}
