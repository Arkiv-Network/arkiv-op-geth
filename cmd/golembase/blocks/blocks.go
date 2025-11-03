package blocks

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strconv"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

func Blocks() *cli.Command {
	return &cli.Command{
		Name:  "blocks",
		Usage: "manage blocks",
		Subcommands: []*cli.Command{
			blockList(),
			blockDetails(),
		},
	}
}

func blockList() *cli.Command {
	cfg := struct {
		nodeURL string
	}{}
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "list blocks",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "node-url",
				Usage:       "The URL of the node to connect to",
				Value:       "http://localhost:8545",
				Destination: &cfg.nodeURL,
				EnvVars:     []string{"NODE_URL"},
			},
		},
		Action: func(c *cli.Context) error {

			ctx, stop := signal.NotifyContext(c.Context, os.Interrupt)
			defer stop()

			rpcClient, err := rpc.DialContext(ctx, cfg.nodeURL)
			if err != nil {
				return fmt.Errorf("failed to connect to node: %w", err)
			}
			defer rpcClient.Close()

			client := ethclient.NewClient(rpcClient)
			defer client.Close()

			lastHeader, err := client.HeaderByNumber(ctx, nil)
			if err != nil {
				return fmt.Errorf("failed to get last header: %w", err)
			}

			for block := range lastHeader.Number.Uint64() {
				header, err := client.HeaderByNumber(ctx, big.NewInt(int64(block)))
				if err != nil {
					return fmt.Errorf("failed to get block: %w", err)
				}

				fmt.Printf("Block: %d %x\n", block, header.Hash())
			}

			return nil
		},
	}
}

func blockDetails() *cli.Command {
	cfg := struct {
		nodeURL string
	}{}
	return &cli.Command{
		Name:    "cat",
		Aliases: []string{"details"},
		Usage:   "get block details",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "node-url",
				Usage:       "The URL of the node to connect to",
				Value:       "http://localhost:8545",
				Destination: &cfg.nodeURL,
				EnvVars:     []string{"NODE_URL"},
			},
		},
		Action: func(c *cli.Context) error {

			ctx, stop := signal.NotifyContext(c.Context, os.Interrupt)
			defer stop()

			block := c.Args().First()
			if block == "" {
				return fmt.Errorf("block number is required")
			}

			blockNumber, err := strconv.ParseUint(block, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse block number: %w", err)
			}

			rpcClient, err := rpc.DialContext(ctx, cfg.nodeURL)
			if err != nil {
				return fmt.Errorf("failed to connect to node: %w", err)
			}
			defer rpcClient.Close()

			client := ethclient.NewClient(rpcClient)
			defer client.Close()

			b, err := client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
			if err != nil {
				return fmt.Errorf("failed to get block: %w", err)
			}

			fmt.Println("block", b.NumberU64())
			fmt.Println("  hash", b.Hash())
			fmt.Println("  parent", b.ParentHash())
			fmt.Println("transactions:")

			receipts, err := client.BlockReceipts(ctx, rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(b.NumberU64())))
			if err != nil {
				return fmt.Errorf("failed to get receipts: %w", err)
			}

			enc := json.NewEncoder(os.Stdout)
			for i, _ := range b.Transactions() {
				fmt.Print("    ")
				enc.SetIndent("    ", "  ")
				// enc.Encode(tx)
				enc.Encode(receipts[i])
			}

			return nil
		},
	}
}
