/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/dell/gocsi/mock/service"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestIdentityCmd(t *testing.T) {
	child := identityCmd

	// test case: no error
	err := child.PersistentPreRunE(child, []string{})
	assert.NoError(t, err)

	// save original func so we can revert
	cmd := RootCmd.PersistentPreRunE

	// test case: error
	// force RootCmd to return error
	RootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return fmt.Errorf("test error")
	}
	err = child.PersistentPreRunE(child, []string{})
	assert.Error(t, err)

	// restore original func back so other UT won't fail
	RootCmd.PersistentPreRunE = cmd
}

func TestGetPluginCapabilitiesCmd(t *testing.T) {
	var b bytes.Buffer
	originalGetStdout := getStdout
	getStdout = func() io.Writer {
		return &b
	}
	defer func() {
		getStdout = originalGetStdout
	}()

	child := pluginCapsCmd
	// set up root as required
	setupRoot(t, pluginCapsFormat)

	// set up the CSI client with a mock
	identity.client = service.NewClient()

	// Valid test case
	err := child.RunE(RootCmd, []string{""})
	assert.NoError(t, err)

	out := b.String()
	assert.Contains(t, out, "CONTROLLER_SERVICE")
	assert.Contains(t, out, "ONLINE")
}

func TestGetPluginInfoCmd(t *testing.T) {
	var b bytes.Buffer
	originalGetStdout := getStdout
	getStdout = func() io.Writer {
		return &b
	}
	defer func() {
		getStdout = originalGetStdout
	}()

	child := pluginInfoCmd
	// set up root as required
	setupRoot(t, pluginInfoFormat)

	// set up the CSI client with a mock
	identity.client = service.NewClient()

	// Valid test case
	err := child.RunE(RootCmd, []string{""})
	assert.NoError(t, err)

	out := b.String()
	want := `"mock.gocsi.rexray.com"	"1.1.0"	"url"="https://github.com/dell/gocsi/tree/master/mock"
`
	assert.Equal(t, want, out)
}

func TestProbeCmd(t *testing.T) {
	var b bytes.Buffer
	originalGetStdout := getStdout
	getStdout = func() io.Writer {
		return &b
	}
	defer func() {
		getStdout = originalGetStdout
	}()

	child := probeCmd
	// set up root as required
	setupRoot(t, probeFormat)

	// set up the CSI client with a mock
	identity.client = service.NewClient()

	// Valid test case
	err := child.RunE(RootCmd, []string{""})
	assert.NoError(t, err)

	out := b.String()
	assert.Contains(t, out, "true")
}
