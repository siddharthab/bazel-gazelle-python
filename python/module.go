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
	"path"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/siddharthab/bazel-gazelle-python/python/parser"
)

const visibilityPublic = "//visibility:public"

// Module represents a parsed Python module.
type Module struct {
	parser.Result
	ImportSpec   string
	PkgPath      string
	Name         string
	Filename     string
	InPkgDeps    map[*Module]struct{} // Module deps within the package.
	ExPkgImports []string             // Import statements not satisfied from within the package.
}

// ProcessImports computes direct InPkgDeps and ExPkgImports.
func (module *Module) ProcessImports(moduleMap map[string]*Module, subPackages map[string]struct{}) {
	for _, imp := range module.Imports {
		dep := module.findInPkgImport(imp, moduleMap, subPackages)
		if dep != nil {
			module.InPkgDeps[dep] = struct{}{}
		} else {
			module.ExPkgImports = append(module.ExPkgImports, imp)
		}
	}
}
func (module *Module) findInPkgImport(imp string, moduleMap map[string]*Module, subPackages map[string]struct{}) *Module {
	ext := path.Ext(imp)
	impParent := strings.TrimSuffix(imp, ext)
	if ext != "" && impParent == strings.ReplaceAll(module.PkgPath, "/", ".") {
		// This could be a symbol from this package, or it could be a subpackage.
		// Preference is symbol > subpackage > module.
		// We currently do not parse symbols out of the modules, so let's ignore that.
		// TODO: Parse global symbols from the module into parser.Result; this will be tricky.
		if _, ok := subPackages[ext[1:]]; ok {
			// This is a subpackage.
			return nil
		}
	}
	dep, ok := moduleMap[imp]
	if ok {
		return dep
	}
	dep, ok = moduleMap[impParent]
	if ok {
		return dep
	}
	return nil
}

// DepsClosure expands the ModuleDeps to get all transitive deps.
func (module *Module) DepsClosure() {
	for dep := range module.InPkgDeps {
		module.depsClosureHelper(dep)
	}
}
func (module *Module) depsClosureHelper(dep *Module) {
	for depdep := range dep.InPkgDeps {
		if _, ok := module.InPkgDeps[depdep]; ok || depdep == module {
			// Cycle.
			continue
		}
		module.InPkgDeps[depdep] = struct{}{}
		module.depsClosureHelper(depdep)
	}
}

func (module Module) GenerateRule(nameTemplate, relPythonRoot string) *rule.Rule {
	var kind = kindPyLibrary
	if module.Name == "__test__" || strings.HasPrefix(module.Name, "test_") || strings.HasSuffix(module.Name, "_test") {
		kind = kindPyTest
	} else if module.Name == "__main__" || module.HasMainNameCheck {
		kind = kindPyBinary
	}

	name := module.Name
	pkgName := path.Base(module.PkgPath)
	switch name {
	case "":
		name = pkgName
	case "__main__":
		name = pkgName + "_bin"
	case "__test__":
		name = pkgName + "_test"
	}
	name = strings.ReplaceAll(nameTemplate, "{module_name}", name)

	rule := rule.NewRule(kind, name)
	if kind == kindPyBinary {
		rule.SetAttr("main", module.Filename)
	}
	rule.SetAttr("tags", []string{tagGazelleManaged})
	if !strings.HasPrefix(module.Name, "_") && kind != kindPyTest {
		rule.SetAttr("visibility", []string{visibilityPublic})
	}

	srcs := []string{module.Filename}
	for dep := range module.InPkgDeps {
		srcs = append(srcs, dep.Filename)
	}
	sort.Strings(srcs)
	rule.SetAttr("srcs", srcs)
	rule.SetAttr("imports", relPythonRoot)

	return rule
}
