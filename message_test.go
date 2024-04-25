// Copyright 2024 The BxELab studyzy Authors
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

package runestone

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

func TestMessageFromIntegers(t *testing.T) {
	r := &Runestone{}
	payload, _ := hex.DecodeString("14f1a39f0114b2071601")
	integers, _ := r.integers(payload)

	message, err := MessageFromIntegers(&wire.MsgTx{}, integers)
	assert.NoError(t, err)
	t.Logf("message: %v", message)
	for tag, values := range message.Fields {
		t.Logf("tag: %v, values: %v", tag, values)
	}
	assert.Equal(t, 2, len(message.Fields))

}
