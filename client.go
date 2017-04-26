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

	"golang.org/x/net/context"
)

// NewClient creates a new ReportingClient.
func NewClient(baseURI, token string) (*ReportingClient, error) {
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

	return &ReportingClient{
		baseURI: baseURI,
		token:   token,
	}, nil
}

// ReportingClient will serialize all calls made to it, and sent them to
// baseURI.
type ReportingClient struct {
	baseURI string
	token   string
}

// RunStarted serializes args and sends them to a remote reporter service.
func (c *ReportingClient) RunStarted(ctx context.Context, args RunStartedArgs) error {
	return c.postJSON("runstarted", args)
}

// RunFinished serializes args and sends them to a remote reporter service.
func (c *ReportingClient) RunFinished(ctx context.Context, args RunFinishedArgs) error {
	return c.postJSON("runfinished", args)
}

// RunStepsAdded serializes args and sends them to a remote reporter service.
func (c *ReportingClient) RunStepsAdded(ctx context.Context, args RunStepsAddedArgs) error {
	return c.postJSON("runstepsadded", args)
}

// RunStepStarted serializes args and sends them to a remote reporter service.
func (c *ReportingClient) RunStepStarted(ctx context.Context, args RunStepStartedArgs) error {
	return c.postJSON("runstepstarted", args)
}

// RunStepFinished serializes args and sends them to a remote reporter service.
func (c *ReportingClient) RunStepFinished(ctx context.Context, args RunStepFinishedArgs) error {
	return c.postJSON("runstepfinished", args)
}

// RunStepLogs serializes args and sends them to a remote reporter service.
func (c *ReportingClient) RunStepLogs(ctx context.Context, args RunStepLogsArgs) error {
	return c.postJSON("runsteplogs", args)
}

// JobStarted serializes args and sends them to a remote reporter service.
func (c *ReportingClient) JobStarted(ctx context.Context, args JobStartedArgs) error {
	return c.postJSON("jobstarted", args)
}

// JobError serializes args and sends them to a remote reporter service.
func (c *ReportingClient) JobError(ctx context.Context, args JobErrorArgs) error {
	return c.postJSON("joberror", args)
}

// JobFinished serializes args and sends them to a remote reporter service.
func (c *ReportingClient) JobFinished(ctx context.Context, args JobFinishedArgs) error {
	return c.postJSON("jobfinished", args)
}

func (c *ReportingClient) generateURL(path string) string {
	return fmt.Sprintf("%s/%s", c.baseURI, path)
}

func (c *ReportingClient) postJSON(path string, payload interface{}) error {
	body, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	uri, err := url.Parse(c.generateURL(path))
	if err != nil {
		return err
	}

	q := uri.Query()
	q.Set("wercker_token", c.token)
	uri.RawQuery = q.Encode()

	resp, err := http.Post(uri.String(), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned invalid statuscode: %d", resp.StatusCode)
	}

	return nil
}
