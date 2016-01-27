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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// NewLogWriter will create and initialize a new LogWriter. It is the callers
// responsiblity to call Close to ensure resources are cleaned up and logs are
// flushed.
func NewLogWriter(baseURI, token string, args *PipelineStepReporterArgs) (*LogWriter, error) {
	if args.BuildID == "" && args.DeployID == "" {
		return nil, errors.New("Both BuildID and DeployID are not set")
	}

	cleanup := make(chan bool)

	w := &LogWriter{
		args:         args,
		baseURI:      baseURI,
		buffer:       new(bytes.Buffer),
		currentChunk: 1,
		cleanup:      cleanup,
		token:        token,
	}

	go func(w *LogWriter) {
		for {
			select {
			case <-time.After(time.Second * 3):
				w.Flush()
			case <-w.cleanup:
				return
			}
		}
	}(w)

	return w, nil
}

// A LogWriter sends everything written to it to the wercker api endpoint. It
// buffers calls and flushes every three seconds. Close will stop this loop and
// will flush out any bytes left in the buffer.
type LogWriter struct {
	args         *PipelineStepReporterArgs
	baseURI      string
	buffer       *bytes.Buffer
	currentChunk int
	l            sync.Mutex
	cleanup      chan bool
	token        string
}

// Write will buffer p and these will be send to the wercker api endpoint.
func (w *LogWriter) Write(p []byte) (int, error) {
	w.l.Lock()
	defer w.l.Unlock()
	n, err := w.buffer.Write(p)
	if err != nil {
		return -1, err
	}

	return n, nil
}

// Close will stop sending the logs and it will flush out any logs still in the
// buffer.
func (w *LogWriter) Close() error {
	w.cleanup <- true
	return w.Flush()
}

// Flush will take the buffer and send it to the wercker api.
func (w *LogWriter) Flush() error {
	var err error
	w.l.Lock()
	defer w.l.Unlock()

	if w.buffer.Len() > 0 {
		buf := w.buffer
		w.buffer = new(bytes.Buffer)

		n := w.currentChunk
		w.currentChunk = n + 1

		u := ""
		if w.args.BuildID != "" {
			u = fmt.Sprintf("%s/internal/builds/%s/steps/%s/%d/log/%d", w.baseURI, w.args.BuildID, w.args.StepName, w.args.Order, n)
		} else if w.args.DeployID != "" {
			u = fmt.Sprintf("%s/internal/deploys/%s/steps/%s/%d/log/%d", w.baseURI, w.args.DeployID, w.args.StepName, w.args.Order, n)
		}

		uri, err := url.Parse(u)
		if err != nil {
			return err
		}

		q := uri.Query()
		q.Set("wercker_token", w.token)
		uri.RawQuery = q.Encode()

		err = sendLogs(uri.String(), buf.Bytes())
	}

	return err
}

func sendLogs(url string, payload []byte) error {
	resp, err := http.Post(url, "text/plain", bytes.NewReader(payload))

	if err != nil {
		return err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("API returned invalid statuscode: %d", resp.StatusCode)
}
