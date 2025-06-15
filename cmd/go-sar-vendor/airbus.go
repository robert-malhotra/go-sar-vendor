package main

// Command‑line helpers for Airbus OneAtlas Radar SAR API.
//
// Follows the same structure as capellaCmd / iceyeCmd / umbraCmd already in the
// repository.  It supports:
//   * POST /feasibility      – gosar airbus feasibility < body.json
//   * POST /catalogue        – gosar airbus catalogue < body.json
//   * add + submit cart      – gosar airbus cart --item ID [...]
//   * GET  /orders/{id}      – gosar airbus order ORD-123
//
// The command inherits global flags (api‑key, token‑url, base‑url) so the SDK
// can target staging or test endpoints.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/robert.malhotra/umbra-client/pkg/airbus"
	"github.com/urfave/cli/v3"
)

/*──────────────────── helpers ──────────────────────────*/

func abClient(cmd *cli.Command) *airbus.Client {
	return airbus.New(
		cmd.String("api-key"),
		nil, // default http.Client
		airbus.WithTokenURL(cmd.String("token-url")),
		airbus.WithBaseURL(cmd.String("base-url")),
	)
}

func pretty(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

/*──────────────────── root command ────────────────────*/

func airbusCmd() *cli.Command {
	return &cli.Command{
		Name:  "airbus",
		Usage: "Airbus OneAtlas Radar SAR helpers",

		Flags: []cli.Flag{
			&cli.StringFlag{Name: "api-key", Required: true, Usage: "OneAtlas API key", Sources: cli.EnvVars("AIRBUS_API_KEY")},
			&cli.StringFlag{Name: "token-url", Value: airbus.DefaultTokenURL, Usage: "Override auth endpoint"},
			&cli.StringFlag{Name: "base-url", Value: airbus.DefaultBaseURL, Usage: "Override API base URL"},
		},

		Commands: []*cli.Command{
			abFeasCmd(),
			abCatCmd(),
			abCartCmd(),
			abOrderCmd(),
		},
	}
}

/*──────────────────── feasibility ─────────────────────*/

func abFeasCmd() *cli.Command {
	return &cli.Command{
		Name:  "feasibility",
		Usage: "Create feasibility request (reads JSON from stdin)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var body airbus.FeasibilityRequest
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				return err
			}
			res, err := abClient(cmd).SearchFeasibility(ctx, body)
			if err != nil {
				return err
			}
			return pretty(res)
		},
	}
}

/*──────────────────── catalogue ───────────────────────*/

func abCatCmd() *cli.Command {
	return &cli.Command{
		Name:  "catalogue",
		Usage: "Archive search (reads JSON from stdin)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var body airbus.CatalogueRequest
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				return err
			}
			res, err := abClient(cmd).SearchCatalogue(ctx, body)
			if err != nil {
				return err
			}
			return pretty(res)
		},
	}
}

/*──────────────────── cart flow ───────────────────────*/

func abCartCmd() *cli.Command {
	return &cli.Command{
		Name:  "cart",
		Usage: "Add items then submit order",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{Name: "item", Usage: "itemId from feasibility / catalogue (repeatable)"},
			&cli.StringFlag{Name: "purpose", Value: "Research"},
			&cli.StringFlag{Name: "product-type", Value: string(airbus.PTSSC)},
			&cli.StringFlag{Name: "resolution", Value: string(airbus.ResRadiometric)},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			items := cmd.StringSlice("item")
			if len(items) == 0 {
				return fmt.Errorf("at least one --item required")
			}
			cli := abClient(cmd)

			// add items
			if err := cli.AddItems(ctx, airbus.AddItemsRequest{
				Items: items,
				OrderOptions: airbus.OrderOptions{
					ProductType:       airbus.ProductType(cmd.String("product-type")),
					ResolutionVariant: airbus.ResolutionVariant(cmd.String("resolution")),
					OrbitType:         airbus.OrbitRapid,
					MapProjection:     airbus.MapAuto,
				},
			}); err != nil {
				return err
			}
			// patch purpose
			if err := cli.PatchShopcart(ctx, airbus.ShopcartPatch{Purpose: airbus.Purpose(cmd.String("purpose"))}); err != nil {
				return err
			}
			// submit
			id, err := cli.SubmitShopcart(ctx)
			if err != nil {
				return err
			}
			fmt.Println(id)
			return nil
		},
	}
}

/*──────────────────── order ───────────────────────────*/

func abOrderCmd() *cli.Command {
	return &cli.Command{
		Name:      "order",
		Usage:     "Fetch order status",
		ArgsUsage: "<orderId>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			id := cmd.Args().First()
			if id == "" {
				return fmt.Errorf("orderId required")
			}
			res, err := abClient(cmd).Order(ctx, id)
			if err != nil {
				return err
			}
			return pretty(res)
		},
	}
}
