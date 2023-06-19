Tests have the following characteristics:

- basic: root directory, outside of Python package, no BUILD file modifications.
- basic/python: python root directory; not an importable package, but any independent modules should have a rule.
- basic/python/pkg1: package initialization and 2 modules that have various interdependencies.
- basic/python/pkg1/subpkg1: subpackage without its own initialization; should depend on parent initialization.
- basic/python/pkg2: no package initialization; depends on pkg1.