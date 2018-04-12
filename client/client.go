package client

import (
	"bytes"
	"context"
	"fmt"
	pbe "github.com/postverta/pv_exec/proto/exec"
	pbp "github.com/postverta/pv_exec/proto/process"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func Exec(host string, port uint, taskName string, keys []string, values []string, input io.Reader, waitForCompletion bool) error {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := pbe.NewExecServiceClient(conn)

	keyValues := make([]*pbe.ExecReq_KeyValuePair, len(keys))
	for i, key := range keys {
		value := values[i]
		keyValues[i] = &pbe.ExecReq_KeyValuePair{
			Key:   key,
			Value: value,
		}
	}

	data := []byte{}
	if input != nil {
		data, err = ioutil.ReadAll(input)
		if err != nil {
			return err
		}
	}

	req := &pbe.ExecReq{
		TaskName:          taskName,
		KeyValues:         keyValues,
		Data:              data,
		WaitForCompletion: waitForCompletion,
	}

	resp, err := client.Exec(context.Background(), req)
	if err != nil {
		return err
	}

	if resp.Completed {
		buf := bytes.NewBuffer(resp.Data)
		_, err = io.Copy(os.Stdout, buf)
		if err != nil {
			return err
		}
	}

	return nil
}

func Process(host string, port uint, processName string, command string, startCmd string, runPath string) error {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := pbp.NewProcessServiceClient(conn)
	if command == "enable" {
		req := &pbp.ConfigureProcessReq{
			ProcessName:   processName,
			StartCmd:      []string{startCmd},
			RunPath:       runPath,
			Enabled:       true,
			ListeningPort: 9090,
		}

		_, err := client.ConfigureProcess(context.Background(), req)
		if err != nil {
			return err
		}

		log.Println("Enabled process " + processName + " to listen on port 9090")
	} else if command == "disable" {
		req := &pbp.ConfigureProcessReq{
			ProcessName: processName,
			Enabled:     false,
		}

		_, err := client.ConfigureProcess(context.Background(), req)
		if err != nil {
			return err
		}

		log.Println("Disabled process " + processName)
	} else if command == "restart" {
		req := &pbp.RestartProcessReq{
			ProcessName: processName,
			StartCmd:    []string{startCmd},
		}

		_, err := client.RestartProcess(context.Background(), req)
		if err != nil {
			return err
		}

		log.Println("Restarted process " + processName)
	} else if command == "state" {
		req := &pbp.GetProcessStateReq{
			ProcessName: processName,
		}

		resp, err := client.GetProcessState(context.Background(), req)
		if err != nil {
			return err
		}

		log.Println("Process state: " + resp.ProcessState.String())
	}
	return nil
}
