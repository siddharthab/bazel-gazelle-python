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

import "github.com/bazelbuild/bazel-gazelle/rule"

const (
	tagGazelleManaged = "py-gazelle-managed"
)

const (
	kindPyLibrary = "py_library"
	kindPyBinary  = "py_binary"
	kindPyTest    = "py_test"
)

var kinds = map[string]rule.KindInfo{
	kindPyLibrary: {
		// We can not rely on sources as the match attribute, because
		// multiple targets may have the same sources if there is a cyclical
		// dependency between modules in the package. For example, a.py and
		// b.py can both depend on each other, and so py_library for both a
		// and b with each list both a.py and b.py.
		// TODO: Change this logic when we start having one src file per rule.
		// So match by the rule name instead (default).
		NonEmptyAttrs: map[string]bool{
			"srcs": true,
		},
		MergeableAttrs: map[string]bool{
			"srcs":    true,
			"imports": true,
		},
		ResolveAttrs: map[string]bool{
			"deps": true,
		},
	},
	kindPyBinary: {
		NonEmptyAttrs: map[string]bool{
			"srcs": true,
		},
		MergeableAttrs: map[string]bool{
			"srcs":    true,
			"imports": true,
			"main":    true,
		},
		ResolveAttrs: map[string]bool{
			"deps": true,
		},
	},
	kindPyTest: {
		NonEmptyAttrs: map[string]bool{
			"srcs": true,
		},
		MergeableAttrs: map[string]bool{
			"srcs":    true,
			"imports": true,
		},
		ResolveAttrs: map[string]bool{
			"deps": true,
		},
	},
}

// NOTE: End users can customize with Gazelle's generic kind_map directive.
var loads = []rule.LoadInfo{
	{
		Name: "@rules_python//python:defs.bzl",
		Symbols: []string{
			kindPyLibrary,
			kindPyBinary,
			kindPyTest,
		},
	},
}
