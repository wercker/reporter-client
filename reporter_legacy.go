//   Copyright 2016 Wercker Holding BV
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package reporter

// NOTE(bvdberg): This API has been deprecated, use the related Pipeline API
// instead.

import "io"

const (
	// SuccessfulDeployResult is the result of a deploy which is interpeted as a
	// Successful deploy.
	SuccessfulDeployResult = "passed"
)

// BuildStarted will report that the build has started.
func (r *Reporter) BuildStarted(buildID string) error {
	return r.PipelineStarted(&PipelineStartedArgs{
		BuildID: buildID,
	})
}

// BuildFinished will report that the build has finished with the result result.
func (r *Reporter) BuildFinished(buildID string, result string) error {
	return r.PipelineFinished(&PipelineFinishedArgs{
		BuildID: buildID,
		Result:  result,
	})
}

// ReportNewSteps will add steps to the buildstep. This needs to be done before
// reporting anything.
func (r *Reporter) ReportNewSteps(buildID string, steps []*NewStep) error {
	return r.NewPipelineSteps(&NewPipelineStepsArgs{
		BuildID: buildID,
		Steps:   steps,
	})
}

// StepStarted will report that the step has started.
func (r *Reporter) StepStarted(buildID, stepName string, order int) error {
	return r.PipelineStepStarted(&PipelineStepStartedArgs{
		BuildID:  buildID,
		StepName: stepName,
		Order:    order,
	})
}

// StepFinished will report that the step has finished with the status successful.
func (r *Reporter) StepFinished(buildID, stepName string, order int, successful bool, artifactURL string, packageURL string, message string, werckerYamlContents string) error {
	werckerConfigContents := ""
	if werckerYamlContents != "" {
		werckerConfigContents = werckerYamlContents
	}

	return r.PipelineStepFinished(&PipelineStepFinishedArgs{
		BuildID:               buildID,
		StepName:              stepName,
		Order:                 order,
		Successful:            successful,
		ArtifactURL:           artifactURL,
		PackageURL:            packageURL,
		Message:               message,
		WerckerYamlContents:   werckerYamlContents,
		WerckerConfigContents: werckerConfigContents,
	})
}

// StepOutput will return a io.WriterCloser which will serialize logs and post
// these to wercker api.
func (r *Reporter) StepOutput(buildID, stepName string, order int) (io.WriteCloser, error) {
	return r.PipelineStepReporter(&PipelineStepReporterArgs{
		BuildID:  buildID,
		StepName: stepName,
		Order:    order,
	})
}

func convertToDeployArgs(args *PipelineFinishedArgs) *deployFinishedArgs {
	return &deployFinishedArgs{
		DeployID: args.DeployID,
		Result: &deployFinishedResult{
			Success: args.Result == SuccessfulDeployResult,
			Log:     args.Log,
			URL:     args.URL,
		},
	}
}

type deployFinishedArgs struct {
	DeployID string                `json:"deploymentId"`
	Result   *deployFinishedResult `json:"result"`
}

type deployFinishedResult struct {
	Success bool   `json:"success"`
	Log     string `json:"log,omitempty"`
	URL     string `json:"url,omitempty"`
}
