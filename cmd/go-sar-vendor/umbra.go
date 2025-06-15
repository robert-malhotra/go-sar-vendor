package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/robert.malhotra/umbra-client/pkg/umbra"
	"github.com/urfave/cli/v3"
)

/*──────────────── root "umbra" command ──────────────────────────────────────*/

func umbraCmd() *cli.Command {
	return &cli.Command{
		Name:  "umbra",
		Usage: "Umbra Space tasking API helpers",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "vendor-base-url",
				Value: "https://api.umbra.space/",
				Usage: "Override Umbra API base URL",
			},
			&cli.StringFlag{
				Name:     "api-key",
				Required: true,
				Usage:    "Umbra bearer token",
			},
		},

		Commands: []*cli.Command{
			feasibilityCmd(),
			taskCmd(),
			collectsCmd(), // NEW
		},
	}
}

/*──────────────── feasibility commands ──────────────────────────────────────*/

func feasibilityCmd() *cli.Command {
	return &cli.Command{
		Name:  "feasibility",
		Usage: "Feasibility operations (create / get)",

		Commands: []*cli.Command{
			{
				Name:   "create",
				Usage:  "Create a feasibility request (reads JSON from stdin)",
				Action: umbraCreateFeasAction,
			},
			{
				Name:      "get",
				Usage:     "Fetch a feasibility by ID",
				ArgsUsage: "<feasibilityId>",
				Action:    umbraGetFeasAction,
			},
		},
	}
}

/*──────────────── task commands ─────────────────────────────────────────────*/

func taskCmd() *cli.Command {
	return &cli.Command{
		Name:  "task",
		Usage: "Task operations (create / get / cancel / search)",

		Commands: []*cli.Command{
			{
				Name:   "create",
				Usage:  "Create a task (reads JSON from stdin)",
				Action: umbraCreateTaskAction,
			},
			{
				Name:      "get",
				Usage:     "Fetch a task by ID",
				ArgsUsage: "<taskId>",
				Action:    umbraGetTaskAction,
			},
			{
				Name:      "cancel",
				Usage:     "Cancel a task by ID",
				ArgsUsage: "<taskId>",
				Action:    umbraCancelTaskAction,
			},
			{
				Name:   "search",
				Usage:  "Search tasks (reads JSON TaskSearchRequest from stdin)",
				Action: umbraSearchTaskAction,
			},
		},
	}
}

/*──────────────── collects commands ─────────────────────────────────────────*/

func collectsCmd() *cli.Command {
	return &cli.Command{
		Name:  "collects",
		Usage: "Collect operations (get / search)",

		Commands: []*cli.Command{
			{
				Name:      "get",
				Usage:     "Fetch a collect by ID",
				ArgsUsage: "<collectId>",
				Action:    umbraGetCollectAction,
			},
			{
				Name:   "search",
				Usage:  "Search collects (reads JSON CollectSearchRequest from stdin)",
				Action: umbraSearchCollectAction,
			},
		},
	}
}

/*──────────────── helpers ───────────────────────────────────────────────────*/

func umbraClientFromCmd(cmd *cli.Command) (*umbra.Client, error) {
	return umbra.NewClient(cmd.String("vendor-base-url"))
}

/*──────────────── feasibility actions ───────────────────────────────────────*/

func umbraCreateFeasAction(ctx context.Context, cmd *cli.Command) error {
	var req umbra.TaskingRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	cli, err := umbraClientFromCmd(cmd)
	if err != nil {
		return err
	}

	resp, err := cli.CreateFeasibility(cmd.String("api-key"), &req)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func umbraGetFeasAction(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().Get(0)
	if id == "" {
		return fmt.Errorf("feasibilityId required")
	}

	cli, err := umbraClientFromCmd(cmd)
	if err != nil {
		return err
	}

	resp, err := cli.GetFeasibility(cmd.String("api-key"), id)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

/*──────────────── task actions ──────────────────────────────────────────────*/

func umbraCreateTaskAction(ctx context.Context, cmd *cli.Command) error {
	var req umbra.TaskRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	cli, err := umbraClientFromCmd(cmd)
	if err != nil {
		return err
	}

	resp, err := cli.CreateTask(cmd.String("api-key"), &req)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func umbraGetTaskAction(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().Get(0)
	if id == "" {
		return fmt.Errorf("taskId required")
	}

	cli, err := umbraClientFromCmd(cmd)
	if err != nil {
		return err
	}

	resp, err := cli.GetTask(cmd.String("api-key"), id)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func umbraCancelTaskAction(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().Get(0)
	if id == "" {
		return fmt.Errorf("taskId required")
	}

	cli, err := umbraClientFromCmd(cmd)
	if err != nil {
		return err
	}

	resp, err := cli.CancelTask(cmd.String("api-key"), id)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func umbraSearchTaskAction(ctx context.Context, cmd *cli.Command) error {
	var req umbra.TaskSearchRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	cli, err := umbraClientFromCmd(cmd)
	if err != nil {
		return err
	}

	seq := cli.SearchTasks(cmd.String("api-key"), req)
	for task, err := range seq {
		if err != nil {
			return err
		}
		if err := printJSON(task); err != nil {
			return err
		}
	}
	return nil
}

/*──────────────── collect actions ───────────────────────────────────────────*/

func umbraGetCollectAction(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().Get(0)
	if id == "" {
		return fmt.Errorf("collectId required")
	}

	cli, err := umbraClientFromCmd(cmd)
	if err != nil {
		return err
	}

	resp, err := cli.GetCollect(cmd.String("api-key"), id)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func umbraSearchCollectAction(ctx context.Context, cmd *cli.Command) error {
	var req umbra.CollectSearchRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	cli, err := umbraClientFromCmd(cmd)
	if err != nil {
		return err
	}

	seq := cli.SearchCollects(cmd.String("api-key"), req)
	for col, err := range seq {
		if err != nil {
			return err
		}
		if err := printJSON(col); err != nil {
			return err
		}
	}
	return nil
}
