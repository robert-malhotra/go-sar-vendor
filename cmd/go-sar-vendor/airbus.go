package main

// Command-line helpers for Airbus OneAtlas Radar SAR API.
//
// Follows the same structure as capellaCmd / iceyeCmd / umbraCmd already in the
// repository. It supports:
//   - POST /feasibility       – gosar airbus feasibility < body.json
//   - POST /catalogue         – gosar airbus catalogue < body.json
//   - POST /baskets           – gosar airbus basket create
//   - POST /baskets/{id}/addItems – gosar airbus basket add --basket-id ID --item ACQID [...]
//   - POST /baskets/{id}/submit – gosar airbus basket submit --basket-id ID
//   - GET /orders/{id}        – gosar airbus order ORD-123
//
// The command inherits global flags (api-key, token-url, base-url) so the SDK
// can target staging or test endpoints.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/robert-malhotra/go-sar-vendor/pkg/airbus"
	"github.com/urfave/cli/v3"
)

/*──────────────────── helpers ──────────────────────────*/

func abClient(cmd *cli.Command) (*airbus.Client, error) {
	opts := []airbus.Option{}
	if tokenURL := cmd.String("token-url"); tokenURL != "" {
		opts = append(opts, airbus.WithTokenURL(tokenURL))
	}
	if baseURL := cmd.String("base-url"); baseURL != "" {
		opts = append(opts, airbus.WithBaseURL(baseURL))
	}
	return airbus.NewClient(cmd.String("api-key"), opts...)
}

func prettyJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

/*──────────────────── root command ────────────────────*/

func airbusCmd() *cli.Command {
	return &cli.Command{
		Name:  "airbus",
		Usage: "Airbus OneAtlas Radar SAR API",

		Flags: []cli.Flag{
			&cli.StringFlag{Name: "api-key", Required: true, Usage: "OneAtlas API key", Sources: cli.EnvVars("AIRBUS_API_KEY")},
			&cli.StringFlag{Name: "token-url", Value: airbus.DefaultTokenURL, Usage: "Override auth endpoint"},
			&cli.StringFlag{Name: "base-url", Value: airbus.DefaultBaseURL, Usage: "Override API base URL"},
		},

		Commands: []*cli.Command{
			abFeasCmd(),
			abCatCmd(),
			abBasketCmd(),
			abOrderCmd(),
			abPingCmd(),
			abWhoAmICmd(),
		},
	}
}

/*──────────────────── ping ─────────────────────────────*/

func abPingCmd() *cli.Command {
	return &cli.Command{
		Name:  "ping",
		Usage: "Check API availability",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cli, err := abClient(cmd)
			if err != nil {
				return err
			}
			if err := cli.Ping(ctx); err != nil {
				return err
			}
			fmt.Println("OK")
			return nil
		},
	}
}

/*──────────────────── whoami ───────────────────────────*/

func abWhoAmICmd() *cli.Command {
	return &cli.Command{
		Name:  "whoami",
		Usage: "Get current user information",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cli, err := abClient(cmd)
			if err != nil {
				return err
			}
			user, err := cli.WhoAmI(ctx)
			if err != nil {
				return err
			}
			return prettyJSON(user)
		},
	}
}

/*──────────────────── feasibility ─────────────────────*/

func abFeasCmd() *cli.Command {
	return &cli.Command{
		Name:  "feasibility",
		Usage: "Create feasibility/tasking request (reads JSON from stdin)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var body airbus.FeasibilityRequest
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				return fmt.Errorf("failed to parse request JSON: %w", err)
			}
			cli, err := abClient(cmd)
			if err != nil {
				return err
			}
			res, err := cli.SearchFeasibility(ctx, &body)
			if err != nil {
				return err
			}
			return prettyJSON(res)
		},
	}
}

/*──────────────────── catalogue ───────────────────────*/

func abCatCmd() *cli.Command {
	return &cli.Command{
		Name:  "catalogue",
		Usage: "Search archive catalogue (reads JSON from stdin)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var body airbus.CatalogueRequest
			if err := json.NewDecoder(os.Stdin).Decode(&body); err != nil {
				return fmt.Errorf("failed to parse request JSON: %w", err)
			}
			cli, err := abClient(cmd)
			if err != nil {
				return err
			}
			res, err := cli.SearchCatalogue(ctx, &body)
			if err != nil {
				return err
			}
			return prettyJSON(res)
		},
	}
}

/*──────────────────── basket flow ───────────────────────*/

func abBasketCmd() *cli.Command {
	return &cli.Command{
		Name:  "basket",
		Usage: "Basket management commands",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List all baskets",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					baskets, err := cli.ListBaskets(ctx)
					if err != nil {
						return err
					}
					return prettyJSON(baskets)
				},
			},
			{
				Name:  "create",
				Usage: "Create a new basket",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "customer-ref", Usage: "Customer reference"},
					&cli.StringFlag{Name: "purpose", Usage: "Order purpose"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					basket, err := cli.CreateBasket(ctx, &airbus.CreateBasketRequest{
						CustomerReference: cmd.String("customer-ref"),
						Purpose:           airbus.Purpose(cmd.String("purpose")),
					})
					if err != nil {
						return err
					}
					return prettyJSON(basket)
				},
			},
			{
				Name:      "get",
				Usage:     "Get basket details",
				ArgsUsage: "<basketId>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					basketID := cmd.Args().First()
					if basketID == "" {
						return fmt.Errorf("basketId required")
					}
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					basket, err := cli.GetBasket(ctx, basketID)
					if err != nil {
						return err
					}
					return prettyJSON(basket)
				},
			},
			{
				Name:  "add",
				Usage: "Add items to a basket",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "basket-id", Required: true, Usage: "Basket ID"},
					&cli.StringSliceFlag{Name: "acquisition", Usage: "Acquisition ID (repeatable)"},
					&cli.StringSliceFlag{Name: "item", Usage: "Item UUID (repeatable)"},
					&cli.StringFlag{Name: "product-type", Value: "EEC", Usage: "Product type (SSC, MGD, GEC, EEC)"},
					&cli.StringFlag{Name: "resolution", Value: "RE", Usage: "Resolution variant (SE, RE)"},
					&cli.StringFlag{Name: "orbit-type", Value: "science", Usage: "Orbit type (rapid, science, NRT)"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					acquisitions := cmd.StringSlice("acquisition")
					items := cmd.StringSlice("item")
					if len(acquisitions) == 0 && len(items) == 0 {
						return fmt.Errorf("at least one --acquisition or --item required")
					}
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					basket, err := cli.AddItemsToBasket(ctx, cmd.String("basket-id"), &airbus.AddItemsRequest{
						Acquisitions: acquisitions,
						Items:        items,
						OrderOptions: &airbus.OrderOptions{
							ProductType:       airbus.ProductType(cmd.String("product-type")),
							ResolutionVariant: airbus.ResolutionVariant(cmd.String("resolution")),
							OrbitType:         airbus.OrbitType(cmd.String("orbit-type")),
							MapProjection:     airbus.MapProjectionAuto,
						},
					})
					if err != nil {
						return err
					}
					return prettyJSON(basket)
				},
			},
			{
				Name:  "update",
				Usage: "Update basket parameters",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "basket-id", Required: true, Usage: "Basket ID"},
					&cli.StringFlag{Name: "purpose", Usage: "Order purpose"},
					&cli.StringFlag{Name: "customer-ref", Usage: "Customer reference"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					basket, err := cli.UpdateBasket(ctx, cmd.String("basket-id"), &airbus.UpdateBasketRequest{
						Purpose: airbus.Purpose(cmd.String("purpose")),
					})
					if err != nil {
						return err
					}
					return prettyJSON(basket)
				},
			},
			{
				Name:  "submit",
				Usage: "Submit basket as order",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "basket-id", Required: true, Usage: "Basket ID"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					order, err := cli.SubmitBasket(ctx, cmd.String("basket-id"))
					if err != nil {
						return err
					}
					return prettyJSON(order)
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a basket",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "basket-id", Required: true, Usage: "Basket ID"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					if err := cli.DeleteBasket(ctx, cmd.String("basket-id")); err != nil {
						return err
					}
					fmt.Println("Basket deleted")
					return nil
				},
			},
		},
	}
}

/*──────────────────── order ───────────────────────────*/

func abOrderCmd() *cli.Command {
	return &cli.Command{
		Name:  "order",
		Usage: "Order management commands",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List all orders",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					orders, err := cli.ListOrders(ctx)
					if err != nil {
						return err
					}
					return prettyJSON(orders)
				},
			},
			{
				Name:      "get",
				Usage:     "Get order details",
				ArgsUsage: "<orderId>",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					id := cmd.Args().First()
					if id == "" {
						return fmt.Errorf("orderId required")
					}
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					order, err := cli.GetOrder(ctx, id)
					if err != nil {
						return err
					}
					return prettyJSON(order)
				},
			},
			{
				Name:  "cancel",
				Usage: "Cancel order items",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{Name: "item", Required: true, Usage: "Item UUID to cancel (repeatable)"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cli, err := abClient(cmd)
					if err != nil {
						return err
					}
					result, err := cli.CancelOrderItems(ctx, &airbus.CancelItemsRequest{
						Items: cmd.StringSlice("item"),
					})
					if err != nil {
						return err
					}
					return prettyJSON(result)
				},
			},
		},
	}
}
