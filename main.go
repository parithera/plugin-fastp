package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	amqp_helper "github.com/CodeClarityCE/utility-amqp-helper"
	dbhelper "github.com/CodeClarityCE/utility-dbhelper/helper"
	types_amqp "github.com/CodeClarityCE/utility-types/amqp"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	plugin_db "github.com/CodeClarityCE/utility-types/plugin_db"
	plugin "github.com/parithera/plugin-fastp/src"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Arguments encapsulates dependencies passed to the callback function.
type Arguments struct {
	codeclarity *bun.DB
}

// main is the entry point of the program.
// It reads the configuration, initializes the database connection,
// and starts listening on the AMQP queue for analysis requests.
func main() {
	config, err := readConfig()
	if err != nil {
		log.Printf("Error reading configuration: %v", err)
		return
	}

	// Retrieve database connection parameters from environment variables.
	host := os.Getenv("PG_DB_HOST")
	if host == "" {
		log.Printf("PG_DB_HOST is not set")
		return
	}
	port := os.Getenv("PG_DB_PORT")
	if port == "" {
		log.Printf("PG_DB_PORT is not set")
		return
	}
	user := os.Getenv("PG_DB_USER")
	if user == "" {
		log.Printf("PG_DB_USER is not set")
		return
	}
	password := os.Getenv("PG_DB_PASSWORD")
	if password == "" {
		log.Printf("PG_DB_PASSWORD is not set")
		return
	}

	// Construct the database connection string.
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbhelper.Config.Database.Results)

	// Initialize the database connection using Bun.
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn), pgdriver.WithTimeout(50*time.Second)))
	db_codeclarity := bun.NewDB(sqldb, pgdialect.New())
	defer db_codeclarity.Close()

	// Create an instance of the Arguments struct to pass dependencies.
	args := Arguments{
		codeclarity: db_codeclarity,
	}

	// Start listening on the AMQP queue for analysis requests.
	amqp_helper.Listen("dispatcher_"+config.Name, callback, args, config)
}

// startAnalysis performs the analysis of a project and generates an SBOM (Software Bill of Materials).
// It takes the analysis configuration, downloads the sample, executes the plugin, and stores the results.
// Parameters:
//
//	args: Arguments containing the database connection.
//	dispatcherMessage: Message from the dispatcher containing analysis details.
//	config: Plugin configuration.
//	analysis_document: Analysis document containing the analysis configuration.
//
// Returns:
//
//	A map containing the analysis result, the analysis status, and any error encountered.
func startAnalysis(args Arguments, dispatcherMessage types_amqp.DispatcherPluginMessage, config plugin_db.Plugin, analysis_document codeclarity.Analysis) (map[string]any, codeclarity.AnalysisStatus, error) {

	// Retrieve analysis configuration from the analysis document.
	messageData := analysis_document.Config[config.Name].(map[string]any)

	// Retrieve the download path from the environment variables.
	path := os.Getenv("DOWNLOAD_PATH")

	// Construct the destination folder path.
	sample := filepath.Join(path, dispatcherMessage.OrganizationId.String(), "samples", messageData["sample"].(string))

	// Execute the plugin with the sample and database connection.
	rOutput := plugin.Start(sample, args.codeclarity, messageData["platform"].(string))

	// Create a result object to store the analysis result in the database.
	result := codeclarity.Result{
		Result:     rOutput,
		AnalysisId: dispatcherMessage.AnalysisId,
		Plugin:     config.Name,
	}

	// Insert the result into the database.
	_, err := args.codeclarity.NewInsert().Model(&result).Exec(context.Background())
	if err != nil {
		return nil, codeclarity.FAILURE, fmt.Errorf("error inserting result into database: %w", err)
	}

	// Prepare the result map to store the sbom key.
	res := make(map[string]any)
	res["rKey"] = result.Id

	// Return the result map, analysis status, and nil error.
	return res, rOutput.AnalysisInfo.Status, nil
}
