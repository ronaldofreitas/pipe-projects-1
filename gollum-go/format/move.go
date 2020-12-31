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

package format

import (
	"github.com/trivago/gollum/core"
)

// Move formatter
//
// This formatter moves data from one location to another. When targeting a
// metadata key, the target key will be created or overwritten. When the source
// is the payload, it will be cleared.
//
// Examples
//
// This example moves the payload produced by consumer.Console to the metadata
// key data.
//
//  exampleConsumer:
//    Type: consumer.Console
//    Streams: stdin
//    Modulators:
//      - format.Move
//        Target: data
type Move struct {
	core.SimpleFormatter `gollumdoc:"embed_type"`
}

func init() {
	core.TypeRegistry.Register(Move{})
}

// Configure initializes this formatter with values from a plugin config.
func (format *Move) Configure(conf core.PluginConfigReader) {
}

// ApplyFormatter update message payload
func (format *Move) ApplyFormatter(msg *core.Message) error {
	srcData := format.GetSourceData(msg)

	format.SetTargetData(msg, srcData)
	format.SetSourceData(msg, nil)
	return nil
}
