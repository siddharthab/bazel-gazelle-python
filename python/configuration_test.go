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

package python

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadManifest(t *testing.T) {
	testCases := []struct {
		content string
		prefix  string
		want    map[string]ExternalModule
	}{
		{
			want: map[string]ExternalModule{},
		},
		{
			content: "dist1\tpkg1.pkg2\tmod\tpy\ndist2\t\tmod\tso\ndist3\tpkg\t\tpy",
			prefix:  "pre_",
			want: map[string]ExternalModule{
				"pkg1.pkg2.mod": {
					Dist:        "dist1",
					PkgPath:     "pkg1/pkg2",
					Module:      "mod",
					BazelTarget: "@pre_dist1//:pkg",
					Type:        "py",
				},
				"mod": {
					Dist:        "dist2",
					PkgPath:     "",
					Module:      "mod",
					BazelTarget: "@pre_dist2//:pkg",
					Type:        "so",
				},
				"pkg": {
					Dist:        "dist3",
					PkgPath:     "pkg",
					Module:      "",
					BazelTarget: "@pre_dist3//:pkg",
					Type:        "py",
				},
			},
		},
	}

	for i, testCase := range testCases {
		got, err := readExternalModuleMapTSV(strings.NewReader(testCase.content), testCase.prefix)
		if err != nil {
			t.Errorf("test %d: unexpected error: %v", i, err)
			continue
		}
		if diff := cmp.Diff(got, testCase.want); diff != "" {
			t.Errorf("test %d: (-got, +want):%s", i, diff)
		}
	}
}
