package importkey

import (
	"fmt"
	"os"
	"strings"

	"github.com/adrg/xdg"
	"github.com/ethereum/go-ethereum/cmd/golembase/account/pkg/useraccount"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

func ImportAccount() *cli.Command {
	return &cli.Command{
		Name:  "import",
		Usage: "Import an account using a hex private key",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "privatekey",
				Aliases:  []string{"key"},
				Usage:    "Private key in hex format",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			hexKey := c.String("privatekey")
			// Remove 0x prefix if present
			hexKey = strings.TrimPrefix(hexKey, "0x")

			// Parse the hex private key
			privateKeyBytes, err := crypto.HexToECDSA(hexKey)
			if err != nil {
				return fmt.Errorf("invalid private key: %w", err)
			}

			// Get path to store the private key
			privateKeyPath, err := xdg.ConfigFile(useraccount.PrivateKeyPath)
			if err != nil {
				return fmt.Errorf("failed to get config file path: %w", err)
			}

			// Convert to bytes and save
			keyBytes := crypto.FromECDSA(privateKeyBytes)
			if err := os.MkdirAll(strings.TrimSuffix(privateKeyPath, "/private.key"), 0700); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			// Write the private key to file
			if err := os.WriteFile(privateKeyPath, keyBytes, 0600); err != nil {
				return fmt.Errorf("failed to write private key: %w", err)
			}

			address := crypto.PubkeyToAddress(privateKeyBytes.PublicKey)
			fmt.Println("Successfully imported account")
			fmt.Println("Address:", address.Hex())

			return nil
		},
	}
}
