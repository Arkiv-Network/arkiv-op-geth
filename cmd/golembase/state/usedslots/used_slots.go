package usedslots

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

func UsedSlots() *cli.Command {
	cfg := struct {
		nodeURL string
	}{}
	return &cli.Command{
		Name:  "used-slots",
		Usage: "Number of used slots for golem base",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "node-url",
				Usage:       "The URL of the node to connect to",
				Value:       "http://localhost:8545",
				EnvVars:     []string{"NODE_URL"},
				Destination: &cfg.nodeURL,
			},
		},
		Action: func(c *cli.Context) error {

			ctx, stop := signal.NotifyContext(c.Context, os.Interrupt)
			defer stop()

			rpcClient, err := rpc.Dial(cfg.nodeURL)
			if err != nil {
				return fmt.Errorf("failed to connect to node: %w", err)
			}
			defer rpcClient.Close()

			var res *hexutil.Big

			err = rpcClient.CallContext(ctx, &res, "golembase_getNumberOfUsedSlots")
			if err != nil {
				return fmt.Errorf("failed to get storage at: %w", err)
			}

			fmt.Println(res.ToInt().String())

			return nil
		},
	}
}
