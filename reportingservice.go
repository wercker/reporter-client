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

import "golang.org/x/net/context"

// ReportingService handles run related events.
type ReportingService interface {
	RunStarted(ctx context.Context, args RunStartedArgs) error
	RunFinished(ctx context.Context, args RunFinishedArgs) error

	RunStepsAdded(ctx context.Context, args RunStepsAddedArgs) error

	RunStepStarted(ctx context.Context, args RunStepStartedArgs) error
	RunStepFinished(ctx context.Context, args RunStepFinishedArgs) error
	RunStepLogs(ctx context.Context, args RunStepLogsArgs) error
}

// RunStartedArgs the arguments asscociated with the RunStarted event.
type RunStartedArgs struct {
	RunID string `json:"runId"`
}

// RunFinishedArgs the arguments asscociated with the RunFinished event.
type RunFinishedArgs struct {
	RunID string `json:"runId"`

	// Valid options: succes, failed
	Result string `json:"result"`
}

// RunStepsAddedArgs the arguments asscociated with the RunStepsAdded event.
type RunStepsAddedArgs struct {
	RunID string    `json:"runId"`
	Steps []NewStep `json:"steps"`
}

// NewStep represents a new Step. To be used with the RunStepAddedArgs.
type NewStep struct {
	StepSafeID  string `json:"stepSafeId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`

	// Valid options: mainSteps, finalSteps
	Phase string `json:"phase,omitempty"`
}

// RunStepStartedArgs the arguments asscociated with the RunStepStarted event.
type RunStepStartedArgs struct {
	RunID      string `json:"runId"`
	StepSafeID string `json:"stepSafeId"`
}

// RunStepFinishedArgs the arguments asscociated with the RunStepFinished event.
type RunStepFinishedArgs struct {
	RunID      string `json:"runId"`
	StepSafeID string `json:"stepSafeId"`

	Result              string `json:"result"` // passed || failed
	ArtifactURL         string `json:"artifactUrl,omitempty"`
	PackageURL          string `json:"packageUrl,omitempty"`
	Message             string `json:"message,omitempty"`
	WerckerYamlContents string `json:"werckerYamlContents,omitempty"`
	Duration            int64  `json:"duration,omitempty"`
}

// RunStepLogsArgs the arguments asscociated with the RunStepLogs event.
type RunStepLogsArgs struct {
	RunID      string `json:"runId"`
	StepSafeID string `json:"stepSafeId"`

	Logs   []byte `json:"logs"`
	Stream string `json:"stream"`
	Chunk  int    `json:"chunk"`
}
