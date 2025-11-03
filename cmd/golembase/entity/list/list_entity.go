package list

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

func List() *cli.Command {
	cfg := struct {
		nodeURL string
	}{}

	return &cli.Command{
		Name:  "list",
		Usage: "List all entity keys",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "node-url",
				Usage:       "HTTP-RPC server endpoint",
				Value:       "http://localhost:8545",
				EnvVars:     []string{"NODE_URL"},
				Destination: &cfg.nodeURL,
			},
		},
		Action: func(c *cli.Context) error {
			ctx, cancel := signal.NotifyContext(c.Context, os.Interrupt)
			defer cancel()

			client, err := rpc.DialContext(ctx, cfg.nodeURL)
			if err != nil {
				return fmt.Errorf("failed to connect to the Ethereum client: %w", err)
			}

			var entityKeys []string
			err = client.CallContext(ctx, &entityKeys, "golembase_getAllEntityKeys")
			if err != nil {
				return fmt.Errorf("failed to get entity keys: %w", err)
			}

			if len(entityKeys) == 0 {
				fmt.Println("No entities found")
				return nil
			}

			fmt.Println("Entity keys:")
			for _, key := range entityKeys {
				fmt.Println(key)
			}

			return nil
		},
	}
}
