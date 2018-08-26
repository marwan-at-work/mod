# Mod 

Command line tool to upgrade/downgrade Semantic Import Versioning in Go Modules

### Install

`go get github.com/marwan-at-work/mod`

### Usage

`mod upgrade` OR `mod downgrade` from any direcory that has go.mod file.


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

## Status

WIP 