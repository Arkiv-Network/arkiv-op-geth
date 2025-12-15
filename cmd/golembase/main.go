package main

import (
	"log"
	"os"

	"github.com/ethereum/go-ethereum/cmd/golembase/account"
	"github.com/ethereum/go-ethereum/cmd/golembase/blocks"
	"github.com/ethereum/go-ethereum/cmd/golembase/cat"
	"github.com/ethereum/go-ethereum/cmd/golembase/entity"
	"github.com/ethereum/go-ethereum/cmd/golembase/query"
	"github.com/ethereum/go-ethereum/cmd/golembase/state"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name:  "golembase CLI",
		Usage: "Golem Base",

		Commands: []*cli.Command{
			account.Account(),
			entity.Entity(),
			// create.Create(),
			blocks.Blocks(),
			cat.Cat(),
			query.Query(),
			state.State(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
