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

// Package parser parses Python files and returns relevant information to the
// Gazelle Python extension.
//
// It is currently using a Python AST parser written in Go, which implements the
// spec from Python 3.4. If this is insufficient, this package will need to be
// rewritten to farm out the processing to a pool of running Python processes.
package parser

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
)

const debugParse = false

type Result struct {
	Imports          []string
	HasMainNameCheck bool
}

// ParsePath parses a Python module at the given path.
func ParsePath(path string) (Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return Result{}, fmt.Errorf("opening Python file: %q: %w", path, err)
	}
	defer f.Close()
	res, err := Parse(f, path)
	if err != nil {
		return res, fmt.Errorf("parsing Python file: %q: %w", path, err)
	}
	return res, nil
}

// Parse parses a Python module read by the reader.
func Parse(r io.Reader, filename string) (Result, error) {
	res := Result{}
	tree, err := parser.Parse(r, filename, py.ExecMode)
	if err != nil {
		return Result{}, err
	}
	if debugParse {
		println(ast.Dump(tree))
	}
	importedSet := make(map[string]struct{})
	ast.Walk(tree, func(node ast.Ast) bool { return visitor(node, importedSet, &res) })
	for name := range importedSet {
		res.Imports = append(res.Imports, name)
	}
	sort.Strings(res.Imports)
	return res, nil
}

// Checks for import statements and the `if __name__ == "__main__"` block.
func visitor(tree ast.Ast, importedSet map[string]struct{}, res *Result) bool {
	stmt, ok := tree.(ast.Stmt)
	if !ok {
		// Let's be simple and try continuing the walk in all cases.
		return true
	}
	switch stmt := stmt.(type) {
	case *ast.Import:
		for _, alias := range stmt.Names {
			importedSet[string(alias.Name)] = struct{}{}
		}
	case *ast.ImportFrom:
		for _, alias := range stmt.Names {
			importedSet[string(stmt.Module)+"."+string(alias.Name)] = struct{}{}
		}
	case *ast.If:
		if got, safeBody := isTypeCheckingConditional(importedSet, stmt); got {
			for _, stmt := range safeBody {
				ast.Walk(stmt, func(tree ast.Ast) bool { return visitor(tree, importedSet, res) })
			}
			return false
		}
		if !res.HasMainNameCheck && isMainNameCheck(stmt) {
			res.HasMainNameCheck = true
		}
	}
	return true
}

// If this is a typing.TYPE_CHECKING conditional, then do not discard the
// positive branch and return the other.
//
// NOTE: This is a little crude, but should work for most cases. If it does not,
// improve the logic as needed.
func isTypeCheckingConditional(imported map[string]struct{}, stmt *ast.If) (res bool, safeBody []ast.Stmt) {
	switch test := stmt.Test.(type) {
	case *ast.UnaryOp:
		if test.Op == ast.Not {
			return isTypeCheckingConditional(imported, &ast.If{Test: test.Operand, Body: stmt.Orelse, Orelse: stmt.Body})
		}
	case *ast.Attribute:
		if _, ok := imported["typing"]; !ok {
			return false, nil
		}
		if name, ok := test.Value.(*ast.Name); ok && name.Id == "typing" && test.Attr == "TYPE_CHECKING" {
			return true, stmt.Orelse
		}
	case *ast.Name:
		if _, ok := imported["typing.TYPE_CHECKING"]; !ok {
			return false, nil
		}
		if test.Id == "TYPE_CHECKING" {
			return true, stmt.Orelse
		}
	}
	return false, nil
}

// Check for `if __name__ == "__main__":` blocks.
func isMainNameCheck(stmt *ast.If) bool {
	v, ok := stmt.Test.(*ast.Compare)
	if !ok || len(v.Ops) != 1 || v.Ops[0] != ast.Eq || len(v.Comparators) != 1 {
		return false
	}
	if left, ok := v.Left.(*ast.Name); !ok || left.Id != "__name__" {
		return false
	}
	if right, ok := v.Comparators[0].(*ast.Str); !ok || right.S != "__main__" {
		return false
	}
	return true
}
