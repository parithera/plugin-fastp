package main

import (
	"database/sql"
	"os"
	"testing"
	"time"

	dbhelper "github.com/CodeClarityCE/utility-dbhelper/helper"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	plugin "github.com/parithera/plugin-fastp/src"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func TestCreateNPMv1(t *testing.T) {
	os.Setenv("PG_DB_HOST", "127.0.0.1")
	os.Setenv("PG_DB_PORT", "5432")
	os.Setenv("PG_DB_USER", "postgres")
	os.Setenv("PG_DB_PASSWORD", "!ChangeMe!")

	dsn := "postgres://postgres:!ChangeMe!@127.0.0.1:5432/" + dbhelper.Config.Database.Results + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn), pgdriver.WithTimeout(50*time.Second)))
	db_codeclarity := bun.NewDB(sqldb, pgdialect.New())
	defer db_codeclarity.Close()

	sourceCodeDir := "/home/cedric/Documents/workspace/parithera-dev/private/bef81cdc-a99f-4856-bde5-7b2ee1968819/f804b115-6eb7-4a41-8476-a239cb744f4e"
	out := plugin.Start(sourceCodeDir, db_codeclarity, "10x")

	// Assert the expected values
	assert.NotNil(t, out)
	assert.Equal(t, codeclarity.SUCCESS, out.AnalysisInfo.Status)

	writeJSON(out, sourceCodeDir+"/fastqc.json")
}
