package create

import (
	"errors"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/ethereum/go-ethereum/cmd/golembase/account/pkg/useraccount"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
)

func Create() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new account",
		Action: func(c *cli.Context) error {
			privageKeyPath, err := xdg.ConfigFile(useraccount.PrivateKeyPath)
			if err != nil {
				return fmt.Errorf("failed to get config file path: %w", err)
			}

			fmt.Println("privageKeyPath", privageKeyPath)

			privageKeyBytes, err := os.ReadFile(privageKeyPath)
			switch {
			case errors.Is(err, os.ErrNotExist):
				privateKey, err := crypto.GenerateKey()
				if err != nil {
					return fmt.Errorf("failed to generate private key: %w", err)
				}
				privateKeyBytes := crypto.FromECDSA(privateKey)
				err = os.WriteFile(privageKeyPath, privateKeyBytes, 0600)
				if err != nil {
					return fmt.Errorf("failed to write private key: %w", err)
				}

				fmt.Println("Private key generated and saved to", privageKeyPath)
				fmt.Println("Address:", crypto.PubkeyToAddress(privateKey.PublicKey).Hex())

			case err != nil:
				return fmt.Errorf("failed to read private key: %w", err)
			default:
				privateKey, err := crypto.ToECDSA(privageKeyBytes)
				if err != nil {
					return fmt.Errorf("failed to deserialize private key: %w", err)
				}
				fmt.Println("Private key already exists")
				fmt.Println("Address:", crypto.PubkeyToAddress(privateKey.PublicKey).Hex())
			}

			return nil

		},
	}
}
