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

package internal

import "testing"

func TestModuleName(t *testing.T) {
	testCases := []struct {
		filename string
		wantName string
		wantType string
		wantOk   bool
	}{
		{"", "", "", false},
		{"foo", "", "", false},
		{"foo.bar", "", "", false},
		{"foo.py", "foo", "py", true},
		{"foo.build-info.so", "foo", "so", true},
	}

	for i, testCase := range testCases {
		gotName, gotType, gotOk := ModuleName(testCase.filename)
		if got, want := gotOk, testCase.wantOk; got != want {
			t.Errorf("test %d: %v != %v", i, got, want)
		}
		if got, want := gotName, testCase.wantName; got != want {
			t.Errorf("test %d: %q != %q", i, got, want)
		}
		if got, want := gotType, testCase.wantType; got != want {
			t.Errorf("test %d: %q != %q", i, got, want)
		}
	}
}

func TestImportSpecifier(t *testing.T) {
	testCases := []struct {
		pkgPath    string
		moduleName string
		want       string
	}{
		{"", "mod", "mod"},
		{"pkg", "", "pkg"},
		{"pkg", "mod", "pkg.mod"},
		{"pkg1/pkg2", "", "pkg1.pkg2"},
		{"pkg1/pkg2", "mod", "pkg1.pkg2.mod"},
	}

	for i, testCase := range testCases {
		got := ImportSpec(testCase.pkgPath, testCase.moduleName)
		if want := testCase.want; got != want {
			t.Errorf("test %d: %q != %q", i, got, want)
		}
	}
}
