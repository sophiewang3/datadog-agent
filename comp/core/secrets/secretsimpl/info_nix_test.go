// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !windows

package secretsimpl

import (
	"bytes"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testConfInfo = []byte(`---
instances:
- password: ENC[pass3]
- password: ENC[pass2]
`)
)

func TestGetExecutablePermissionsError(t *testing.T) {
	resolver := newEnabledSecretResolver()
	resolver.backendCommand = "some_command"

	_, err := resolver.getExecutablePermissions()
	assert.Error(t, err, "getExecutablePermissions should fail when secretBackendCommand file does not exists")
}

func setupSecretCommand(t *testing.T, resolver *secretResolver) (string, string) {
	dir := t.TempDir()

	resolver.backendCommand = filepath.Join(dir, "executable")
	f, err := os.Create(resolver.backendCommand)
	require.NoError(t, err)
	f.Close()
	os.Chmod(resolver.backendCommand, 0700)

	u, err := user.Current()
	require.NoError(t, err)
	currentUser, err := user.LookupId(u.Uid)
	require.NoError(t, err)
	currentGroup, err := user.LookupGroupId(u.Gid)
	require.NoError(t, err)

	return currentUser.Username, currentGroup.Name
}

func TestGetExecutablePermissionsSuccess(t *testing.T) {
	resolver := newEnabledSecretResolver()
	currentUser, currentGroup := setupSecretCommand(t, resolver)

	res, err := resolver.getExecutablePermissions()
	require.NoError(t, err)
	require.IsType(t, permissionsDetails{}, res)
	details := res.(permissionsDetails)
	assert.Equal(t, "100700", details.FileMode)
	assert.Equal(t, currentUser, details.Owner)
	assert.Equal(t, currentGroup, details.Group)
}

func TestDebugInfo(t *testing.T) {
	resolver := newEnabledSecretResolver()
	currentUser, currentGroup := setupSecretCommand(t, resolver)

	resolver.commandHookFunc = func(string) ([]byte, error) {
		res := []byte("{\"pass1\":{\"value\":\"password1\"},")
		res = append(res, []byte("\"pass2\":{\"value\":\"password2\"},")...)
		res = append(res, []byte("\"pass3\":{\"value\":\"password3\"}}")...)
		return res, nil
	}

	_, err := resolver.Decrypt(testConf, "test")
	require.NoError(t, err)
	_, err = resolver.Decrypt(testConfInfo, "test2")
	require.NoError(t, err)

	var buffer bytes.Buffer
	resolver.GetDebugInfo(&buffer)

	expectedResult := `=== Checking executable permissions ===
Executable path: ` + resolver.backendCommand + `
Executable permissions: OK, the executable has the correct permissions

Permissions Detail:
File mode: 100700
Owner: ` + currentUser + `
Group: ` + currentGroup + `

=== Secrets stats ===
Number of secrets decrypted: 3
Secrets handle decrypted:

- 'pass1':
	used in 'test' configuration in entry 'instances/password'
- 'pass2':
	used in 'test' configuration in entry 'instances/password'
	used in 'test2' configuration in entry 'instances/password'
- 'pass3':
	used in 'test2' configuration in entry 'instances/password'
`

	assert.Equal(t, expectedResult, buffer.String())
}

func TestDebugInfoError(t *testing.T) {
	resolver := newEnabledSecretResolver()
	resolver.backendCommand = "some_command"

	resolver.commandHookFunc = func(string) ([]byte, error) {
		res := []byte("{\"pass1\":{\"value\":\"password1\"},")
		res = append(res, []byte("\"pass2\":{\"value\":\"password2\"},")...)
		res = append(res, []byte("\"pass3\":{\"value\":\"password3\"}}")...)
		return res, nil
	}

	_, err := resolver.Decrypt(testConf, "test")
	require.NoError(t, err)
	_, err = resolver.Decrypt(testConfInfo, "test2")
	require.NoError(t, err)

	var buffer bytes.Buffer
	resolver.GetDebugInfo(&buffer)

	expectedResult := `=== Checking executable permissions ===
Executable path: some_command
Executable permissions: error: invalid executable 'some_command': can't stat it: no such file or directory

Permissions Detail:
Could not stat some_command: no such file or directory

=== Secrets stats ===
Number of secrets decrypted: 3
Secrets handle decrypted:

- 'pass1':
	used in 'test' configuration in entry 'instances/password'
- 'pass2':
	used in 'test' configuration in entry 'instances/password'
	used in 'test2' configuration in entry 'instances/password'
- 'pass3':
	used in 'test2' configuration in entry 'instances/password'
`

	assert.Equal(t, expectedResult, buffer.String())
}
