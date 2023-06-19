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
	"log"
	"path"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/siddharthab/bazel-gazelle-python/internal"
)

type Resolver struct{}

var _ resolve.Resolver = (*Resolver)(nil)

// Name implements resolve.Resolve.
func (pr Resolver) Name() string {
	return languageName
}

// Imports implements resolve.Resolver.
//
// Returns all Python module import specs defined by the files in "srcs" attribute.
func (pr Resolver) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	kind := r.Kind()
	if kind != kindPyLibrary && kind != kindPyBinary {
		return nil
	}

	config := c.Exts[languageName].(Configuration)
	pkgPath := config.PythonPackagePath
	if strings.HasPrefix(pkgPath, "..") {
		// We have not yet reached a Python workspace.
		return nil
	}

	var res []resolve.ImportSpec
	for _, src := range r.AttrStrings("srcs") {
		moduleName, _, ok := internal.ModuleName(src)
		if !ok {
			continue
		}
		res = append(res, resolve.ImportSpec{
			Lang: languageName,
			Imp:  internal.ImportSpec(pkgPath, moduleName),
		})
	}
	return res
}

// Embeds implements resolve.Resolver.
func (pr Resolver) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

// Resolve implements resolve.Resolver.
func (pr Resolver) Resolve(c *config.Config, ix *resolve.RuleIndex, _ *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {
	config := c.Exts[languageName].(Configuration)
	module := imports.(*Module)
	deps := make(map[string]struct{})
	for _, imp := range transitiveImports(module) {
		target, ok := findRuleByImportFuzzy(imp, ix, config.ExternalModuleMap, config.InternalModuleList)
		if target != "" {
			deps[target] = struct{}{}
			continue
		}
		if !ok {
			log.Printf("could not find Bazel rule for import %q", imp)
		}
	}
	// Depend on parent package for module initialization.
	imp := module.ImportSpec
	ext := path.Ext(imp)
	for ext != "" {
		imp = strings.TrimSuffix(imp, ext)
		ext = path.Ext(imp)
		if target, ok := findRuleByImport(imp, ix, nil, nil); ok && target != "" {
			deps[target] = struct{}{}
			break
		}
	}

	// Set the attribute on the rule.
	var depsAttr []string
	for dep := range deps {
		depsAttr = append(depsAttr, dep)
	}
	r.SetAttr("deps", depsAttr)
}

// Extract all ExPkgImports from modules and its sibling dependencies.
func transitiveImports(module *Module) []string {
	modules := []*Module{module}
	for dep := range module.InPkgDeps {
		modules = append(modules, dep)
	}
	var res []string
	for _, module := range modules {
		res = append(res, module.ExPkgImports...)
	}
	return res
}

func findRuleByImportFuzzy(imp string, ix *resolve.RuleIndex, externalModuleMap map[string]ExternalModule, internalModuleList map[string]struct{}) (string, bool) {
	// Check exact matches.
	if target, ok := findRuleByImport(imp, ix, externalModuleMap, internalModuleList); ok {
		return target, ok
	}
	// Check exact matches for parent import, in case the import specifier is for a symbol.
	ext := path.Ext(imp)
	if ext == "" {
		return "", false
	}
	return findRuleByImport(strings.TrimSuffix(imp, ext), ix, externalModuleMap, internalModuleList)
}

func findRuleByImport(imp string, ix *resolve.RuleIndex, externalModuleMap map[string]ExternalModule, internalModuleList map[string]struct{}) (string, bool) {
	results := ix.FindRulesByImport(resolve.ImportSpec{Lang: languageName, Imp: imp}, languageName)
	for _, res := range results {
		if res.Label.Name == path.Base(res.Label.Pkg) {
			return res.Label.String(), true
		}
	}
	if len(results) > 0 {
		return results[0].Label.String(), true
	}
	if externalModuleMap != nil {
		if dep, ok := externalModuleMap[imp]; ok {
			return dep.BazelTarget, true
		}
	}
	if internalModuleList != nil {
		if _, ok := internalModuleList[imp]; ok {
			return "", true
		}
	}
	return "", false
}
