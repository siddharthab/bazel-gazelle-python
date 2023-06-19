"""Rules for Python module manifest."""

def _modules_manifest_impl(ctx):
    manifest = ctx.outputs.manifest
    manifest_dst = ctx.attr.manifest_dst
    if ctx.label.package:
        manifest_dst = ctx.label.package + "/" + manifest_dst

    args = ctx.actions.args()
    args.add("--output-path", manifest.path)
    args.add("--excluded-patterns", ",".join(ctx.attr.exclude_patterns))
    args.add_all([whl.path for whl in ctx.files.wheels])
    ctx.actions.run(
        inputs = ctx.files.wheels,
        outputs = [manifest],
        executable = ctx.executable._generator,
        arguments = [args],
        use_default_shell_env = False,
    )

    ctx.actions.write(
        output = ctx.outputs.updater,
        is_executable = True,
        content = """#!/bin/bash
set -euo pipefail
dst="${{BUILD_WORKSPACE_DIRECTORY}}/{manifest_dst}"
tmp="$(mktemp)"
if [[ -f "${{dst}}" ]]; then
  # Keep user-added comments at the top of the file; this is crude.
  grep '^#' "${{dst}}" >"${{tmp}}"
  grep -v '^#' "{manifest}" >>"${{tmp}}"
else
  cat "{manifest}" >"${{tmp}}"
fi
mv "${{tmp}}" "${{dst}}"
""".format(manifest = manifest.short_path, manifest_dst = manifest_dst),
    )

    ctx.actions.write(
        output = ctx.outputs.differ,
        is_executable = True,
        content = """#!/bin/bash
set -euo pipefail
diff <(grep -v '^#' "{manifest}" || true) <(grep -v '^#' "{manifest_dst}" || true)
""".format(manifest = manifest.short_path, manifest_dst = manifest_dst),
    )

    files = [ctx.outputs.manifest, ctx.outputs.updater, ctx.outputs.differ]
    return [DefaultInfo(
        files = depset(files),
        runfiles = ctx.runfiles(files = files),
    )]

_modules_manifest = rule(
    _modules_manifest_impl,
    attrs = {
        "differ": attr.output(mandatory = True),
        "exclude_patterns": attr.string_list(
            default = ["^_|(\\._)+"],
            doc = "A set of regex patterns to match against each calculated module path. By default, exclude the modules starting with underscores.",
            mandatory = False,
        ),
        "manifest": attr.output(mandatory = True),
        "manifest_dst": attr.string(mandatory = True),
        "updater": attr.output(mandatory = True),
        "wheels": attr.label_list(
            allow_files = True,
            doc = "The list of wheels, usually the 'all_whl_requirements' from @<pip_repository>//:requirements.bzl",
            mandatory = True,
        ),
        "_generator": attr.label(
            cfg = "exec",
            default = "//manifest",
            executable = True,
        ),
    },
    doc = "Creates a TSV file for mapping module names to wheel distribution names.",
)

def python_external_modules_manifest(name, wheels, exclude_patterns, **kwargs):
    """Rules for Python module manifest.

    Generating, updating a checked-in version, and testing the checked-in
    version for a Python module manifest.

    Args:
        name: Name of the manifest generation rule; other rule names are derivatives.
        wheels: List of wheel files for which to generate a manifest.
        exclude_patterns: File path patterns to exclude from within the wheels.
        **kwargs: Other general rule attributes
    """
    _manifest = name + ".internal.tsv"
    _manifest_dst = name + ".tsv"
    _updater = name + ".updater"
    _differ = name + ".differ"

    _modules_manifest(
        name = name,
        wheels = wheels,
        exclude_patterns = exclude_patterns,
        manifest = _manifest,
        manifest_dst = _manifest_dst,
        updater = _updater,
        differ = _differ,
        **kwargs
    )

    native.sh_binary(
        name = name + ".update",
        srcs = [_updater],
        data = [name],
    )

    native.sh_test(
        name = name + ".test",
        srcs = [_differ],
        data = [name, _manifest_dst],
    )
