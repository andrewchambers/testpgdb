package testpgdb

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	extratesting "github.com/andrewchambers/go-extra/testing"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type TestDB struct {
	testDbContext context.Context
	baseDir       string
}

func (tdb *TestDB) SqlxDial(t *testing.T, options string) *sqlx.DB {
	db, err := sqlx.ConnectContext(tdb.testDbContext, "postgres", "host="+tdb.baseDir+" "+options)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func MakeDB(t *testing.T) (*TestDB, func()) {
	ctx, cancelContext := context.WithCancel(context.Background())

	baseDir, dirCleanup := extratesting.ScratchDir(t)
	dataDir := filepath.Join(baseDir, "data")

	cmd := exec.CommandContext(ctx, "initdb", dataDir, "-U", "postgres")
	err := cmd.Run()
	if err != nil {
		cancelContext()
		dirCleanup()
		t.Fatal(err)
	}

	config, err := os.OpenFile(filepath.Join(dataDir, "postgresql.conf"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer config.Close()

	_, err = config.WriteString("listen_addresses=''\n")
	if err != nil {
		t.Fatal(err)
	}

	_, err = fmt.Fprintf(config, "unix_socket_directories='%s'\n", baseDir)
	if err != nil {
		t.Fatal(err)
	}

	err = config.Close()
	if err != nil {
		t.Fatal(err)
	}

	cmd = exec.CommandContext(ctx, "pg_ctl", "-w", "-D", dataDir, "start")
	cmd.Dir = baseDir
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		cancelContext()
		cmd = exec.Command("pg_ctl", "-w", "-D", dataDir, "stop")
		cmd.Dir = baseDir
		err = cmd.Run()
		if err != nil {
			t.Fatal(err)
		}
		dirCleanup()
	}

	return &TestDB{
		testDbContext: ctx,
		baseDir:       baseDir,
	}, cleanup
}
