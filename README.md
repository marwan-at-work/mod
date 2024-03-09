# Mod 

Command line tool to upgrade/downgrade Semantic Import Versioning in Go Modules

## Motivation 

There are two good use cases to do this: 

1. If you own a library and you want to introduce a breaking change, then you have to go around all your Go files and sub-packages and update the import paths to include v2, v3, etc. This tool just does it automatically with one command. 

2. If you own a library that is already tagged v2 or above but is incompatible with Semantic Import Versioning, then 
this tool can solve the problem for you with one command as well. Introduce a go.mod file with the correct import path, and just run `mod upgrade` once or `mod -t=X upgrade` (where x is the latest tag major) to update the import paths of your go files to match whatever tag you're at.

## Install

`go install github.com/marwan-at-work/mod/cmd/mod@latest`

## Usage

### upgrade/downgrade

`mod upgrade` OR `mod downgrade` from any directory that has go.mod file.


The tool will update the major version in the go.mod file and it will 
traverse all the Go Files and upgrades/downgrades any imports as well. 

For example, if you have the following package: 

```
// go.mod

module github.com/me/example

require github.com/pkg/errors v0.8.0
```

that imports a sub directory in its Go files:

```golang
// example.go

package example

import (
    "github.com/me/example/subdir"
)

// rest of file...
```

You can run `mod upgrade` from the root of that directory and the above two files will be rewritten as such: 

```
module github.com/me/example/v2

require github.com/pkg/errors
```

```golang
package example

import (
    "github.come/me/example/v2/subdir"
)

// rest of file...
```

You can of course, downgrade again or upgrade more incrementally. 

You can also run this command inside the example folder 
and notice how the import paths and the module name alike get updated.

#### Upgrading a dependency

Say you are using a dependency like `github.com/x/y/v2` and you want to use the new major version they just released. You don't want to change your own import paths, but you want to change the import paths of a dependency you're using from `v2` to `v3`. 

All you have to do is 

```
mod upgrade --mod-name=github.com/x/y
```

#### Downgrading a dependency

Say you are using a dependency like `github.com/x/y/v3` and you want to use the previous major version. You don't want to change your own import paths, but you want to change the import paths of a dependency you're using from `v3` to `v2`.

All you have to do is

```
mod downgrade --mod-name=github.com/x/y
```

### migrate

### replace

If the repos location changed (renamed, moving to vanity urls, etc.) you have to update all application depending on 
that repo. Meaning you have to search and replace all occurrences of the old location by a new one.

This is what this `replace` command does for you.

```
mod replace --mod-old=github.com/x/y --mod-new=something.com/x/y
```

## Status

Works as intended. Feel free to report any issues.

### Troubleshooting

If you are using iTerm2 and see this error: 

```
<html><head></head><body><p><em>Error:</em> Your request has found no candidate provider [hs="(null)", id="(null)"]</p></body</html>
```

Make sure you close iTerm2 and start again. See https://github.com/marwan-at-work/mod/issues/3 for details. 
