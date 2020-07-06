/*
 Copyright 2020 Qiniu Cloud (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoverModeFlag(t *testing.T) {
	var tcs = []struct {
		value         string
		expectedValue interface{}
		err           interface{}
	}{
		{
			value:         "",
			expectedValue: "count",
			err:           nil,
		},
		{
			value:         "set",
			expectedValue: "set",
			err:           nil,
		},
		{
			value:         "count",
			expectedValue: "count",
			err:           nil,
		},
		{
			value:         "atomic",
			expectedValue: "atomic",
			err:           nil,
		},
		{
			value:         "xxxxx",
			expectedValue: "",
			err:           errors.New("unknown mode"),
		},
		{
			value:         "123333",
			expectedValue: "",
			err:           errors.New("unknown mode"),
		},
	}
	for _, tc := range tcs {
		mode := &CoverMode{}
		err := mode.Set(tc.value)
		actual := mode.String()
		assert.Equal(t, actual, tc.expectedValue, fmt.Sprintf("check mode flag value failed, expected %s, got %s", tc.expectedValue, actual))
		assert.Equal(t, err, tc.err, fmt.Sprintf("check mode flag error, expected %s, got %s", tc.err, err))
	}
}

func TestAgentPortFlag(t *testing.T) {
	var tcs = []struct {
		value         string
		expectedValue interface{}
		isErr         bool
	}{
		{
			value:         "",
			expectedValue: "",
			isErr:         false,
		},
		{
			value:         ":8888",
			expectedValue: ":8888",
			isErr:         false,
		},
		{
			value:         "8888",
			expectedValue: "",
			isErr:         true,
		},
		{
			value:         "::8888",
			expectedValue: "",
			isErr:         true,
		},
	}
	for _, tc := range tcs {
		agent := &AgentPort{}
		err := agent.Set(tc.value)
		if tc.isErr {
			assert.NotEqual(t, nil, err, fmt.Sprintf("check agentport flag error, expected %v, got %v", nil, err))
		} else {
			actual := agent.String()
			assert.Equal(t, tc.expectedValue, actual, fmt.Sprintf("check agentport flag value failed, expected %s, got %s", tc.expectedValue, actual))
		}
	}
}
