package entity

import (
	"github.com/ethereum/go-ethereum/cmd/golembase/entity/create"
	"github.com/ethereum/go-ethereum/cmd/golembase/entity/delete"
	"github.com/ethereum/go-ethereum/cmd/golembase/entity/history"
	"github.com/ethereum/go-ethereum/cmd/golembase/entity/list"
	"github.com/ethereum/go-ethereum/cmd/golembase/entity/update"
	"github.com/urfave/cli/v2"
)

func Entity() *cli.Command {
	return &cli.Command{
		Name:  "entity",
		Usage: "Manage entities",
		Subcommands: []*cli.Command{
			create.Create(),
			delete.Delete(),
			update.Update(),
			list.List(),
			history.History(),
		},
	}
}
