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
	"os"
	"path/filepath"
	"sort"

	"github.com/siddharthab/bazel-gazelle-python/internal"
	"github.com/siddharthab/bazel-gazelle-python/python/parser"
)

// Analyzes the Python package with slash separated path given by pkgPath,
// located at absPath dir in the system with the given subDirs, and comprised of
// files given by filenames. Returns a sorted list of Python modules.
func analyzePythonPackage(pkgPath, absPath string, subDirs, filenames []string) []*Module {
	// TODO: Parallelize this if this is slow.
	var (
		importSpecs []string
		moduleMap   = make(map[string]*Module) // Keyed by import specifier.
	)
	for _, filename := range filenames {
		moduleName, moduleType, ok := internal.ModuleName(filename)
		if !ok || moduleType != "py" {
			continue
		}
		path := filepath.Join(absPath, filename)
		res, err := parser.ParsePath(path)
		if err != nil {
			log.Printf("unable to generate rule for Python module: %v", err)
			continue
		}
		importSpec := internal.ImportSpec(pkgPath, moduleName)
		importSpecs = append(importSpecs, importSpec)
		moduleMap[importSpec] = &Module{
			Result:     res,
			ImportSpec: importSpec,
			PkgPath:    pkgPath,
			Name:       moduleName,
			Filename:   filename,
			InPkgDeps:  make(map[*Module]struct{}),
		}
	}
	subPackages := make(map[string]struct{})
	for _, subdir := range subDirs {
		if _, err := os.Stat(filepath.Join(absPath, subdir, "__init__.py")); err == nil {
			subPackages[subdir] = struct{}{}
		}
	}

	// Compute transitive closures of dependencies within the package.
	for _, module := range moduleMap {
		module.ProcessImports(moduleMap, subPackages)
	}
	for _, module := range moduleMap {
		module.DepsClosure()
	}

	// Sort the modules by name and return.
	sort.Strings(importSpecs)
	var res []*Module
	for _, spec := range importSpecs {
		res = append(res, moduleMap[spec])
	}
	return res
}
