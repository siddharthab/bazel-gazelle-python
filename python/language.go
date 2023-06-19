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
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func NewLanguage() language.Language {
	return &Language{}
}

type Language struct {
	Configurer
	Resolver
}

// Kinds implements language.Language.
func (l *Language) Kinds() map[string]rule.KindInfo {
	return kinds
}

// Loads implements language.Language.
func (l *Language) Loads() []rule.LoadInfo {
	return loads
}

// GenerateRules implements language.Language.
//
// Generates one py_library (or py_binary) rule for each Python module,
// including one for __init__.py with name as the last component of the Python
// package name. If modules within a package depend on each other, then their
// src files are included in multiple rules, which allows us to have cyclical
// dependencies in the package. The inefficiency of repeating src files is
// expected to be minor because py_library is a very thin implementation. An
// alternative would be to disallow cycles and not repeat src files and depend
// on the rules for the modules, with the possible exception for __init__.py
// where cycles are common.
//
// Each module also depends on the py_library rule for the package __init__.py,
// and the package __init__.py depends on its parent package __init__.py.
//
// REQUIRES: No cyclical dependencies across packages.
//
// REQUIRES: This is not a split namespace package.
//
// REQUIRES: PYTHONSAFEPATH is enabled.
func (l *Language) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	config := args.Config.Exts[languageName].(Configuration)
	pkgPath := config.PythonPackagePath
	if strings.HasPrefix(pkgPath, "..") {
		// We have not yet reached a Python workspace.
		return language.GenerateResult{}
	}
	// Compute the value for the "imports" attribute.
	relRoot, err := filepath.Rel(filepath.FromSlash(args.Rel), filepath.FromSlash(config.RootDir))
	if err != nil {
		log.Panicf("computing relRoot with %q and %q", args.Rel, config.RootDir)
	}
	relRoot = filepath.ToSlash(relRoot)

	var filenames []string
	filenames = append(filenames, args.RegularFiles...)
	filenames = append(filenames, args.GenFiles...)
	modules := analyzePythonPackage(pkgPath, args.Dir, args.Subdirs, filenames)

	// Generate a rule for each .py module in this package.
	ruleNames := make(map[string]struct{})
	res := language.GenerateResult{
		Gen:     make([]*rule.Rule, len(modules)),
		Imports: make([]interface{}, len(modules)),
	}
	for i, module := range modules {
		rule := module.GenerateRule(config.NameTemplate, relRoot)
		ruleNames[rule.Name()] = struct{}{}
		res.Gen[i] = rule
		res.Imports[i] = module
	}

	// Check if any rules need to be deleted.
	if args.File != nil {
		for _, rule := range args.File.Rules {
			if _, ok := ruleNames[rule.Name()]; ok {
				// Will be merged with generated rules.
				continue
			}
			if !isRuleManaged(rule) {
				continue
			}
			rule.DelAttr("srcs")
			res.Empty = append(res.Empty, rule)
		}
	}

	return res
}

// Fix implements language.Language.
func (l *Language) Fix(c *config.Config, f *rule.File) {
	// Nothing to fix.
}

func isRuleManaged(rule *rule.Rule) bool {
	for _, tag := range rule.AttrStrings("tags") {
		if tag == tagGazelleManaged {
			return true
		}
	}
	return false
}
