# Copyright 2023 The Bazel Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"This file managed by `bazel run //:gazelle_update_repos`"

load(
    "@bazel_gazelle//:deps.bzl",
    "go_repository",
)

def gazelle_deps():
    "Fetch go dependencies"
    go_repository(
        name = "com_github_bazelbuild_bazel_gazelle",
        importpath = "github.com/bazelbuild/bazel-gazelle",
        sum = "h1:kRymV9q+24Mbeg25fJehw+gvrtVIlwZZAefOSUq4MzU=",
        version = "v0.20.0",
    )
    go_repository(
        name = "com_github_bazelbuild_buildtools",
        importpath = "github.com/bazelbuild/buildtools",
        sum = "h1:OfyUN/Msd8yqJww6deQ9vayJWw+Jrbe6Qp9giv51QQI=",
        version = "v0.0.0-20190731111112-f720930ceb60",
    )
    go_repository(
        name = "com_github_bazelbuild_rules_go",
        importpath = "github.com/bazelbuild/rules_go",
        sum = "h1:wzbawlkLtl2ze9w/312NHZ84c7kpUCtlkD8HgFY27sw=",
        version = "v0.0.0-20190719190356-6dae44dc5cab",
    )
    go_repository(
        name = "com_github_burntsushi_toml",
        importpath = "github.com/BurntSushi/toml",
        sum = "h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_davecgh_go_spew",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_fsnotify_fsnotify",
        importpath = "github.com/fsnotify/fsnotify",
        sum = "h1:IXs+QLmnXW2CcXuY+8Mzv/fWEsPGWxqefPtCP5CnV9I=",
        version = "v1.4.7",
    )
    go_repository(
        name = "com_github_go_python_gpython",
        importpath = "github.com/go-python/gpython",
        sum = "h1:3ZbFGiuP384IpRm2GefQ2C2QtVwF3/J2nvhrEfQG3uk=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_go_cmp",
        importpath = "github.com/google/go-cmp",
        sum = "h1:O2Tfq5qg4qc4AmwVlvv0oLiVAGB7enBSJ2x2DqQFi38=",
        version = "v0.5.9",
    )
    go_repository(
        name = "com_github_gopherjs_gopherjs",
        importpath = "github.com/gopherjs/gopherjs",
        sum = "h1:16eHWuMGvCjSfgRJKqIzapE78onvvTbdi1rMkU00lZw=",
        version = "v0.0.0-20180825215210-0210a2f0f73c",
    )
    go_repository(
        name = "com_github_gopherjs_gopherwasm",
        importpath = "github.com/gopherjs/gopherwasm",
        sum = "h1:fA2uLoctU5+T3OhOn2vYP0DVT6pxc7xhTlBB1paATqQ=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_kr_pretty",
        importpath = "github.com/kr/pretty",
        sum = "h1:L/CwN0zerZDmRFUapSPitk6f+Q3+0za1rQkzVuMiMFI=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_kr_pty",
        importpath = "github.com/kr/pty",
        sum = "h1:VkoXIwSboBpnk99O/KFauAEILuNHv5DVFKZMBN/gUgw=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_kr_text",
        importpath = "github.com/kr/text",
        sum = "h1:45sCR5RtlFHMR4UwH9sdQ5TC8v0qDQCHnXt+kaKSTVE=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_mattn_go_runewidth",
        importpath = "github.com/mattn/go-runewidth",
        sum = "h1:a+kO+98RDGEfo6asOGMmpodZq4FNtnGP54yps8BzLR4=",
        version = "v0.0.3",
    )
    go_repository(
        name = "com_github_pelletier_go_toml",
        importpath = "github.com/pelletier/go-toml",
        sum = "h1:T5zMGML61Wp+FlcbWjRDT7yAxhJNAiPPLOFECq181zc=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_peterh_liner",
        importpath = "github.com/peterh/liner",
        sum = "h1:aJ4AOodmL+JxOZZEL2u9iJf8omNRpqHc/EbrK+3mAXw=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_pmezard_go_difflib",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "in_gopkg_check_v1",
        importpath = "gopkg.in/check.v1",
        sum = "h1:qIbj1fsPNlZgppZ+VLlY7N33q108Sa+fhmuc+sWQYwY=",
        version = "v1.0.0-20180628173108-788fd7840127",
    )
    go_repository(
        name = "in_gopkg_yaml_v2",
        importpath = "gopkg.in/yaml.v2",
        sum = "h1:ZCJp+EgiOT7lHqUV2J862kp8Qj64Jo6az82+3Td9dZw=",
        version = "v2.2.2",
    )
    go_repository(
        name = "org_golang_x_sync",
        importpath = "golang.org/x/sync",
        sum = "h1:vcxGaoTs7kV8m5Np9uUNQin4BrLOthgV7252N8V+FwY=",
        version = "v0.0.0-20190911185100-cd5d95a43a6e",
    )
    go_repository(
        name = "org_golang_x_sys",
        importpath = "golang.org/x/sys",
        sum = "h1:kwrAHlwJ0DUBZwQ238v+Uod/3eZ8B2K5rYsUHBQvzmI=",
        version = "v0.0.0-20211117180635-dee7805ff2e1",
    )
    go_repository(
        name = "org_golang_x_tools",
        importpath = "golang.org/x/tools",
        sum = "h1:FkAkwuYWQw+IArrnmhGlisKHQF4MsZ2Nu/fX4ttW55o=",
        version = "v0.0.0-20190122202912-9c309ee22fab",
    )
