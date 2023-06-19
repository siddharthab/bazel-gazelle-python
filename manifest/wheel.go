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

import (
	"archive/zip"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/siddharthab/bazel-gazelle-python/internal"
)

// Crude regular expression that uses the fact that canonicalized names can rely on using '-' as the component separator.
// https://packaging.python.org/en/latest/specifications/binary-distribution-format/#file-name-convention.
var distFilenameRegex = regexp.MustCompile("(?P<distribution>[^-]+)-(?P<version>[^-]+)(-(?P<build_tag>[^-]+))?-(?P<python_tag>[^-]+)-(?P<abi_tag>[^-]+)-(?P<platform_tag>[^-]+).whl")

type manifestEntry struct {
	DistName    string
	Pkg, Module string
	Type        string
}

func analyzeWheel(wheelPath string, excludedPatterns []*regexp.Regexp) ([]manifestEntry, error) {
	distName, distVersion, err := parseWheelName(path.Base(wheelPath))
	if err != nil {
		return nil, fmt.Errorf("analyzing wheel path %q: %w", wheelPath, err)
	}
	distInfoDir := fmt.Sprintf("%s-%s.dist-info", distName, distVersion)
	dataDir := fmt.Sprintf("%s-%s.data", distName, distVersion)

	files, err := listFilesInZip(wheelPath, excludedPatterns)
	if err != nil {
		return nil, err
	}

	entryMap := make(map[string]manifestEntry)
	for _, name := range files {
		if pkg, module, typ, importSpec := moduleForFilename(name, distInfoDir, dataDir); pkg != "" || module != "" {
			entryMap[importSpec] = manifestEntry{distName, pkg, module, typ}
		}
	}

	var manifestEntries []manifestEntry
	for _, entry := range entryMap {
		manifestEntries = append(manifestEntries, entry)
	}
	return manifestEntries, nil
}

func parseWheelName(wheelName string) (distribution, version string, err error) {
	components := distFilenameRegex.FindStringSubmatch(wheelName)
	if components == nil {
		return "", "", fmt.Errorf("wheel name (%s) does not follow format %q", wheelName, distFilenameRegex.String())
	}
	return components[1], components[2], nil
}

func listFilesInZip(path string, excludedPatterns []*regexp.Regexp) ([]string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("opening zip file at %q: %w", path, err)
	}
	defer r.Close()
	var res []string
	for _, f := range r.File {
		if matchesAnyPattern(f.Name, excludedPatterns) {
			continue
		}
		res = append(res, f.Name)
	}
	return res, nil
}

func matchesAnyPattern(s string, pats []*regexp.Regexp) bool {
	for _, pat := range pats {
		if pat.MatchString(s) {
			return true
		}
	}
	return false
}

// Returns the package path (dot separated), module name and module type (py
// or so) for the module given by the file name inside a wheel distribution.
// Returns empty strings if the filename does not correspond to a Python module.
// https://packaging.python.org/en/latest/specifications/binary-distribution-format/#file-contents
func moduleForFilename(name, distInfoDir, dataDir string) (pkg, module, typ, importSpec string) {
	const pathSep = "/" // Separator for file name in zip will always be '/'.

	components := strings.Split(name, pathSep)
	switch strings.ToLower(components[0]) {
	case distInfoDir:
		return "", "", "", ""
	case dataDir:
		second := strings.ToLower(components[1])
		if second != "purelib" && second != "platlib" {
			return "", "", "", ""
		}
		components = components[2:]
	}

	pkgPath := path.Join(components[:len(components)-1]...)
	pkg = strings.Join(components[:len(components)-1], ".")
	module, typ, ok := internal.ModuleName(components[len(components)-1])
	if ok {
		importSpec = internal.ImportSpec(pkgPath, module)
		return pkg, module, typ, importSpec
	}
	return "", "", "", ""
}
