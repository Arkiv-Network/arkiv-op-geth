package balance

import (
	"fmt"
	"math/big"
	"os"
	"os/signal"

	"github.com/dustin/go-humanize"
	"github.com/ethereum/go-ethereum/cmd/golembase/account/pkg/useraccount"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/urfave/cli/v2"
)

func AccountBalance() *cli.Command {
	cfg := struct {
		nodeURL string
	}{}
	return &cli.Command{
		Name:  "balance",
		Usage: "Get the balance of an account",
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
			userAccount, err := useraccount.Load()
			if err != nil {
				return fmt.Errorf("failed to load user account: %w", err)
			}

			ctx, cancel := signal.NotifyContext(c.Context, os.Interrupt)
			defer cancel()

			ethclient, err := ethclient.Dial(cfg.nodeURL)
			if err != nil {
				return fmt.Errorf("failed to dial node: %w", err)
			}

			balance, err := ethclient.BalanceAt(ctx, userAccount.Address, nil)
			if err != nil {
				return fmt.Errorf("failed to get balance: %w", err)
			}

			fmt.Println("Address:", userAccount.Address.Hex())
			fmt.Println("Balance:", humanize.Commaf(EthToFloat(balance)), "ETH")

			return nil
		},
	}
}

func EthToFloat(n *big.Int) float64 {
	f := new(big.Rat).SetFrac(n, big.NewInt(params.Ether))
	res, _ := f.Float64()
	return res
}
