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

package main

import "testing"

func TestParseWheelName(t *testing.T) {
	testCases := []struct {
		query       string
		wantName    string
		wantVersion string
	}{
		{
			"pip-18.0-py2.py3-none-any.whl", "pip", "18.0",
		},
	}

	for i, testCase := range testCases {
		gotName, gotVersion, err := parseWheelName(testCase.query)
		if err != nil {
			t.Errorf("test %d: unexpected error: %v", i, err)
			continue
		}
		if got, want := gotName, testCase.wantName; got != want {
			t.Errorf("test %d: %s != %s", i, got, want)
		}
		if got, want := gotVersion, testCase.wantVersion; got != want {
			t.Errorf("test %d: %s != %s", i, got, want)
		}
	}
}

func TestModuleForFilename(t *testing.T) {
	testCases := []struct {
		filename    string
		distInfoDir string
		dataDir     string
		wantPkg     string
		wantModule  string
		wantType    string
	}{
		{"", "", "", "", "", ""},
		{"__init__.py", "", "", "", "", "py"},
		{"a.py", "", "", "", "a", "py"},
		{"a.so", "", "", "", "a", "so"},
		{"pkg/__init__.py", "", "", "pkg", "", "py"},
		{"pkg1/pkg2/__init__.py", "", "", "pkg1.pkg2", "", "py"},
		{"pkg/a.py", "", "", "pkg", "a", "py"},
		{"pkg/a.so", "", "", "pkg", "a", "so"},
		{"pkg/a.build.so", "", "", "pkg", "a", "so"},
		{"pkg1/pkg2/a.py", "", "", "pkg1.pkg2", "a", "py"},
		{"pkg1/pkg2/a.so", "", "", "pkg1.pkg2", "a", "so"},
		{"dist-info/a.py", "dist-info", "", "", "", ""},
		{"data/a.py", "", "data", "", "", ""},
		{"data/purelib/a.py", "", "data", "", "a", "py"},
		{"data/platlib/a.py", "", "data", "", "a", "py"},
		{"data/platlib/pkg/a.py", "", "data", "pkg", "a", "py"},
		{"data/platlib/__init__.py", "", "data", "", "", "py"},
	}

	for i, testCase := range testCases {
		gotPkg, gotModule, gotType := moduleForFilename(testCase.filename, testCase.distInfoDir, testCase.dataDir)
		if got, want := gotPkg, testCase.wantPkg; got != want {
			t.Errorf("test %d: %s != %s", i, got, want)
		}
		if got, want := gotModule, testCase.wantModule; got != want {
			t.Errorf("test %d: %s != %s", i, got, want)
		}
		if got, want := gotType, testCase.wantType; got != want {
			t.Errorf("test %d: %s != %s", i, got, want)
		}
	}
}
