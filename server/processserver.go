package server

import (
	"fmt"
	pb "github.com/postverta/pv_exec/proto/process"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type RestartRequest struct {
	StartCmd []string
	EnvVars  []*pb.KeyValuePair
}

type Process struct {
	Name        string
	RunPath     string
	State       pb.ProcessState
	Port        uint32
	RestartChan chan *RestartRequest
	StopChan    chan bool
}

type ProcessServer struct {
	process map[string]*Process
	mutex   sync.Mutex
}

func NewProcessServer() *ProcessServer {
	return &ProcessServer{
		process: make(map[string]*Process),
	}
}

func startProxy(listenPort uint32, forwardPort uint32) (chan bool, error) {
	stopChan := make(chan bool, 1)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			c := make(chan net.Conn, 1)
			go func() {
				conn, err := ln.Accept()
				if err == nil {
					c <- conn
				} else {
					c <- nil
				}
			}()

			select {
			case conn1 := <-c:
				if conn1 == nil {
					break
				}

				conn2, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", forwardPort))
				if err != nil {
					conn1.Close()
					continue
				}

				go func() {
					io.Copy(conn1, conn2)
					conn1.Close()
					conn2.Close()
				}()

				go func() {
					io.Copy(conn2, conn1)
					conn1.Close()
					conn2.Close()
				}()
			case <-stopChan:
				ln.Close()
				return
			}
		}
	}()

	log.Printf("Start proxy %d->%d\n", listenPort, forwardPort)

	return stopChan, nil
}

func monitorPorts(pgid int, portChan chan uint32, stopChan chan bool) {
	for {
		// This is ugly, but is the simplest solution
		cmd := exec.Command(
			"lsof",
			"-Pan",
			"-g",
			fmt.Sprintf("%d", pgid),
			"-iTCP",
			"-sTCP:LISTEN",
		)
		output, _ := cmd.Output()
		port := uint32(0)
		for i, line := range strings.Split(string(output), "\n") {
			if i == 0 {
				continue // header line
			}

			re := regexp.MustCompile(":(\\d+) \\(LISTEN\\)")
			match := re.FindStringSubmatch(line)
			if match == nil || len(match) < 2 {
				continue
			}
			p, err := strconv.Atoi(match[1])
			if err != nil {
				continue
			}
			port = uint32(p)
			break
		}

		if port != 0 {
			portChan <- port
			return
		}

		select {
		case <-stopChan:
			return
		case <-time.After(100 * time.Millisecond):
			break
		}
	}
}

func handleProcess(process *Process, startCmd []string, envVars []*pb.KeyValuePair) {
	cmd := exec.Command(startCmd[0], startCmd[1:]...)
	cmd.Dir = process.RunPath
	cmd.Env = make([]string, len(os.Environ()))
	copy(cmd.Env, os.Environ())
	for _, envVar := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envVar.Key, envVar.Value))
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	log.Println("Start process " + process.Name)

	portChan := make(chan uint32, 1)
	monitorStopChan := make(chan bool, 1)
	exitChan := make(chan bool, 1)

	process.State = pb.ProcessState_STARTING
	err := cmd.Start()
	if err != nil {
		log.Println("Process "+process.Name+" failed:", err)
		process.State = pb.ProcessState_FINISHED
		// Don't quit as we can still be restarted
	} else {
		go monitorPorts(cmd.Process.Pid, portChan, monitorStopChan)

		go func() {
			cmd.Wait()
			exitChan <- true
		}()
	}

	var proxyStopChan chan bool
	for {
		select {
		case forwardPort := <-portChan:
			// Process starts listening (keep listening for other events)
			if process.State == pb.ProcessState_STARTING {
				if process.Port != forwardPort {
					proxyStopChan, err = startProxy(process.Port, forwardPort)
				}
				process.State = pb.ProcessState_RUNNING
			}
			break
		case <-exitChan:
			// Process exits by itself (keep listening for other events, as we still can be restarted)
			process.State = pb.ProcessState_FINISHED
			break
		case restartReq := <-process.RestartChan:
			// Restart the process group
			monitorStopChan <- true
			if proxyStopChan != nil {
				proxyStopChan <- true
			}
			if process.State != pb.ProcessState_FINISHED {
				// The process is still running, kill it first
				// and wait for the process to die.
				syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
				cmd.Wait()
				process.State = pb.ProcessState_FINISHED
			}
			go handleProcess(process, restartReq.StartCmd, restartReq.EnvVars)
			return
		case <-process.StopChan:
			// Stop the process group
			monitorStopChan <- true
			if proxyStopChan != nil {
				proxyStopChan <- true
			}
			if process.State != pb.ProcessState_FINISHED {
				// The process is still running, kill it first
				// and wait for the process to die.
				syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
				cmd.Wait()
				process.State = pb.ProcessState_FINISHED
			}
			return
		}
	}
}

func (s *ProcessServer) Shutdown() {
	// Kill all processes and make sure they are all stopped before quiting
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, process := range s.process {
		process.StopChan <- true
	}

	// Wait a small amount of time...
	<-time.After(10 * time.Millisecond)

	for {
		allFinished := true
		for _, process := range s.process {
			if process.State != pb.ProcessState_FINISHED {
				allFinished = false
				break
			}
		}

		if allFinished {
			return
		}

		// XXX: have a timeout?
		<-time.After(10 * time.Millisecond)
	}
}

func (s *ProcessServer) ConfigureProcess(c context.Context, req *pb.ConfigureProcessReq) (*pb.ConfigureProcessResp, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	process, found := s.process[req.ProcessName]
	if req.Enabled && !found {
		process = &Process{
			Name:        req.ProcessName,
			RunPath:     req.RunPath,
			Port:        req.ListeningPort,
			State:       pb.ProcessState_STARTING,
			RestartChan: make(chan *RestartRequest, 1),
			StopChan:    make(chan bool, 1),
		}
		go handleProcess(process, req.StartCmd, req.EnvVars)
		s.process[req.ProcessName] = process
	} else if !req.Enabled && found {
		process.StopChan <- true
		delete(s.process, req.ProcessName)
	}

	return &pb.ConfigureProcessResp{}, nil
}

func (s *ProcessServer) RestartProcess(c context.Context, req *pb.RestartProcessReq) (*pb.RestartProcessResp, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	process, found := s.process[req.ProcessName]
	if !found {
		return nil, grpc.Errorf(codes.InvalidArgument, "process is not enabled")
	}

	process.RestartChan <- &RestartRequest{
		StartCmd: req.StartCmd,
		EnvVars:  req.EnvVars,
	}
	return &pb.RestartProcessResp{}, nil
}

func (s *ProcessServer) GetProcessState(c context.Context, req *pb.GetProcessStateReq) (*pb.GetProcessStateResp, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	process, found := s.process[req.ProcessName]
	if !found {
		return &pb.GetProcessStateResp{
			ProcessState: pb.ProcessState_NOT_RUNNING,
		}, nil
	}

	return &pb.GetProcessStateResp{
		ProcessState: process.State,
	}, nil
}
