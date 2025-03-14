package output_generator

import (
	"time"

	sbomTypes "github.com/CodeClarityCE/plugin-sbom-javascript/src/types/sbom/js"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
)

// GetAnalysisTiming calculates the analysis start time, end time, and duration.
// It takes the analysis start time as input and returns the formatted start time,
// formatted end time, and the duration in seconds.
func GetAnalysisTiming(startTime time.Time) (string, string, float64) {
	endTime := time.Now()
	elapsed := endTime.Sub(startTime)
	return startTime.Local().String(), endTime.Local().String(), elapsed.Seconds()
}

// WriteFailureOutput constructs a failure output for the analysis.
// It sets the analysis status to FAILURE, records the analysis timing,
// and includes any collected errors in the output.
// It takes the current output and the analysis start time as input and returns the updated output.
func WriteFailureOutput(output sbomTypes.Output, startTime time.Time) sbomTypes.Output {
	output.AnalysisInfo.Status = codeclarity.FAILURE
	formattedStart, formattedEnd, delta := GetAnalysisTiming(startTime)
	output.AnalysisInfo.Time.AnalysisStartTime = formattedStart
	output.AnalysisInfo.Time.AnalysisEndTime = formattedEnd
	output.AnalysisInfo.Time.AnalysisDeltaTime = delta

	output.AnalysisInfo.Errors = exceptionManager.GetErrors()

	return output
}
