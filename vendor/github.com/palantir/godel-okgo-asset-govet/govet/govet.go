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
	"strings"

	"github.com/palantir/okgo/checker"
	"github.com/palantir/okgo/okgo"
	"github.com/pkg/errors"
)

const (
	TypeName okgo.CheckerType     = "govet"
	Priority okgo.CheckerPriority = 0
)

type Checker struct{}

func (c *Checker) Type() (okgo.CheckerType, error) {
	return TypeName, nil
}

func (c *Checker) Priority() (okgo.CheckerPriority, error) {
	return Priority, nil
}

var lineRegexp = regexp.MustCompile(`(.+):(\d+): (.+)`)

func (c *Checker) Check(pkgPaths []string, projectDir string, stdout io.Writer) {
	wd, err := os.Getwd()
	if err != nil {
		okgo.WriteErrorAsIssue(errors.Wrapf(err, "failed to determine working directory"), stdout)
		return
	}

	// go vet does not accept package paths that start with "./.." because they are not considered canonical paths. Deal
	// with this specific case manually by converting paths that start with "./.." to start with "..".
	cleanedPaths := make([]string, len(pkgPaths))
	for i, v := range pkgPaths {
		if strings.HasPrefix(v, "./..") {
			v = strings.TrimPrefix(v, "./")
		}
		cleanedPaths[i] = v
	}
	pkgPaths = cleanedPaths

	cmd := exec.Command("go", append(
		[]string{"vet"},
		pkgPaths...)...,
	)
	checker.RunCommandAndStreamOutput(cmd, func(line string) okgo.Issue {
		// if govet finds issue, it ends with the output "exit status 1", but we don't want to include it as part of the output
		if line == "exit status 1" {
			return okgo.Issue{}
		}
		// ignore output in the form of comments
		if strings.HasPrefix(line, "#") {
			return okgo.Issue{}
		}
		if match := lineRegexp.FindStringSubmatch(line); match != nil {
			// Checker does not include column info, so add it in manually
			line = fmt.Sprintf("%s:%s:0: %s", match[1], match[2], match[3])
		}
		return okgo.NewIssueFromLine(line, wd)
	}, stdout)
}

func (c *Checker) RunCheckCmd(args []string, stdout io.Writer) {
	checker.AmalgomatedRunRawCheck(string(TypeName), args, stdout)
}
