# mod_test

* * *

`mod_test` is a tool that creates files, directories and go modules for testing commands. useful for
development.

### *details*

running

    go test [-v] [-name DIR_NAME] [-cleanup]

creates a directory, by default, called `test_module`.

the following is an example of running
the `fork rewrite` command against the generated go module `test_module`.

if you've installed the `binary`:

    ~$ cd test_module
    ~$ mod fork rewrite github.com/old/test_module

or using `go run`:

    ~$ cd ../mod
    ~$ go run main.go fork rewrite -r ../mod_test/test_module/ github.com/old/test_module

note that in this case, on running the `rewrite` command you may see some errors regarding `imports` and `missing packages`, but that is to be expected as these packages don't actually exist for the `packages` library to work correctly. for any real and existing packages, this _should_ be a non-issue.