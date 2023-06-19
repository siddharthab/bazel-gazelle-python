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

import (
	"log"
	"path"
	"strings"
)

// ModuleName removes the extension(s) from a source file to give the module
// name. Returns false if filename is not a valid Python module.
func ModuleName(filename string) (moduleName, typ string, ok bool) {
	ext := path.Ext(filename)
	switch ext {
	case ".py":
		if filename == "__init__.py" {
			return "", "py", true
		}
		return strings.TrimSuffix(filename, ext), "py", true
	case ".so":
		for ext != "" {
			filename = strings.TrimSuffix(filename, ext)
			ext = path.Ext(filename)
		}
		return filename, "so", true
	}
	return "", "", false
}

// MustModuleName is a version of ModuleName which panics if file name can not
// generate a valid Python module name.
func MustModuleName(filename string) (moduleName, typ string) {
	res, typ, ok := ModuleName(filename)
	if !ok {
		log.Panicf("unable to compute Python module name from %q", filename)
	}
	return res, typ
}

// ImportSpec returns the import specifier for a given slash separated package
// path and the module name. At least one of pkg path or module name must not be
// blank.
func ImportSpec(pkgPath, moduleName string) string {
	if pkgPath == "" && moduleName == "" {
		panic("bug")
	}
	var res string
	if pkgPath != "" {
		res = strings.ReplaceAll(pkgPath, "/", ".")
	}
	if moduleName != "" {
		if res != "" {
			res += "."
		}
		res += moduleName
	}
	return res
}
