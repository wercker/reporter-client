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
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"
)

const (
	// DefaultMaxChunkSize is the max bytes that the LogWriter will buffer until
	// flushing.
	DefaultMaxChunkSize = 10240
)

// NewLogWriter will create and initialize a new LogWriter. It is the callers
// responsiblity to call Close to ensure resources are cleaned up and logs are
// flushed.
func NewLogWriter(svc ReportingService, runID, stepSafeID, stream string) (*LogWriter, error) {
	if svc == nil {
		return nil, errors.New("svc cannot be nil")
	}

	if runID == "" {
		return nil, errors.New("RunID, BuildID and DeployID are not set")
	}

	cleanup := make(chan bool)

	w := &LogWriter{
		svc:        svc,
		runID:      runID,
		stepSafeID: stepSafeID,
		stream:     stream,

		buffer:       []byte{},
		currentChunk: 1,
		cleanup:      cleanup,
		maxChunkSize: DefaultMaxChunkSize,
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
	svc        ReportingService
	runID      string
	stepSafeID string
	stream     string

	buffer       []byte
	currentChunk int
	bufferLock   sync.Mutex
	chunkLock    sync.Mutex
	cleanup      chan bool
	maxChunkSize int
}

// SetMaxChunkSize changes the MaxChunkSize for w. Panics if n <= 0.
func (w *LogWriter) SetMaxChunkSize(n int) {
	if n <= 0 {
		panic(fmt.Sprintf("Invalid MaxChunkSize: %d", n))
	}

	w.maxChunkSize = n
}

// Write will buffer p and these will be send to the wercker api endpoint.
func (w *LogWriter) Write(p []byte) (int, error) {
	w.bufferLock.Lock()
	defer w.bufferLock.Unlock()

	w.buffer = append(w.buffer, p...)

	for len(w.buffer) > w.maxChunkSize {
		chunk := w.buffer[:w.maxChunkSize]
		w.buffer = w.buffer[w.maxChunkSize:]
		w.send(chunk)
	}

	return len(p), nil
}

// Close will stop sending the logs and it will flush out any logs still in the
// buffer.
func (w *LogWriter) Close() error {
	w.cleanup <- true
	return w.Flush()
}

// Flush will take the buffer and send it to the wercker api.
func (w *LogWriter) Flush() error {
	w.bufferLock.Lock()
	defer w.bufferLock.Unlock()

	var err error
	if len(w.buffer) > 0 {
		err = w.send(w.buffer)
		w.buffer = []byte{}
	}

	return err
}

func (w *LogWriter) send(b []byte) error {
	n := w.currentChunk
	w.currentChunk = n + 1

	args := RunStepLogsArgs{
		RunID:      w.runID,
		StepSafeID: w.stepSafeID,
		Logs:       b,
		Stream:     w.stream,
		Chunk:      n,
	}

	err := w.svc.RunStepLogs(context.TODO(), args)
	return err
}
