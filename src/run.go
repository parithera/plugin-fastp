package fastp

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
	"github.com/uptrace/bun"

	"github.com/parithera/plugin-fastp/src/types"
	"github.com/parithera/plugin-fastp/src/utils/output_generator"
)

// Start analyzes the source code directory and generates a software bill of materials (SBOM) output.
// It returns an sbomTypes.Output struct containing the analysis results.
func Start(sourceCodeDir string, codeclarityDB *bun.DB, platform string) types.Output {
	return ExecuteScript(sourceCodeDir, platform)
}

// ExecuteScript performs the Fastp analysis on the provided data.
func ExecuteScript(sourceCodeDir string, platform string) types.Output {
	startTime := time.Now()

	// Find fastq files matching the pattern "*R1*.fastq.gz".
	files, err := filepath.Glob(sourceCodeDir + "/*R1*.fastq.gz")
	if err != nil {
		log.Fatal(err)
	}

	// If no fastq files are found, return a specific output.
	if len(files) == 0 {
		return generate_output(startTime, "no fastq file", codeclarity.SUCCESS, []exceptionManager.Error{})
	}

	// Create the output directory for Fastp results.
	outputPath := path.Join(sourceCodeDir, "fastp")
	os.MkdirAll(outputPath, os.ModePerm)

	// Prepare the pairs of R1 and R2 files for analysis.
	var filePairs [][]string
	for _, r1File := range files {
		var pair []string
		// Handle 10x Genomics data where R1 and R2 are switched.
		if platform == "10x" {
			pair = append(pair, strings.ReplaceAll(r1File, "R1", "R2"), r1File)
		} else {
			pair = append(pair, r1File, strings.ReplaceAll(r1File, "R1", "R2"))
		}
		filePairs = append(filePairs, pair)
	}

	// Iterate through the file pairs and run Fastp on each.
	for _, pair := range filePairs {
		// Construct the output name based on the R1 file.
		outputName := strings.ReplaceAll(pair[0], sourceCodeDir+"/", "")
		outputName = strings.ReplaceAll(outputName, ".fastq.gz", "")

		// Define the arguments for the Fastp command.
		args := append(
			[]string{
				"-i",
				pair[0],
				"-I",
				pair[1],
				"-h",
				path.Join(outputPath, outputName+".html"),
				"-j",
				path.Join(outputPath, outputName+".json"),
				"-w",
				"8",
			}, pair...)

		// Execute the Fastp command.
		cmd := exec.Command("fastp", args...)
		_, err = cmd.CombinedOutput()
		if err != nil {
			// Handle errors during Fastp execution.
			codeclarityError := exceptionManager.Error{
				Private: exceptionManager.ErrorContent{
					Description: err.Error(),
					Type:        exceptionManager.GENERIC_ERROR,
				},
				Public: exceptionManager.ErrorContent{
					Description: "The script failed to execute",
					Type:        exceptionManager.GENERIC_ERROR,
				},
			}
			return generate_output(startTime, nil, codeclarity.FAILURE, []exceptionManager.Error{codeclarityError})
		}
	}

	// Return a success output.
	return generate_output(startTime, "done", codeclarity.SUCCESS, []exceptionManager.Error{})
}

// generate_output creates the final output struct with analysis results and timing information.
func generate_output(startTime time.Time, data any, status codeclarity.AnalysisStatus, errors []exceptionManager.Error) types.Output {
	formattedStart, formattedEnd, delta := output_generator.GetAnalysisTiming(startTime)

	output := types.Output{
		Result: types.Result{
			Data: data,
		},
		AnalysisInfo: types.AnalysisInfo{
			Errors: errors,
			Time: types.Time{
				AnalysisStartTime: formattedStart,
				AnalysisEndTime:   formattedEnd,
				AnalysisDeltaTime: delta,
			},
			Status: status,
		},
	}
	return output
}
