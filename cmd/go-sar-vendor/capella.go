package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/robert.malhotra/umbra-client/pkg/capella"
	"github.com/urfave/cli/v3"
)

// -----------------------------------------------------------------------------
// top-level command ------------------------------------------------------------
// -----------------------------------------------------------------------------
func capellaCmd() *cli.Command {
	root := &cli.Command{
		Name:  "capella",
		Usage: "Command-line helper for Capella Space Tasking & Access API",

		// global flags
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "api-key",
				Usage:    "Capella API key (or set CAPELLA_API_KEY env var)",
				Required: true,
				Sources:  cli.EnvVars("CAPELLA_API_KEY"),
			},
			&cli.StringFlag{
				Name:  "base-url",
				Value: "https://api.capellaspace.com/",
				Usage: "Override API base URL (useful for staging/dev)",
			},
		},

		// sub-commands
		Commands: []*cli.Command{
			accessCmd(),
			tasksCmd(),
		},
	}
	return root
}

// -----------------------------------------------------------------------------
// helpers ---------------------------------------------------------------------
// -----------------------------------------------------------------------------

func clientFromCmd(cmd *cli.Command) *capella.Client {
	return capella.NewClient(capella.WithBaseURL(cmd.String("base-url")))
}

// -----------------------------------------------------------------------------
// Access-request sub-tree ------------------------------------------------------
// -----------------------------------------------------------------------------

func accessCmd() *cli.Command {
	return &cli.Command{
		Name:  "access",
		Usage: "Create & inspect access-requests (Feasibility Studies)",
		Commands: []*cli.Command{
			{
				Name:   "create",
				Usage:  "Create an access request (reads JSON from stdin, prints result)",
				Action: accessCreateAction,
			},
			{
				Name:      "get",
				Usage:     "Fetch an access request by ID",
				ArgsUsage: "<accessRequestId>",
				Action:    accessGetAction,
			},
		},
	}
}

func accessCreateAction(ctx context.Context, cmd *cli.Command) error {
	var req capella.AccessRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	cli := clientFromCmd(cmd)
	resp, err := cli.CreateAccessRequest(cmd.String("api-key"), req)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func accessGetAction(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().Get(0)
	if id == "" {
		return fmt.Errorf("accessRequestId required")
	}
	cli := clientFromCmd(cmd)
	resp, err := cli.GetAccessRequest(cmd.String("api-key"), id)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

// -----------------------------------------------------------------------------
// Tasking-request sub-tree ----------------------------------------------------
// -----------------------------------------------------------------------------

func tasksCmd() *cli.Command {
	return &cli.Command{
		Name:  "tasks",
		Usage: "Create, approve, list & search tasking-requests",
		Commands: []*cli.Command{
			{
				Name:   "create",
				Usage:  "Create a tasking request (reads JSON from stdin)",
				Action: tasksCreateAction,
			},
			{
				Name:      "get",
				Usage:     "Fetch a tasking request by ID",
				ArgsUsage: "<taskingRequestId>",
				Action:    tasksGetAction,
			},
			{
				Name:      "approve",
				Usage:     "Approve a tasking request (cost review)",
				ArgsUsage: "<taskingRequestId>",
				Action:    tasksApproveAction,
			},
			{
				Name:  "list",
				Usage: "Stream request list as newline-delimited JSON (paged API)",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "customer-id"},
					&cli.StringFlag{Name: "organization-id"},
					&cli.IntFlag{Name: "limit", Value: 25},
				},
				Action: tasksListAction,
			},
			{ // NEW: /tasks/search ------------------------------------------
				Name:   "search",
				Usage:  "Advanced search (reads JSON from stdin, prints paged result)",
				Action: tasksSearchAction,
			},
		},
	}
}

func tasksCreateAction(ctx context.Context, cmd *cli.Command) error {
	var req capella.TaskingRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return err
	}
	cli := clientFromCmd(cmd)
	resp, err := cli.CreateTaskingRequest(cmd.String("api-key"), req)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func tasksGetAction(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().Get(0)
	if id == "" {
		return fmt.Errorf("taskingRequestId required")
	}
	cli := clientFromCmd(cmd)
	resp, err := cli.GetTaskingRequest(cmd.String("api-key"), id)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func tasksApproveAction(ctx context.Context, cmd *cli.Command) error {
	id := cmd.Args().Get(0)
	if id == "" {
		return fmt.Errorf("taskingRequestId required")
	}
	cli := clientFromCmd(cmd)
	resp, err := cli.ApproveTaskingRequest(cmd.String("api-key"), id)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func tasksListAction(ctx context.Context, cmd *cli.Command) error {
	params := capella.PagedTasksParams{
		CustomerID:     cmd.String("customer-id"),
		OrganizationID: cmd.String("organization-id"),
		Limit:          cmd.Int("limit"),
	}
	cli := clientFromCmd(cmd)
	for t, err := range cli.ListTasksPaged(cmd.String("api-key"), params) {
		if err != nil {
			return err
		}
		if err := printJSON(t); err != nil {
			return err
		}
	}
	return nil
}

// -------- NEW: tasksSearchAction ---------------------------------------------

func tasksSearchAction(ctx context.Context, cmd *cli.Command) error {
	var req capella.TaskSearchRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	cli := clientFromCmd(cmd)
	resp, err := cli.SearchTasks(cmd.String("api-key"), req)
	if err != nil {
		return err
	}
	return printJSON(resp)
}
