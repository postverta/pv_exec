package main

import (
	"github.com/postverta/pv_exec/client"
	"github.com/postverta/pv_exec/server"
	"gopkg.in/urfave/cli.v2"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func daemonAction(c *cli.Context) error {
	getAbsRoots := func(roots []string) ([]string, error) {
		absRoots := make([]string, len(roots))
		for i, dir := range roots {
			var err error
			absRoots[i], err = filepath.Abs(dir)
			if err != nil {
				return nil, err
			}
		}
		return absRoots, nil
	}

	if len(c.StringSlice("exec-config-root")) == 0 {
		return cli.Exit("Need at least one exec config root", 1)
	}

	execConfigRoots, err := getAbsRoots(c.StringSlice("exec-config-root"))
	if err != nil {
		return cli.Exit("Failed to read config root, err:"+err.Error(), 1)
	}

	config := &server.Config{
		Host:                    c.String("host"),
		Port:                    c.Uint("p"),
		ExecConfigRoots:         execConfigRoots,
		AccountName:             c.String("account-name"),
		AccountKey:              c.String("account-key"),
		Container:               c.String("container"),
		SourceWorktreeId:        c.String("source-worktree"),
		WorktreeId:              c.String("worktree"),
		MountPoint:              c.String("mount-point"),
		AutoSaveIntervalSeconds: c.Uint("autosave-interval"),
	}

	err = server.Start(config)
	if err != nil {
		return cli.Exit("Failed to start server, err:"+err.Error(), 1)
	} else {
		return nil
	}
}

func execAction(c *cli.Context) error {
	if c.NArg() == 0 {
		return cli.Exit("Need to specify the task name", 1)
	}
	taskName := c.Args().First()
	keys := make([]string, c.NArg()-1)
	values := make([]string, c.NArg()-1)
	for i, keyValuePair := range c.Args().Tail() {
		parts := strings.Split(keyValuePair, "=")
		if len(parts) != 2 {
			return cli.Exit("Not following a 'key=value' format: "+keyValuePair, 1)
		}
		keys[i] = parts[0]
		values[i] = parts[1]
	}

	var input io.Reader
	if c.Bool("i") {
		input = os.Stdin
	} else {
		input = nil
	}

	err := client.Exec(c.String("host"), c.Uint("p"), taskName, keys, values, input, c.Bool("w"))
	if err != nil {
		return cli.Exit("Failed to send task, err:"+err.Error(), 1)
	} else {
		return nil
	}
}

func processAction(c *cli.Context) error {
	if c.NArg() < 2 {
		return cli.Exit("Need to specify the process name and the command", 1)
	}
	processName := c.Args().Get(0)
	command := c.Args().Get(1)
	if command != "enable" && command != "disable" && command != "restart" && command != "state" {
		return cli.Exit("Unknown command", 1)
	}

	if (command == "enable" || command == "restart") && c.String("start-cmd") == "" {
		return cli.Exit("Must specify start command", 1)
	}

	if command == "enable" && c.String("run-path") == "" {
		return cli.Exit("Must specify run path", 1)
	}

	err := client.Process(c.String("host"), c.Uint("p"), processName, command, c.String("start-cmd"), c.String("run-path"))
	if err != nil {
		return cli.Exit("Failed to manage process, err:"+err.Error(), 1)
	} else {
		return nil
	}
}

func main() {
	app := &cli.App{
		Name: "pv_exec",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "p",
				Usage: "Port to listen/connect to",
				Value: 50000,
			},
			&cli.StringFlag{
				Name:  "host",
				Usage: "Host to listen/connect to",
				Value: "127.0.0.1",
			},
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "daemon",
				Usage: "pv_exec [-p PORT] [-host HOST] daemon [-exec-config-root EXEC_CONFIG_ROOT0]... [-account-name AZURE_ACCOUNT_NAME] [-account-key AZURE_ACCOUNT_KEY] [-container AZURE_BLOB_CONTAINER] [-source-image SOURCE_IMAGE_ID] [-image IMAGE_ID] [-mount-point MOUNT_POINT] [-autosave-interval AUTOSAVE_INTERVAL]",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "exec-config-root",
						Usage: "Exec service config roots",
					},
					&cli.StringFlag{
						Name:  "account-name",
						Usage: "Azure storage account name",
					},
					&cli.StringFlag{
						Name:  "account-key",
						Usage: "Azure storage account key",
					},
					&cli.StringFlag{
						Name:  "container",
						Usage: "Azure blob container name",
					},
					&cli.StringFlag{
						Name:  "source-worktree",
						Usage: "Source worktree ID (optional)",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "worktree",
						Usage: "Worktree ID",
					},
					&cli.StringFlag{
						Name:  "mount-point",
						Usage: "Mount point",
					},
					&cli.UintFlag{
						Name:  "autosave-interval",
						Usage: "Autosave interval in seconds",
					},
				},
				Action: daemonAction,
			},
			&cli.Command{
				Name:  "exec",
				Usage: "pv_exec [-p PORT] [-host HOST] exec [-w] [-i] TASK_NAME [KEY0=VALUE0 ...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "w",
						Usage: "Wait until the task is completed and print output to stdout",
					},
					&cli.BoolFlag{
						Name:  "i",
						Usage: "Pass stdin as the stdin of the task (WARNING: the CLI might block)",
					},
				},
				Action: execAction,
			},
			&cli.Command{
				Name:  "process",
				Usage: "pv_exec [-p PORT] [-host HOST] process [-start-cmd Start command] [-run-path Running path] PROCESS_NAME enable/disable/restart/state",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "start-cmd",
						Usage: "Command to start the process",
					},
					&cli.StringFlag{
						Name:  "run-path",
						Usage: "Running path of the process",
					},
				},
				Action: processAction,
			},
		},
	}

	app.Run(os.Args)
}
