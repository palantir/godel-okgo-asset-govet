package fail

fail

/*
This is a non-compiling file that has been added to explicitly ensure that CI fails.
It also contains the command that caused the failure and its output.
Remove this file if debugging locally.

go mod operation failed. This may mean that there are legitimate dependency issues with the "go.mod" definition in the repository and the updates performed by the gomod check. This branch can be cloned locally to debug the issue.

Command that caused error:
./godelw check compiles

Output:
Running compiles...
-: This application uses version go1.24 of the source-processing packages but runs version go1.25 of 'go list'. It may fail to process source files that rely on newer language features. If so, rebuild the application using a newer version of Go.
Finished compiles
Check(s) produced output: [compiles]

*/
