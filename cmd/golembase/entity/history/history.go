package history

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/golem-base/storagetx"
	"github.com/holiman/uint256"
	"github.com/urfave/cli/v2"
)

func History() *cli.Command {

	cfg := struct {
		nodeURL string
		key     string
	}{}
	return &cli.Command{
		Name:  "history",
		Usage: "Get the history of a given entity",
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
			ctx, cancel := signal.NotifyContext(c.Context, os.Interrupt)
			defer cancel()

			if c.Args().Len() != 1 {
				return fmt.Errorf("entity key is required")
			}

			entityKey := common.HexToHash(c.Args().Get(0))

			// Connect to the geth node
			client, err := ethclient.DialContext(ctx, cfg.nodeURL)
			if err != nil {
				return fmt.Errorf("failed to connect to node: %w", err)
			}
			defer client.Close()

			logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
				Topics: [][]common.Hash{
					{
						storagetx.GolemBaseStorageEntityDeleted,
						storagetx.GolemBaseStorageEntityCreated,
						storagetx.GolemBaseStorageEntityUpdated,
						storagetx.GolemBaseStorageEntityBTLExtended,
					},
					{
						entityKey,
					},
				},
			})

			for _, log := range logs {
				switch log.Topics[0] {
				case storagetx.GolemBaseStorageEntityDeleted:
					fmt.Println("Deleted", log.BlockNumber, log.TxHash)
				case storagetx.GolemBaseStorageEntityCreated:
					expiresAtBlock := new(uint256.Int).SetBytes(log.Data)
					fmt.Println("Created", log.BlockNumber, log.TxHash, "expires at block", expiresAtBlock.Uint64())
				case storagetx.GolemBaseStorageEntityUpdated:
					expiresAtBlock := new(uint256.Int).SetBytes(log.Data)
					fmt.Println("Updated", log.BlockNumber, log.TxHash, "expires at block", expiresAtBlock.Uint64())
				case storagetx.GolemBaseStorageEntityBTLExtended:
					expiresAtBlock := new(uint256.Int).SetBytes(log.Data)
					fmt.Println("BTLExtended", log.BlockNumber, log.TxHash, "expires at block", expiresAtBlock.Uint64())
				}
			}

			if err != nil {
				return fmt.Errorf("failed to filter logs: %w", err)
			}

			return nil

		},
	}

}
