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
	"encoding/csv"
	"flag"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

// Assumes file names inside wheels do not have commas.
var (
	outputPath       = flag.String("output-path", "", "Output file path; if empty, will output to stdout")
	excludedPatterns = flag.String("excluded-patterns", "", "File patterns to exclude (comma-separated)")
)

// Analyze the given wheels (paths taken as command args) and output a TSV (on
// stdout) of distribution name, pkg path (dot separated), module name and
// module type (py or so) in the installation. It does so without unzipping the
// wheels so should be very fast (<1s for ~100 wheels).
func main() {
	flag.Parse()
	excludedRegex := compilePatterns(strings.Split(*excludedPatterns, ","))
	wheelPaths := flag.Args()

	var manifest []manifestEntry
	for _, path := range wheelPaths {
		entries, err := analyzeWheel(path, excludedRegex)
		if err != nil {
			log.Fatalf("collecting manifest entries for %q: %v", path, err)
		}
		manifest = append(manifest, entries...)
	}
	sortManifest(manifest)
	writeManifest(manifest, *outputPath)
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	var res []*regexp.Regexp
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			log.Fatalf("compiling regular expression %q: %v", pattern, err)
		}
		res = append(res, compiled)
	}
	return res
}

func sortManifest(manifest []manifestEntry) {
	sort.Slice(manifest, func(i, j int) bool {
		if a, b := manifest[i].DistName, manifest[j].DistName; a != b {
			return a < b
		}
		if a, b := manifest[i].Pkg, manifest[j].Pkg; a != b {
			return a < b
		}
		if a, b := manifest[i].Module, manifest[j].Module; a != b {
			return a < b
		}
		if a, b := manifest[i].Type, manifest[j].Type; a != b {
			return a < b
		}
		log.Panicf("bug: duplicate manifest entries: %v, %v", manifest[i], manifest[j])
		return false
	})
}

func writeManifest(manifest []manifestEntry, path string) {
	f := os.Stdout
	if path != "" {
		var err error
		f, err = os.Create(path)
		if err != nil {
			log.Fatalf("could not create output file %q", path)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatalf("could not close file %q", path)
			}
		}()
	}
	_, err := f.WriteString("# GENERATED FILE - DO NOT EDIT!\n")
	if err != nil {
		log.Fatalf("could not write to output: %v", err)
	}
	w := csv.NewWriter(f)
	w.Comma = '\t'
	for _, entry := range manifest {
		if err := w.Write([]string{entry.DistName, entry.Pkg, entry.Module, entry.Type}); err != nil {
			log.Fatalf("writing manifest record: %v", err)
		}
	}
	w.Flush()
}
