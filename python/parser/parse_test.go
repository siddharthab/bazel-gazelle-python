// Copyright 2023 The Bazel Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License.  You may obtain a copy
// of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  See the
// License for the specific language governing permissions and limitations under
// the License.

package parser

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	cases := []struct {
		content string
		want    Result
	}{
		// Base case.
		{"", Result{}},
		// Top-level imports.
		{"import mod1", Result{[]string{"mod1"}, false}},
		{"from mod1 import foo", Result{[]string{"mod1.foo"}, false}},
		{"from mod1 import foo, bar", Result{[]string{"mod1.bar", "mod1.foo"}, false}},
		{"from mod1 import (foo, bar)", Result{[]string{"mod1.bar", "mod1.foo"}, false}},
		{"from mod1 import foo, bar; import mod2.baz", Result{[]string{"mod1.bar", "mod1.foo", "mod2.baz"}, false}},
		// Conditional imports.
		{"if False:\n\timport mod1", Result{[]string{"mod1"}, false}},
		{"if False:\n\tfrom mod1 import foo", Result{[]string{"mod1.foo"}, false}},
		{"def fn():\n\timport foo", Result{[]string{"foo"}, false}},
		{"def fn():\n\tfrom mod1 import foo", Result{[]string{"mod1.foo"}, false}},
		// Type checking imports.
		{"from typing import TYPE_CHECKING\nif TYPE_CHECKING:\n\timport mod1", Result{[]string{"typing.TYPE_CHECKING"}, false}},
		{"import typing\nif typing.TYPE_CHECKING:\n\timport mod1", Result{[]string{"typing"}, false}},
		// Type checking imports -- negations.
		{"from typing import TYPE_CHECKING\nif not TYPE_CHECKING:\n\timport mod1\nelse:\n\timport mod2", Result{[]string{"mod1", "typing.TYPE_CHECKING"}, false}},
		{"import typing\nif not typing.TYPE_CHECKING:\n\timport mod1\nelse:\n\timport mod2", Result{[]string{"mod1", "typing"}, false}},
		// Main block.
		{"if __name__ == \"__main__\":\n\tmain()", Result{nil, true}},
	}

	for i, testCase := range cases {
		res, err := Parse(strings.NewReader(testCase.content), "")
		if err != nil {
			t.Errorf("test %d: unexpected error: %v", i, err)
			continue
		}
		if diff := cmp.Diff(res, testCase.want); diff != "" {
			t.Errorf("test %d: (-got, +want):%s", i, diff)
		}
	}
}
