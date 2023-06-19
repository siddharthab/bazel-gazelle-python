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
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/siddharthab/bazel-gazelle-python/internal"
	yaml "gopkg.in/yaml.v2"
)

const languageName = "py"

// Accepted directives for this extension.
const (
	directiveExtension              = "py_extension"
	directiveRoot                   = "py_root_dir"
	directiveInternalModuleListPath = "py_internal_module_list_path"
	directiveExternalModuleMapPath  = "py_external_module_map_path"
	directiveExternalRepoNamePrefix = "py_external_repo_name_prefix"
	directiveNameTemplate           = "py_name_template"
)

var directiveKeys = []string{directiveExtension, directiveRoot, directiveInternalModuleListPath, directiveExternalModuleMapPath, directiveExternalRepoNamePrefix, directiveNameTemplate}

// ExternalModule is a Python module available from an external distribution.
type ExternalModule struct {
	Dist        string // Distribution name.
	PkgPath     string // A slash separated path to the package.
	Module      string // Name of the module, can be blank for __init__.py (but not when PkgPath is also blank).
	BazelTarget string // Bazel target for this import.
	Type        string // py or so (currently not relevant).
}

// Configuration is configuration for the Python language extension. A default
// configuration is set through command line flags and their default values.
// Each directory gets its own copy and the values may be changed by
// Configurer.Configure().
type Configuration struct {
	// Enable the Python Gazelle language extension.
	Enable bool
	// Root directory for Python code.
	RootDir string
	// Python package path (slash separated); its value is Bazel package path
	// relative to RootDir. If this starts with '..', then it means we have not
	// yet reached the Python workspace.
	PythonPackagePath string
	// List of modules internal to the interpreter, whether part of stdlib, or
	// available as a system installation. A common default for such a list
	// would be given by
	// `python3 -c "import sys; print(sorted(sys.stdlib_module_names))"`.
	InternalModuleList map[string]struct{}
	// Path to list (one per line; comment char '#') of internal modules.
	// Functions as a caching key for InternalModuleList.
	InternalModuleListPath string
	// Map of import specifiers for external module names to their sources.
	ExternalModuleMap map[string]ExternalModule
	// Path to map of external modules from where ExternalModuleMap is
	// read. Functions as a caching key for ExternalModuleMap.
	ExternalModuleMapPath string
	// Name prefix under which the external repositories are defined, e.g. "pip_".
	ExternalRepoNamePrefix string
	// Name template to use for naming targets.
	NameTemplate string
}

// Configurer manages the configuration at root and for each subdirectory.
type Configurer struct {
	// Initial copy of the configuration, before it is copied as an extension configuration to Gazelle.
	initial Configuration
}

var _ config.Configurer = &Configurer{}

// RegisterFlags implements config.Configurer.
//
// It registers the flags to their relevant fields in an initial copy of the
// Configuration object. The configuration extension is not set yet.
func (pc *Configurer) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
	fs.BoolVar(&pc.initial.Enable, "py-extension", true, "Enable Python language extension.")
	fs.StringVar(&pc.initial.RootDir, "py-root-dir", "", "Root directory for Python code.")
	fs.StringVar(&pc.initial.InternalModuleListPath, "py-internal-modules-path", "", "Path to manifest of external modules.")
	fs.StringVar(&pc.initial.ExternalModuleMapPath, "py-external-modules-path", "", "Path to manifest of external modules.")
	fs.StringVar(&pc.initial.ExternalRepoNamePrefix, "py-external-repo-name-prefix", "", "Name prefix under which the external repositories are defined.")
	fs.StringVar(&pc.initial.NameTemplate, "py-name-template", "{module_name}", "Name prefix under which the external repositories are defined.")
}

// CheckFlags implements config.Configurer.
//
// It performs any resource initialization based on the flags, validates the
// initial configuration, and sets the configuration extension.
func (pc *Configurer) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	var err error
	var config Configuration
	config, pc.initial = pc.initial, Configuration{} // Swap out the value in the configurer.
	if config.InternalModuleListPath != "" {
		config.InternalModuleList, err = readInternalModuleListPath(filepath.Join(c.RepoRoot, config.InternalModuleListPath))
		if err != nil {
			return err
		}
	}
	if config.ExternalModuleMapPath != "" {
		config.ExternalModuleMap, err = readExternalModuleMapPath(filepath.Join(c.RepoRoot, config.ExternalModuleMapPath), pc.initial.ExternalRepoNamePrefix)
		if err != nil {
			return err
		}
	}
	c.Exts[languageName] = config
	return nil
}

// KnownDirectives implements config.Configurer.
func (pc Configurer) KnownDirectives() []string {
	return directiveKeys
}

// Configure implements config.Configurer.
func (pc Configurer) Configure(c *config.Config, rel string, f *rule.File) {
	config := c.Exts[languageName].(Configuration)
	var directives []rule.Directive
	if f != nil {
		directives = f.Directives
	}

	var err error
	var readInternalModuleList, readExternalModuleMap bool
	for _, d := range directives {
		switch d.Key {
		case directiveExtension:
			config.Enable, err = strconv.ParseBool(d.Value)
			if err != nil {
				log.Fatalf("invalid directive value %q for %q in %q: %v", d.Value, d.Key, rel, err)
			}
		case directiveRoot:
			config.RootDir = path.Join(rel, d.Value)
		case directiveInternalModuleListPath:
			if config.InternalModuleListPath != d.Value {
				if config.InternalModuleListPath != "" {
					readInternalModuleList = true
				} else {
					config.InternalModuleList = nil
				}
			}
			config.InternalModuleListPath = d.Value
		case directiveExternalModuleMapPath:
			if config.ExternalModuleMapPath != d.Value {
				if config.ExternalModuleMapPath != "" {
					readExternalModuleMap = true
				} else {
					config.ExternalModuleMap = make(map[string]ExternalModule)
				}
			}
			config.ExternalModuleMapPath = d.Value
		case directiveExternalRepoNamePrefix:
			config.ExternalRepoNamePrefix = d.Value
		case directiveNameTemplate:
			config.NameTemplate = d.Value
		}
	}
	if readInternalModuleList {
		config.InternalModuleList, err = readInternalModuleListPath(filepath.Join(c.RepoRoot, config.InternalModuleListPath))
		if err != nil {
			log.Fatal(err)
		}
	}
	if readExternalModuleMap {
		config.ExternalModuleMap, err = readExternalModuleMapPath(filepath.Join(c.RepoRoot, config.ExternalModuleMapPath), config.ExternalRepoNamePrefix)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Compute the Python package path for this directory.
	rootRel, err := filepath.Rel(filepath.FromSlash(config.RootDir), filepath.FromSlash(rel))
	if err != nil {
		log.Fatalf("computing Python package path from Bazel package %q and Python root dir %q", rel, config.RootDir)
	}
	rootRel = filepath.ToSlash(rootRel)
	if rootRel == "." {
		rootRel = ""
	}
	config.PythonPackagePath = rootRel
	c.Exts[languageName] = config
}

func readInternalModuleListPath(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening Python internal module list: %w", err)
	}
	defer f.Close()

	res, err := readInternalModuleList(f)
	if err != nil {
		return nil, fmt.Errorf("parsing Python internal module list at path %q: %w", path, err)
	}
	return res, nil
}

func readInternalModuleList(r io.Reader) (map[string]struct{}, error) {
	res := make(map[string]struct{})
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		res[strings.TrimSpace(line)] = struct{}{}
	}
	return res, scanner.Err()
}

func readExternalModuleMapPath(path, namePrefix string) (map[string]ExternalModule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening Python external module map: %w", err)
	}
	defer f.Close()

	readerFn := readExternalModuleMapTSV
	if ext := filepath.Ext(path); ext == ".yaml" || ext == ".yml" {
		readerFn = readExternalModuleMapYaml
	}
	res, err := readerFn(f, namePrefix)
	if err != nil {
		return nil, fmt.Errorf("parsing Python external module manifest at path %q: %v", path, err)
	}
	return res, nil
}

func readExternalModuleMapTSV(r io.Reader, namePrefix string) (map[string]ExternalModule, error) {
	csvR := csv.NewReader(r)
	csvR.Comma = '\t'
	csvR.Comment = '#'
	allRecords, err := csvR.ReadAll()
	if err != nil {
		return nil, err
	}

	res := make(map[string]ExternalModule)
	for _, records := range allRecords {
		// We currently don't care if the module is .py or .so, but might in the
		// future if we figure out how to get fine grained deps from .so
		// modules, and can then create fine-grained py_library rules in
		// installed distributions.
		dist, pkg, moduleName, typ := records[0], records[1], records[2], records[3]
		importSpec := internal.ImportSpec(pkg, moduleName)
		if val, exists := res[importSpec]; exists {
			if val.Type == typ {
				return nil, fmt.Errorf("duplicate entries in Python external module manifest for %q: %v and %v", importSpec, val.Dist, dist)
			} else {
				continue
			}
		}
		module := ExternalModule{
			Dist:        dist,
			PkgPath:     strings.ReplaceAll(pkg, ".", "/"),
			Module:      moduleName,
			BazelTarget: fmt.Sprintf("@%s%s//:pkg", namePrefix, dist),
			Type:        typ,
		}
		res[importSpec] = module
	}
	return res, nil
}

func readExternalModuleMapYaml(r io.Reader, namePrefix string) (map[string]ExternalModule, error) {
	type manifest struct {
		ModulesMapping map[string]string `yaml:"modules_mapping"`
	}
	type container struct {
		Manifest *manifest `yaml:"manifest"`
	}
	var c container
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to decode yaml manifest file: %w", err)
	}
	res := make(map[string]ExternalModule)
	for k, v := range c.Manifest.ModulesMapping {
		res[k] = ExternalModule{
			Dist: v,
		}
	}
	return res, nil
}
