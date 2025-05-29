package etlworld

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type etlProcess struct {
	*exec.Cmd
	output  *bytes.Buffer
	dbPath  string
	cleanup func()
}

func startETLProcess(
	ctx context.Context,
	slqliteETHBinaryPath string,
	walDir string,
	rpcEndpoint string,
) (_ *etlProcess, err error) {
	// Start geth in dev mode

	td, err := os.MkdirTemp("", "sqlite-etl")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	dbPath := filepath.Join(td, "db")

	// open the database in WAL mode, making sure that there is no race condition between the etl process and the further clients
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL", dbPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	err = db.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close database: %w", err)
	}

	cmd := exec.CommandContext(
		ctx,
		slqliteETHBinaryPath,
		"--db",
		dbPath,
		"--wal",
		walDir,
		"--rpc-endpoint",
		rpcEndpoint,
	)

	output := &bytes.Buffer{}

	cmd.Stdout = output
	cmd.Stderr = output

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start sqlite-etl: %w", err)
	}

	defer func() {
		if err != nil {
			ek := cmd.Process.Kill()
			if ek != nil {
				err = errors.Join(err, ek)
			}
			_, we := cmd.Process.Wait()
			if we != nil {
				err = errors.Join(err, we)
			}
			os.RemoveAll(td)
		}
	}()

	// Return cleanup function
	cleanup := func() {
		cmd.Process.Kill()
	}

	return &etlProcess{
		Cmd:     cmd,
		output:  output,
		dbPath:  dbPath,
		cleanup: cleanup,
	}, nil
}
