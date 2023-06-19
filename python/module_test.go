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
	"testing"

	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/google/go-cmp/cmp"
	"github.com/siddharthab/bazel-gazelle-python/internal"
	"github.com/siddharthab/bazel-gazelle-python/python/parser"
)

func TestModuleInPkgDeps(t *testing.T) {
	moduleMap := map[string]*Module{
		"pkg1.pkg2": {
			Result: parser.Result{
				Imports: []string{"pkg1", "pkg1.pkg2.subpkg1", "pkg1.pkg2.mod1"},
			},
			ImportSpec: "pkg1.pkg2",
			PkgPath:    "pkg1/pkg2",
			Name:       "",
		},
		"pkg1.pkg2.mod1": {
			Result: parser.Result{
				Imports: []string{"pkg1", "pkg1.pkg2.subpkg1", "pkg1.pkg2.mod2", "pkg1.pkg2.sym1", "pkg1.pkg2.mod2.sym2"},
			},
			ImportSpec: "pkg1.pkg2.mod1",
			PkgPath:    "pkg1/pkg2",
			Name:       "mod1",
		},
		"pkg1.pkg2.mod2": {
			Result: parser.Result{
				Imports: []string{"pkg1.pkg2.subpkg2"},
			},
			ImportSpec: "pkg1.pkg2.mod2",
			PkgPath:    "pkg1/pkg2",
			Name:       "mod2",
		},
	}
	subPackages := map[string]struct{}{
		"subpkg1": {},
		"subpkg2": {},
	}
	testCases := []struct {
		module                  *Module
		wantInPkgImports        map[*Module]struct{}
		wantInPkgImportsClosure map[*Module]struct{}
	}{
		{
			module: moduleMap["pkg1.pkg2"],
			wantInPkgImports: map[*Module]struct{}{
				moduleMap["pkg1.pkg2.mod1"]: {},
			},
			wantInPkgImportsClosure: map[*Module]struct{}{
				moduleMap["pkg1.pkg2.mod1"]: {},
				moduleMap["pkg1.pkg2.mod2"]: {},
			},
		},
		{
			module: moduleMap["pkg1.pkg2.mod1"],
			wantInPkgImports: map[*Module]struct{}{
				moduleMap["pkg1.pkg2"]:      {},
				moduleMap["pkg1.pkg2.mod2"]: {},
			},
			wantInPkgImportsClosure: map[*Module]struct{}{
				moduleMap["pkg1.pkg2"]:      {},
				moduleMap["pkg1.pkg2.mod2"]: {},
			},
		},
		{
			module:                  moduleMap["pkg1.pkg2.mod2"],
			wantInPkgImports:        map[*Module]struct{}{},
			wantInPkgImportsClosure: map[*Module]struct{}{},
		},
	}

	for i, testCase := range testCases {
		module := testCase.module
		module.InPkgDeps = make(map[*Module]struct{})
		module.ProcessImports(moduleMap, subPackages)
		got := testCase.module.InPkgDeps
		if diff := cmp.Diff(got, testCase.wantInPkgImports); diff != "" {
			t.Errorf("test %d: (-got, +want):%s", i, diff)
		}
	}
	for i, testCase := range testCases {
		module := testCase.module
		module.DepsClosure()
		got := testCase.module.InPkgDeps
		if diff := cmp.Diff(got, testCase.wantInPkgImportsClosure); diff != "" {
			t.Errorf("test %d: (-got, +want):%s", i, diff)
		}
	}
}

func TestGenerateRule(t *testing.T) {
	testCases := []struct {
		module        Module
		nameTemplate  string
		relPythonRoot string
		want          *rule.Rule
	}{
		// Top-level package.
		{
			module: Module{
				Name:      "",
				PkgPath:   "pkg1",
				Filename:  "__init__.py",
				InPkgDeps: map[*Module]struct{}{},
			},
			nameTemplate:  "{module_name}",
			relPythonRoot: "..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyLibrary, "pkg1")
				r.SetAttr("srcs", []string{"__init__.py"})
				r.SetAttr("imports", "..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				r.SetAttr("visibility", []string{visibilityPublic})
				return r
			}(),
		},
		// Top-level module.
		{
			module: Module{
				Name:      "mod",
				PkgPath:   "",
				Filename:  "mod.py",
				InPkgDeps: map[*Module]struct{}{},
			},
			nameTemplate:  "{module_name}",
			relPythonRoot: "..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyLibrary, "mod")
				r.SetAttr("srcs", []string{"mod.py"})
				r.SetAttr("imports", "..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				r.SetAttr("visibility", []string{visibilityPublic})
				return r
			}(),
		},
		// Subpackage.
		{
			module: Module{
				Name:      "",
				PkgPath:   "pkg1/subpkg1",
				Filename:  "__init__.py",
				InPkgDeps: map[*Module]struct{}{},
			},
			nameTemplate:  "{module_name}",
			relPythonRoot: "../..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyLibrary, "subpkg1")
				r.SetAttr("srcs", []string{"__init__.py"})
				r.SetAttr("imports", "../..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				r.SetAttr("visibility", []string{visibilityPublic})
				return r
			}(),
		},
		// __main__
		{
			module: Module{
				Name:      "__main__",
				PkgPath:   "pkg1",
				Filename:  "__main__.py",
				InPkgDeps: map[*Module]struct{}{},
			},
			nameTemplate:  "{module_name}",
			relPythonRoot: "..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyBinary, "pkg1_bin")
				r.SetAttr("srcs", []string{"__main__.py"})
				r.SetAttr("main", "__main__.py")
				r.SetAttr("imports", "..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				return r
			}(),
		},
		// __test__
		{
			module: Module{
				Name:      "__test__",
				PkgPath:   "pkg1",
				Filename:  "__test__.py",
				InPkgDeps: map[*Module]struct{}{},
			},
			nameTemplate:  "{module_name}",
			relPythonRoot: "..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyTest, "pkg1_test")
				r.SetAttr("srcs", []string{"__test__.py"})
				r.SetAttr("imports", "..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				return r
			}(),
		},
		// Module
		{
			module: Module{
				Name:      "foo",
				PkgPath:   "pkg1",
				Filename:  "foo.py",
				InPkgDeps: map[*Module]struct{}{},
			},
			nameTemplate:  "{module_name}",
			relPythonRoot: "..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyLibrary, "foo")
				r.SetAttr("srcs", []string{"foo.py"})
				r.SetAttr("imports", "..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				r.SetAttr("visibility", []string{visibilityPublic})
				return r
			}(),
		},
		// In package dep
		{
			module: Module{
				Name:     "bar",
				PkgPath:  "pkg1",
				Filename: "bar.py",
				InPkgDeps: map[*Module]struct{}{
					{Filename: "baz.py"}: {},
				},
			},
			nameTemplate:  "{module_name}",
			relPythonRoot: "..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyLibrary, "bar")
				r.SetAttr("srcs", []string{"bar.py", "baz.py"})
				r.SetAttr("imports", "..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				r.SetAttr("visibility", []string{visibilityPublic})
				return r
			}(),
		},
		// HasMainNameCheck == true
		{
			module: Module{
				Name:      "baz",
				PkgPath:   "pkg1",
				Filename:  "baz.py",
				InPkgDeps: map[*Module]struct{}{},
				Result: parser.Result{
					HasMainNameCheck: true,
				},
			},
			nameTemplate:  "{module_name}",
			relPythonRoot: "..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyBinary, "baz")
				r.SetAttr("srcs", []string{"baz.py"})
				r.SetAttr("main", "baz.py")
				r.SetAttr("imports", "..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				r.SetAttr("visibility", []string{visibilityPublic})
				return r
			}(),
		},
		// _test in name
		{
			module: Module{
				Name:     "foo_test",
				PkgPath:  "pkg1",
				Filename: "foo_test.py",
				InPkgDeps: map[*Module]struct{}{
					{Filename: "foo.py"}: {},
				},
			},
			nameTemplate:  "{module_name}",
			relPythonRoot: "..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyTest, "foo_test")
				r.SetAttr("srcs", []string{"foo.py", "foo_test.py"})
				r.SetAttr("imports", "..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				return r
			}(),
		},
		// Name template.
		{
			module: Module{
				Name:      "mod",
				PkgPath:   "pkg1",
				Filename:  "mod.py",
				InPkgDeps: map[*Module]struct{}{},
			},
			nameTemplate:  "gazelle_{module_name}",
			relPythonRoot: "..",
			want: func() *rule.Rule {
				r := rule.NewRule(kindPyLibrary, "gazelle_mod")
				r.SetAttr("srcs", []string{"mod.py"})
				r.SetAttr("imports", "..")
				r.SetAttr("tags", []string{tagGazelleManaged})
				r.SetAttr("visibility", []string{visibilityPublic})
				return r
			}(),
		},
	}

	for _, testCase := range testCases {
		name := internal.ImportSpec(testCase.module.PkgPath, testCase.module.Name)
		got := testCase.module.GenerateRule(testCase.nameTemplate, testCase.relPythonRoot)
		want := testCase.want
		if diff := cmp.Diff(got.Kind(), want.Kind()); diff != "" {
			t.Errorf("test %s: (-got, +want):%s", name, diff)
		}
		if diff := cmp.Diff(got.Name(), want.Name()); diff != "" {
			t.Errorf("test %s: (-got, +want):%s", name, diff)
		}
		if diff := cmp.Diff(got.AttrKeys(), want.AttrKeys()); diff != "" {
			t.Errorf("test %s: (-got, +want):%s", name, diff)
		}
		if diff := cmp.Diff(got.AttrStrings("srcs"), want.AttrStrings("srcs")); diff != "" {
			t.Errorf("test %s: (-got, +want):%s", name, diff)
		}
		if diff := cmp.Diff(got.AttrString("imports"), want.AttrString("imports")); diff != "" {
			t.Errorf("test %s: (-got, +want):%s", name, diff)
		}
		if diff := cmp.Diff(got.AttrStrings("deps"), want.AttrStrings("deps")); diff != "" {
			t.Errorf("test %s: (-got, +want):%s", name, diff)
		}
		if diff := cmp.Diff(got.AttrStrings("tags"), want.AttrStrings("tags")); diff != "" {
			t.Errorf("test %s: (-got, +want):%s", name, diff)
		}
		if diff := cmp.Diff(got.AttrStrings("visibility"), want.AttrStrings("visibility")); diff != "" {
			t.Errorf("test %s: (-got, +want):%s", name, diff)
		}
	}
}
