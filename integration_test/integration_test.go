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

package integration_test

import (
	"os"
	"testing"

	"github.com/nmiyake/pkg/gofiles"
	"github.com/palantir/godel/framework/pluginapitester"
	"github.com/palantir/godel/pkg/products"
	"github.com/palantir/okgo/okgotester"
	"github.com/stretchr/testify/require"
)

const (
	okgoPluginLocator  = "com.palantir.okgo:check-plugin:1.0.0-rc3"
	okgoPluginResolver = "https://palantir.bintray.com/releases/{{GroupPath}}/{{Product}}/{{Version}}/{{Product}}-{{Version}}-{{OS}}-{{Arch}}.tgz"

	godelYML = `exclude:
  names:
    - "\\..+"
    - "vendor"
  paths:
    - "godel"
`
)

func TestCheck(t *testing.T) {
	assetPath, err := products.Bin("govet-asset")
	require.NoError(t, err)

	configFiles := map[string]string{
		"godel/config/godel.yml":        godelYML,
		"godel/config/check-plugin.yml": "",
	}

	pluginProvider, err := pluginapitester.NewPluginProviderFromLocator(okgoPluginLocator, okgoPluginResolver)
	require.NoError(t, err)

	prevVal := os.Getenv("GOMAXPROCS")
	if prevVal != "" {
		err = os.Setenv("GOMAXPROCS", prevVal)
		require.NoError(t, err)
	}

	err = os.Setenv("GOMAXPROCS", "1")
	require.NoError(t, err)

	okgotester.RunAssetCheckTest(t,
		pluginProvider,
		pluginapitester.NewAssetProvider(assetPath),
		"govet",
		"",
		[]okgotester.AssetTestCase{
			{
				Name: "vet failures",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

import "fmt"

func Foo() {
    num := 13
	fmt.Printf("%s", num)
}
`,
					},
					{
						RelPath: "bar/bar.go",
						Src: `package bar

import "fmt"

func Bar() {
    num := 13
	fmt.Printf("%s", num)
}
`,
					},
				},
				ConfigFiles: configFiles,
				WantError:   true,
				WantOutput: `Running govet...
./foo.go:7: Printf format %s has arg num of wrong type int
bar/bar.go:7: Printf format %s has arg num of wrong type int
Finished govet
Check(s) produced output: [govet]
`,
			},
			{
				Name: "vet failures from inner directory",
				Specs: []gofiles.GoFileSpec{
					{
						RelPath: "foo.go",
						Src: `package foo

import "fmt"

func Foo() {
    num := 13
	fmt.Printf("%s", num)
}
`,
					},
					{
						RelPath: "bar/bar.go",
						Src: `package bar

import "fmt"

func Bar() {
    num := 13
	fmt.Printf("%s", num)
}
`,
					},
					{
						RelPath: "inner/bar",
					},
				},
				ConfigFiles: configFiles,
				Wd:          "inner",
				WantError:   true,
				WantOutput: `Running govet...
../foo.go:7: Printf format %s has arg num of wrong type int
../bar/bar.go:7: Printf format %s has arg num of wrong type int
Finished govet
Check(s) produced output: [govet]
`,
			},
		},
	)
}

func TestUpgradeConfig(t *testing.T) {
	pluginProvider, err := pluginapitester.NewPluginProviderFromLocator(okgoPluginLocator, okgoPluginResolver)
	require.NoError(t, err)

	assetPath, err := products.Bin("govet-asset")
	require.NoError(t, err)
	assetProvider := pluginapitester.NewAssetProvider(assetPath)

	pluginapitester.RunUpgradeConfigTest(t,
		pluginProvider,
		[]pluginapitester.AssetProvider{assetProvider},
		[]pluginapitester.UpgradeConfigTestCase{
			{
				Name: `legacy configuration with empty "args" field is updated`,
				ConfigFiles: map[string]string{
					"godel/config/godel.yml": godelYML,
					"godel/config/check-plugin.yml": `
legacy-config: true
checks:
  govet:
    filters:
      - value: "should have comment or be unexported"
      - type: name
        value: ".*.pb.go"
`,
				},
				WantOutput: `Upgraded configuration for check-plugin.yml
`,
				WantFiles: map[string]string{
					"godel/config/check-plugin.yml": `release-tag: ""
checks:
  govet:
    skip: false
    priority: null
    config: {}
    filters:
    - type: ""
      value: should have comment or be unexported
    exclude:
      names:
      - .*.pb.go
      paths: []
exclude:
  names: []
  paths: []
`,
				},
			},
			{
				Name: `legacy configuration with non-empty "args" field fails`,
				ConfigFiles: map[string]string{
					"godel/config/godel.yml": godelYML,
					"godel/config/check-plugin.yml": `
legacy-config: true
checks:
  govet:
    args:
      - "-foo"
`,
				},
				WantError: true,
				WantOutput: `Failed to upgrade configuration:
	godel/config/check-plugin.yml: failed to upgrade check "govet" legacy configuration: failed to upgrade asset configuration: govet-asset does not support legacy configuration with a non-empty "args" field
`,
				WantFiles: map[string]string{
					"godel/config/check-plugin.yml": `
legacy-config: true
checks:
  govet:
    args:
      - "-foo"
`,
				},
			},
			{
				Name: `empty v0 config works`,
				ConfigFiles: map[string]string{
					"godel/config/godel.yml": godelYML,
					"godel/config/check-plugin.yml": `
checks:
  govet:
    skip: true
    # comment preserved
    config:
`,
				},
				WantOutput: ``,
				WantFiles: map[string]string{
					"godel/config/check-plugin.yml": `
checks:
  govet:
    skip: true
    # comment preserved
    config:
`,
				},
			},
			{
				Name: `non-empty v0 config does not work`,
				ConfigFiles: map[string]string{
					"godel/config/godel.yml": godelYML,
					"godel/config/check-plugin.yml": `
checks:
  govet:
    config:
      # comment
      key: value
`,
				},
				WantError: true,
				WantOutput: `Failed to upgrade configuration:
	godel/config/check-plugin.yml: failed to upgrade check "govet" configuration: failed to upgrade asset configuration: govet-asset does not currently support configuration
`,
				WantFiles: map[string]string{
					"godel/config/check-plugin.yml": `
checks:
  govet:
    config:
      # comment
      key: value
`,
				},
			},
		},
	)
}