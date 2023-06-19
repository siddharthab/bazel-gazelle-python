# Bazel Gazelle Python Extension

Gazelle is an extensible BUILD file generator for Bazel. This repository extends
Gazelle for generating Bazel rules for Python code.

This is currently experimental and is meant as an alternative to the
[pseudo-official option](https://github.com/bazelbuild/rules_python/tree/main/gazelle).

## Salient Differences

1. Written in pure Go with more idiomatic Go code. Consequently:
   1. Uses [go-python/gpython](https://github.com/go-python/gpython) to parse
      Python files, which follows the specs of Python 3.4.
   2. Assumes a fixed list of root packages available with the interpreter, i.e. stdlib and other system installed packages. The list can be overridden through command line flags and directives.
2. Finer grained dependencies where modules are the build units not packages,
   this allows for better test caching and has better fidelity to Python build tooling.
3. Simpler infrastructure. Try to keep the focus on the logic in
   [python/module.go](python/module.go) and [python/analyzer.go](python/analyzer.go).

## Wishlist

1. Collapse cyclical dependencies, at least within a single package, to a single
   `py_library` rule. Currently, the sources and their direct deps are transferred
   up the chain. This should be doable with a little more work.
2. Include/exclude dependencies for specific rules through directives.
3. Design a better structured format for inputs, like for external modules map
   and internal modules list.
4. Parse global symbols out of a Python module without loading the module to
   improve the accuracy of resolution of IMPORT-FROM statemetns. This may be
   impossible or unreliable.
5. For binary distribution wheels, extract imports from compiled modules (.so
   files). This may be impossible or unreliable.

## Usage

Currently meant for usage by advanced users only. See
[configuration.go](/python/configuration.go) for command line flags and
directives.
