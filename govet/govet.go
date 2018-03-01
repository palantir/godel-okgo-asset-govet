// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package govet

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"

	"github.com/palantir/okgo/checker"
	"github.com/palantir/okgo/okgo"
	"github.com/pkg/errors"
)

const (
	TypeName okgo.CheckerType     = "govetCheck"
	Priority okgo.CheckerPriority = 0
)

func Creator() checker.Creator {
	return checker.NewCreator(
		TypeName,
		Priority,
		func(cfgYML []byte) (okgo.Checker, error) {
			return &govetCheck{}, nil
		},
	)
}

type govetCheck struct{}

func (c *govetCheck) Type() (okgo.CheckerType, error) {
	return TypeName, nil
}

func (c *govetCheck) Priority() (okgo.CheckerPriority, error) {
	return Priority, nil
}

var lineRegexp = regexp.MustCompile(`(.+):(\d+): (.+)`)

func (c *govetCheck) Check(pkgPaths []string, projectDir string, stdout io.Writer) {
	wd, err := os.Getwd()
	if err != nil {
		okgo.WriteErrorAsIssue(errors.Wrapf(err, "failed to determine working directory"), stdout)
		return
	}
	cmd := exec.Command("go", append(
		[]string{"vet"},
		pkgPaths...)...,
	)
	checker.RunCommandAndStreamOutput(cmd, func(line string) okgo.Issue {
		// if govetCheck finds issue, it ends with the output "exit status 1", but we don't want to include it as part of the output
		if line == "exit status 1" {
			return okgo.Issue{}
		}
		if match := lineRegexp.FindStringSubmatch(line); match != nil {
			// govetCheck does not include column info, so add it in manually
			line = fmt.Sprintf("%s:%s:0: %s", match[1], match[2], match[3])
		}
		return okgo.NewIssueFromLine(line, wd)
	}, stdout)
}

func (c *govetCheck) RunCheckCmd(args []string, stdout io.Writer) {
	checker.AmalgomatedRunRawCheck(string(TypeName), args, stdout)
}
