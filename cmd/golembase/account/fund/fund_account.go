package fund

import (
	"fmt"
	"math/big"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/cmd/golembase/account/pkg/useraccount"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/params"
	"github.com/urfave/cli/v2"
)

func FundAccount() *cli.Command {
	cfg := struct {
		nodeURL string
		value   int64
	}{}
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "node-url",
				Usage:       "The URL of the node to connect to",
				Value:       "http://localhost:8545",
				EnvVars:     []string{"NODE_URL"},
				Destination: &cfg.nodeURL,
			},
			&cli.Int64Flag{
				Name:        "value",
				Usage:       "The amount of ETH to fund the account with",
				Value:       100,
				EnvVars:     []string{"VALUE"},
				Destination: &cfg.value,
			},
		},
		Name:  "fund",
		Usage: "Fund an account",
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

			rpcClient := ethclient.Client()

			// Get the available accounts
			var accounts []common.Address
			err = rpcClient.CallContext(ctx, &accounts, "eth_accounts")
			if err != nil {
				return fmt.Errorf("failed to get accounts: %w", err)
			}
			if len(accounts) == 0 {
				return fmt.Errorf("no accounts found")
			}

			from := accounts[0]

			nonce, err := ethclient.PendingNonceAt(ctx, from)
			if err != nil {
				return fmt.Errorf("failed to get nonce: %w", err)
			}

			chainID, err := ethclient.ChainID(ctx)
			if err != nil {
				return fmt.Errorf("failed to get chain ID: %w", err)
			}

			tx := ethapi.TransactionArgs{
				From:                 pointerOf(from),
				ChainID:              (*hexutil.Big)(chainID),
				Nonce:                (*hexutil.Uint64)(&nonce),
				MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(1e9)), // 1 Gwei
				MaxFeePerGas:         (*hexutil.Big)(big.NewInt(5e9)), // 5 Gwei
				Gas:                  (*hexutil.Uint64)(pointerOf(uint64(2_800_000))),
				To:                   pointerOf(userAccount.Address), //
				Value:                (*hexutil.Big)(EthToWei(cfg.value)),
			}

			var txHash common.Hash

			err = rpcClient.CallContext(ctx, &txHash, "eth_sendTransaction", tx)
			if err != nil {
				return fmt.Errorf("failed to send tx: %w", err)
			}

			_, err = bind.WaitMinedHash(ctx, ethclient, txHash)
			if err != nil {
				return fmt.Errorf("failed to wait for tx: %w", err)
			}

			return nil
		},
	}
}

func pointerOf[T any](v T) *T {
	return &v
}

func EthToWei(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(n), big.NewInt(params.Ether))
}
