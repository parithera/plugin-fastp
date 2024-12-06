package js

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

// Start is a function that analyzes the source code directory and generates a software bill of materials (SBOM) output.
// It returns an sbomTypes.Output struct containing the analysis results.
func Start(sourceCodeDir string, codeclarityDB *bun.DB, platform string) types.Output {

	return ExecuteScript(sourceCodeDir, platform)

}

func ExecuteScript(sourceCodeDir string, platform string) types.Output {
	start := time.Now()

	files, err := filepath.Glob(sourceCodeDir + "/*R1*.fastq.gz")
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		return generate_output(start, "no fastq file", codeclarity.SUCCESS, []exceptionManager.Error{})
	}
	outputPath := path.Join(sourceCodeDir, "fastp")
	os.MkdirAll(outputPath, os.ModePerm)

	// We retrieves the pairs of R1 and R2
	var groups [][]string
	for _, r1 := range files {
		var couple []string
		// With 10x r1 and r2 are switched
		if platform == "10x" {
			couple = append(couple, strings.ReplaceAll(r1, "R1", "R2"), r1)
		} else {
			couple = append(couple, r1, strings.ReplaceAll(r1, "R1", "R2"))
		}
		groups = append(groups, couple)
	}

	// We run fastp on each pair
	for _, group := range groups {
		// ./bin/fastp -i $first -I $second -h $outputdir/fastp/$output_name.html -j $outputdir/fastp/$output_name.json -w 8
		output_name := strings.ReplaceAll(group[0], sourceCodeDir+"/", "")
		output_name = strings.ReplaceAll(output_name, ".fastq.gz", "")
		args := append(
			[]string{
				"-i",
				group[0],
				"-I",
				group[1],
				"-h",
				path.Join(outputPath, output_name+".html"),
				"-j",
				path.Join(outputPath, output_name+".json"),
				"-w",
				"8",
			}, group...)

		// Run Rscript in sourceCodeDir
		cmd := exec.Command("fastp", args...)
		_, err = cmd.CombinedOutput()
		if err != nil {
			// panic(fmt.Sprintf("Failed to run Rscript: %s", err.Error()))
			codeclarity_error := exceptionManager.Error{
				Private: exceptionManager.ErrorContent{
					Description: err.Error(),
					Type:        exceptionManager.GENERIC_ERROR,
				},
				Public: exceptionManager.ErrorContent{
					Description: "The script failed to execute",
					Type:        exceptionManager.GENERIC_ERROR,
				},
			}
			return generate_output(start, nil, codeclarity.FAILURE, []exceptionManager.Error{codeclarity_error})
		}
	}

	return generate_output(start, "done", codeclarity.SUCCESS, []exceptionManager.Error{})
}

func generate_output(start time.Time, data any, status codeclarity.AnalysisStatus, errors []exceptionManager.Error) types.Output {
	formattedStart, formattedEnd, delta := output_generator.GetAnalysisTiming(start)

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
