package cat

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

func Cat() *cli.Command {
	cfg := struct {
		NodeURL string
	}{}
	return &cli.Command{
		Name:  "cat",
		Usage: "cat entity",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "node-url",
				Usage:       "The URL of the node to connect to",
				Value:       "http://localhost:8545",
				Destination: &cfg.NodeURL,
				EnvVars:     []string{"NODE_URL"},
			},
		},
		Action: func(c *cli.Context) error {

			ctx, stop := signal.NotifyContext(c.Context, os.Interrupt)
			defer stop()

			key := c.Args().First()
			if key == "" {
				return fmt.Errorf("key is required")
			}
			// Connect to the geth node
			rpcClient, err := rpc.Dial(cfg.NodeURL)
			if err != nil {
				return fmt.Errorf("failed to connect to node: %w", err)
			}
			defer rpcClient.Close()

			var v []byte

			err = rpcClient.CallContext(
				ctx,
				&v,
				"golembase_getStorageValue",
				key,
			)
			if err != nil {
				return fmt.Errorf("failed to get storage value: %w", err)
			}

			fmt.Println("data:", string(v))

			return nil
		},
	}
}

func pointerOf[T any](v T) *T {
	return &v
}
