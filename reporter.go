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

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// New creates a new Reporter. baseURI can be prefixed with http or https. If
// no protocol is specified, then http will be used. Trailing slashes are
// removed.
func New(baseURI string, token string) (*Reporter, error) {
	if baseURI == "" {
		return nil, errors.New("baseURI cannot be empty")
	}

	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	if !strings.HasPrefix(baseURI, "http://") &&
		!strings.HasPrefix(baseURI, "https://") {
		baseURI = fmt.Sprintf("http://%s", baseURI)
	}

	if strings.HasSuffix(baseURI, "/") {
		baseURI = baseURI[0 : len(baseURI)-1]
	}

	return &Reporter{
		baseURI: baseURI,
		token:   token,
	}, nil
}

// Reporter will report status back to wercker api.
type Reporter struct {
	baseURI string
	token   string
}

func (r *Reporter) generateURL(path string) string {
	return fmt.Sprintf("%s/%s", r.baseURI, path)
}

func (r *Reporter) postJSON(path string, payload interface{}) error {
	body, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	uri, err := url.Parse(r.generateURL(path))
	if err != nil {
		return err
	}

	q := uri.Query()
	q.Set("wercker_token", r.token)
	uri.RawQuery = q.Encode()

	println(string(body))

	resp, err := http.Post(uri.String(), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned invalid statuscode: %d", resp.StatusCode)
	}

	return nil
}

// PipelineStartedArgs are the arguments associated with PipelineStarted.
type PipelineStartedArgs struct {
	BuildID  string `json:"buildId,omitempty"`
	DeployID string `json:"deploymentId,omitempty"`
}

// PipelineStarted will report that the a Pipeline has started.
func (r *Reporter) PipelineStarted(args *PipelineStartedArgs) error {
	path := ""
	if args.BuildID != "" {
		path = "buildstarted"
	} else if args.DeployID != "" {
		path = "deploystarted"
	} else {
		return errors.New("Both BuildID and DeployID are not set")
	}

	return r.postJSON(path, args)
}

// PipelineFinishedArgs are the arguments associated with PipelineFinished.
type PipelineFinishedArgs struct {
	BuildID  string `json:"buildId,omitempty"`
	DeployID string `json:"-"`
	Result   string `json:"buildResult"`

	// Properties related to Deploy
	Log string `json:"-"`
	URL string `json:"-"`
}

// PipelineFinished will report that the a Pipeline has finished.
func (r *Reporter) PipelineFinished(args *PipelineFinishedArgs) error {
	var payload interface{}
	path := ""
	if args.BuildID != "" {
		payload = args
		path = "buildfinished"
	} else if args.DeployID != "" {
		payload = convertToDeployArgs(args) // defined in `reporter_legacy.go`
		path = "deployfinished"
	} else {
		return errors.New("Both BuildID and DeployID are not set")
	}

	return r.postJSON(path, payload)
}

// NewStep represents a new step.
type NewStep struct {
	// Required
	Name  string `json:"name"`
	Order int    `json:"order"`

	// Optional
	DisplayName string `json:"displayName,omitempty"`

	// Valid options: mainSteps, finalSteps
	Phase string `json:"phase,omitempty"`
}

// NewPipelineStepsArgs are the arguments associated with NewPipelineSteps.
type NewPipelineStepsArgs struct {
	BuildID  string     `json:"buildId,omitempty"`
	DeployID string     `json:"deployId,omitempty"`
	Steps    []*NewStep `json:"steps"`
}

// NewPipelineSteps will add new steps to the Pipeline. This needs to be done
// before reporting anything.
func (r *Reporter) NewPipelineSteps(args *NewPipelineStepsArgs) error {
	path := ""
	if args.BuildID != "" {
		path = "addbuildsteps"
	} else if args.DeployID != "" {
		path = "adddeploysteps"
	} else {
		return errors.New("Both BuildID and DeployID are not set")
	}

	return r.postJSON(path, args)
}

// PipelineStepStartedArgs are the arguments associated with
// PipelineStepStarted.
type PipelineStepStartedArgs struct {
	BuildID  string `json:"buildId,omitempty"`
	DeployID string `json:"deployId,omitempty"`
	StepName string `json:"step"`
	Order    int    `json:"order"`
}

// PipelineStepStarted will report that the step has started.
func (r *Reporter) PipelineStepStarted(args *PipelineStepStartedArgs) error {
	path := ""
	if args.BuildID != "" {
		path = "buildstepstarted"
	} else if args.DeployID != "" {
		path = "deploystepstarted"
	} else {
		return errors.New("Both BuildID and DeployID are not set")
	}

	return r.postJSON(path, args)
}

// PipelineStepFinishedArgs are the arguments associated with
// PipelineStepFinished.
type PipelineStepFinishedArgs struct {
	BuildID               string `json:"buildId,omitempty"`
	DeployID              string `json:"deployId,omitempty"`
	StepName              string `json:"step"`
	Order                 int    `json:"order"`
	Successful            bool   `json:"isStepSuccessful"`
	ArtifactURL           string `json:"artifactsUrl,omitempty"`
	PackageURL            string `json:"packageUrl,omitempty"`
	Message               string `json:"message,omitempty"`
	WerckerYamlContents   string `json:"werckerYamlContents,omitempty"`
	WerckerConfigContents string `json:"werckerConfigContents,omitempty"`
}

// PipelineStepFinished will report that the step has finished.
func (r *Reporter) PipelineStepFinished(args *PipelineStepFinishedArgs) error {
	path := ""
	if args.BuildID != "" {
		path = "reportbuildstep"
	} else if args.DeployID != "" {
		path = "reportdeploystep"
	} else {
		return errors.New("Both BuildID and DeployID are not set")
	}

	return r.postJSON(path, args)
}

// PipelineStepReporterArgs are the arguments associated with
// PipelineStepReporter.
type PipelineStepReporterArgs struct {
	BuildID  string
	DeployID string
	StepName string
	Order    int
}

// PipelineStepReporter will return a io.WriterCloser which will serialize logs
// and post these to wercker api.
func (r *Reporter) PipelineStepReporter(args *PipelineStepReporterArgs) (*LogWriter, error) {
	return NewLogWriter(r.baseURI, r.token, args)
}
