// Copyright 2015-2018 trivago N.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/trivago/gollum/logger"
)

// LogConsumer is an internal consumer plugin used indirectly by the gollum log
// package.
type LogConsumer struct {
	Consumer
	control         chan PluginControl
	logRouter       Router
	metric          string
	stopped         bool
	queue           MessageQueue
	metricErrors    metrics.Counter
	metricWarning   metrics.Counter
	metricInfo      metrics.Counter
	metricDebug     metrics.Counter
	metricsRegistry metrics.Registry
}

// Configure initializes this consumer with values from a plugin config.
func (cons *LogConsumer) Configure(conf PluginConfigReader) {
	cons.control = make(chan PluginControl, 1)
	cons.logRouter = StreamRegistry.GetRouter(LogInternalStreamID)
	cons.metric = conf.GetString("MetricKey", "")
	cons.queue = NewMessageQueue(1024)

	if cons.metric != "" {
		cons.metricsRegistry = NewMetricsRegistry(cons.metric)

		cons.metricErrors = metrics.NewCounter()
		cons.metricWarning = metrics.NewCounter()
		cons.metricInfo = metrics.NewCounter()
		cons.metricDebug = metrics.NewCounter()
		cons.metricsRegistry.Register("errors", cons.metricErrors)
		cons.metricsRegistry.Register("warnings", cons.metricWarning)
		cons.metricsRegistry.Register("info", cons.metricInfo)
		cons.metricsRegistry.Register("debug", cons.metricDebug)
	}
}

// GetState always returns PluginStateActive
func (cons *LogConsumer) GetState() PluginState {
	if cons.stopped {
		return PluginStateDead
	}
	return PluginStateActive
}

// Streams always returns an array with one member - the internal log stream
func (cons *LogConsumer) Streams() []MessageStreamID {
	return []MessageStreamID{LogInternalStreamID}
}

// IsBlocked always returns false
func (cons *LogConsumer) IsBlocked() bool {
	return false
}

// GetID returns the pluginID of the message source
func (cons *LogConsumer) GetID() string {
	return "core.LogConsumer"
}

// GetShutdownTimeout always returns 1 millisecond
func (cons *LogConsumer) GetShutdownTimeout() time.Duration {
	return time.Millisecond
}

// Control returns a handle to the control channel
func (cons *LogConsumer) Control() chan<- PluginControl {
	return cons.control
}

// Consume starts listening for control statements
func (cons *LogConsumer) Consume(threads *sync.WaitGroup) {
	// Wait for control statements
	for {
		select {
		case msg := <-cons.queue:
			cons.logRouter.Enqueue(msg)

		case command := <-cons.control:
			if command == PluginControlStopConsumer {
				cons.queue.Close()
				for msg := range cons.queue {
					cons.logRouter.Enqueue(msg)
				}
				cons.stopped = true
				return // ### return ###
			}
		}
	}
}

// Levels and Fire() implement the logrus.Hook interface
func (cons *LogConsumer) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire and Levels() implement the logrus.Hook interface
func (cons *LogConsumer) Fire(logrusEntry *logrus.Entry) error {
	// Have Logrus format the log entry
	formattedMessage, err := logrusEntry.String()
	if err != nil {
		return err
	}

	// The formatter adds an unnecessary linefeed, strip it out
	if formattedMessage[len(formattedMessage)-1] == '\n' {
		formattedMessage = formattedMessage[:len(formattedMessage)-1]
	}

	// Set message metadata: level, time and logrus's ad-hoc fields. The fields
	// also contain the plugin-specific log scope.
	metadata := NewMetadata()
	metadata.Set("Level", logrusEntry.Level.String())
	metadata.Set("Time", logrusEntry.Time.String())
	//  string,    interface{}
	for fieldName, fieldValue := range logrusEntry.Data {
		metadata.Set(fieldName, []byte(fmt.Sprintf("%v", fieldValue)))
	}

	// Wrap it in a Gollum message
	msg := NewMessage(cons, []byte(formattedMessage), metadata, LogInternalStreamID)

	// Push it to a channel to allow logging while logging
	// In case Push fails, fallback to stdout.
	result := cons.queue.Push(msg, time.Second)
	if result != MessageQueueOk {
		fmt.Fprintln(logger.FallbackLogDevice, msg.String())
	}

	// Metrics handling from .Write() (TODO: support all message levels?)
	if cons.metric != "" {
		switch logrusEntry.Level {
		case logrus.ErrorLevel:
			cons.metricErrors.Inc(1)

		case logrus.WarnLevel:
			cons.metricWarning.Inc(1)

		case logrus.InfoLevel:
			cons.metricInfo.Inc(1)

		case logrus.DebugLevel:
			cons.metricDebug.Inc(1)
		}
	}

	// Success
	return nil
}
