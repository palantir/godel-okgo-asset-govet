package fail

fail

/*
This is a non-compiling file that has been added to explicitly ensure that CI fails.
It also contains the command that caused the failure and its output.
Remove this file if debugging locally.

go mod operation failed. This may mean that there are legitimate dependency issues with the "go.mod" definition in the repository and the updates performed by the gomod check. This branch can be cloned locally to debug the issue.

Command that caused error:
./godelw lint compiles

Output:
govet/creator/creator.go:1: : # github.com/palantir/godel-okgo-asset-govet/govet/creator
govet/creator/creator.go:28:11: cannot use &govet.Checker{} (value of type *govet.Checker) as okgo.Checker value in return statement: *govet.Checker does not implement okgo.Checker (missing method MultiCPU) (typecheck)
// Copyright 2016 Palantir Technologies, Inc.
1 issues:
* typecheck: 1

*/
