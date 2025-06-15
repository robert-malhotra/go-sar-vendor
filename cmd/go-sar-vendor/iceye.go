// cmd/iceye.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/robert.malhotra/go-sar-vendor/pkg/iceye"
	"github.com/urfave/cli/v3"
)

func iceyeCmd() *cli.Command {
	return &cli.Command{
		Name:  "iceye",
		Usage: "ICEYE Tasking v2 helpers",

		Flags: []cli.Flag{
			&cli.StringFlag{Name: "base-url", Value: "https://api.iceye.com/api"},
			&cli.StringFlag{Name: "token-url", Value: "https://auth.iceye.com/oauth2/token"},
			&cli.StringFlag{Name: "client-id", Required: true, Sources: cli.EnvVars("ICEYE_CLIENT_ID")},
			&cli.StringFlag{Name: "client-secret", Required: true, Sources: cli.EnvVars("ICEYE_CLIENT_SECRET")},
		},

		Commands: []*cli.Command{
			iceyeContractsCmd(),
			iceyeTasksCmd(),
		},
	}
}

/* ---------- helper ---------- */

func iceyeClient(cmd *cli.Command) *iceye.Client {
	return iceye.NewClient(iceye.Config{
		BaseURL:      cmd.String("base-url"),
		TokenURL:     cmd.String("token-url"),
		ClientID:     cmd.String("client-id"),
		ClientSecret: cmd.String("client-secret"),
	})
}

func print(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

/* ---------- contracts ---------- */

func iceyeContractsCmd() *cli.Command {
	return &cli.Command{
		Name:  "contracts",
		Usage: "List contracts (paged iterator)",
		Flags: []cli.Flag{&cli.IntFlag{Name: "limit", Value: 50}},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cli := iceyeClient(cmd)
			limit := cmd.Int("limit")
			seq := cli.ListContracts(ctx, limit)
			for page, err := range seq {
				if err != nil {
					return err
				}
				for _, c := range page.Data {
					if err := print(c); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
}

/* ---------- tasks ---------- */

func iceyeTasksCmd() *cli.Command {
	return &cli.Command{
		Name:  "tasks",
		Usage: "Create / get / cancel / list tasks + scene, price, products",
		Commands: []*cli.Command{
			{
				Name:   "create",
				Usage:  "Create a task (reads JSON from stdin, prints Task)",
				Action: iceyeCreateTask,
			},
			{
				Name:      "get",
				Usage:     "Fetch a task by ID",
				ArgsUsage: "<taskId>",
				Action:    iceyeGetTask,
			},
			{
				Name:      "cancel",
				Usage:     "Cancel a task",
				ArgsUsage: "<taskId>",
				Action:    iceyeCancelTask,
			},
			{
				Name:   "list",
				Usage:  "List tasks paged",
				Flags:  []cli.Flag{&cli.IntFlag{Name: "limit", Value: 50}},
				Action: iceyeListTasks,
			},
			{
				Name:      "scene",
				Usage:     "Get delivered scene metadata",
				ArgsUsage: "<taskId>",
				Action:    iceyeGetScene,
			},
			{
				Name:      "products",
				Usage:     "List delivered products",
				ArgsUsage: "<taskId>",
				Flags:     []cli.Flag{&cli.IntFlag{Name: "limit", Value: 100}},
				Action:    iceyeGetProducts,
			},
			{
				Name:   "price",
				Usage:  "Quote a task price (reads JSON CreateTaskRequest from stdin)",
				Action: iceyeGetPrice,
			},
		},
	}
}

/* ---- task actions ---- */

func iceyeCreateTask(ctx context.Context, cmd *cli.Command) error {
	var req iceye.CreateTaskRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return err
	}
	cli := iceyeClient(cmd)
	resp, err := cli.CreateTask(ctx, req)
	if err != nil {
		return err
	}
	return print(resp)
}

func iceyeGetTask(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().First()
	if id == "" {
		return fmt.Errorf("taskId required")
	}
	cli := iceyeClient(cmd)
	resp, err := cli.GetTask(ctx, id)
	if err != nil {
		return err
	}
	return print(resp)
}

func iceyeCancelTask(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().First()
	if id == "" {
		return fmt.Errorf("taskId required")
	}
	cli := iceyeClient(cmd)
	return cli.CancelTask(ctx, id)
}

func iceyeListTasks(ctx context.Context, cmd *cli.Command) error {
	cli := iceyeClient(cmd)
	limit := cmd.Int("limit")
	seq := cli.ListTasks(ctx, limit, nil)
	for page, err := range seq {
		if err != nil {
			return err
		}
		for _, t := range page {
			if err := print(t); err != nil {
				return err
			}
		}
	}
	return nil
}

func iceyeGetScene(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().First()
	if id == "" {
		return fmt.Errorf("taskId required")
	}
	cli := iceyeClient(cmd)
	resp, err := cli.GetTaskScene(ctx, id)
	if err != nil {
		return err
	}
	return print(resp)
}

func iceyeGetProducts(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().First()
	if id == "" {
		return fmt.Errorf("taskId required")
	}
	cli := iceyeClient(cmd)
	limit := cmd.Int("limit")
	seq := cli.GetTaskProducts(ctx, id, limit)
	for page, err := range seq {
		if err != nil {
			return err
		}
		for _, p := range page {
			if err := print(p); err != nil {
				return err
			}
		}
	}
	return nil
}

func iceyeGetPrice(ctx context.Context, cmd *cli.Command) error {
	var req iceye.CreateTaskRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return err
	}
	cli := iceyeClient(cmd)
	resp, err := cli.GetTaskPrice(ctx, req)
	if err != nil {
		return err
	}
	return print(resp)
}
