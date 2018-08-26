# Mod 

Command line tool to upgrade/downgrade Semantic Import Versioning in Go Modules

### Usage

Run the command from any direcory that has go.mod file as such: 

`mod upgrade` OR `mod downgrade` 

The tool will update the major version in the go.mod file and it will 
traverse all the Go Files and upgrades/downgrades any imports as well. 

You can run this command inside the example folder 
and notice how the import paths and the module name alike get updated.
