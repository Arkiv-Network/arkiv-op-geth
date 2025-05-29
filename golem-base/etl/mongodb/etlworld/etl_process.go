package etlworld

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type etlProcess struct {
	*exec.Cmd
	output  *bytes.Buffer
	dbPath  string
	cleanup func()
}

func startETLProcess(
	ctx context.Context,
	mongoETLBinaryPath string,
	walDir string,
	rpcEndpoint string,
	mongoURI string,
	dbName string,
) (_ *etlProcess, err error) {
	// Create command with MongoDB-specific arguments
	cmd := exec.CommandContext(
		ctx,
		mongoETLBinaryPath,
		"--wal", walDir,
		"--rpc-endpoint", rpcEndpoint,
		"--mongo-uri", mongoURI,
		"--db-name", dbName,
	)

	output := &bytes.Buffer{}

	cmd.Stdout = output
	cmd.Stderr = output

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start mongodb-etl: %w", err)
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
		}
	}()

	// Return cleanup function
	cleanup := func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Process.Wait()
		}
	}

	return &etlProcess{
		Cmd:     cmd,
		output:  output,
		cleanup: cleanup,
	}, nil
}
