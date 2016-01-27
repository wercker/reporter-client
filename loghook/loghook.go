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

package loghook

import (
	"errors"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/wercker/reporter"
)

// New will create a new LogHook
func New(r *reporter.Reporter) *LogHook {
	h := &LogHook{
		writers: make(map[string]io.WriteCloser),
		r:       r,
	}
	return h
}

// A LogHook is a hook for Logrus, which serializes Logs if they contain the
// property Report. (deprecated?)
type LogHook struct {
	writers map[string]io.WriteCloser
	r       *reporter.Reporter
}

func (l *LogHook) logToWerckerEndpoint(e *log.Entry) error {
	var stepName string
	if s, ok := e.Data["StepName"]; ok {
		if stepName, ok = s.(string); !ok {
			return errors.New("StepName is not a string")
		}
	} else {
		return errors.New("StepName is not specified")
	}

	var buildID string
	if b, ok := e.Data["BuildID"]; ok {
		if buildID, ok = b.(string); !ok {
			return errors.New("BuildID is not a string")
		}
	} else {
		return errors.New("BuildID is not specified")
	}

	var order int
	if o, ok := e.Data["Order"]; ok {
		if order, ok = o.(int); !ok {
			return errors.New("Order is not a int")
		}
	} else {
		return errors.New("Order is not specified")
	}

	var w io.WriteCloser
	var ok bool
	if w, ok = l.writers[stepName]; !ok {
		var err error
		w, err = l.r.StepOutput(buildID, stepName, order)
		if err != nil {
			return nil
		}

		l.writers[stepName] = w
	}

	w.Write([]byte(e.Message))

	return nil
}

// Fire occurs when a call is made to Logrus and this Hook is added.
func (l *LogHook) Fire(e *log.Entry) error {
	if r, ok := e.Data["Report"]; ok {
		if shouldReport, ok := r.(bool); ok && shouldReport {
			return l.logToWerckerEndpoint(e)
		}
	}

	return nil
}

// Levels will return the levels for which this hook should fire. Currently
// only InfoLevel.
func (l *LogHook) Levels() []log.Level {
	return []log.Level{
		log.InfoLevel,
	}
}

// Close will close all LogWriters that were created during the logging.
func (l *LogHook) Close() error {
	for _, v := range l.writers {
		err := v.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
