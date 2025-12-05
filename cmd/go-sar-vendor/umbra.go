package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/robert-malhotra/go-sar-vendor/pkg/umbra"
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
				Value: "https://api.canopy.umbra.space/",
				Usage: "Override Umbra API base URL",
			},
			&cli.StringFlag{
				Name:     "api-key",
				Required: true,
				Usage:    "Umbra bearer token",
				Sources:  cli.EnvVars("UMBRA_API_KEY"),
			},
		},

		Commands: []*cli.Command{
			feasibilityCmd(),
			taskCmd(),
			collectsCmd(),
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

func umbraClientFromCmd(cmd *cli.Command) *umbra.Client {
	return umbra.NewClient(
		cmd.String("api-key"),
		umbra.WithBaseURL(cmd.String("vendor-base-url")),
	)
}

/*──────────────── feasibility actions ───────────────────────────────────────*/

func umbraCreateFeasAction(ctx context.Context, cmd *cli.Command) error {
	var req umbra.CreateFeasibilityRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	cli := umbraClientFromCmd(cmd)
	resp, err := cli.CreateFeasibility(ctx, &req)
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

	cli := umbraClientFromCmd(cmd)
	resp, err := cli.GetFeasibility(ctx, id)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

/*──────────────── task actions ──────────────────────────────────────────────*/

func umbraCreateTaskAction(ctx context.Context, cmd *cli.Command) error {
	var req umbra.CreateTaskRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	cli := umbraClientFromCmd(cmd)
	resp, err := cli.CreateTask(ctx, &req)
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

	cli := umbraClientFromCmd(cmd)
	resp, err := cli.GetTask(ctx, id)
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

	cli := umbraClientFromCmd(cmd)
	resp, err := cli.CancelTask(ctx, id)
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

	cli := umbraClientFromCmd(cmd)
	seq := cli.SearchTasks(ctx, req)
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

	cli := umbraClientFromCmd(cmd)
	resp, err := cli.GetCollect(ctx, id)
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

	cli := umbraClientFromCmd(cmd)
	seq := cli.SearchCollects(ctx, req)
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
