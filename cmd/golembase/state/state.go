package state

import (
	"github.com/ethereum/go-ethereum/cmd/golembase/state/usedslots"
	"github.com/urfave/cli/v2"
)

func State() *cli.Command {
	return &cli.Command{
		Name:  "state",
		Usage: "Debug the state of the storage layer",
		Subcommands: []*cli.Command{
			usedslots.UsedSlots(),
		},
	}
}
