package cmd

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePorts(t *testing.T) {
	ports, err := parsePorts([]string{"22", "80-82", "22"})

	require.NoError(t, err)
	assert.Equal(t, []int{22, 80, 81, 82}, ports)
}

func TestParsePortsErrors(t *testing.T) {
	tests := [][]string{
		{"abc"},
		{"0"},
		{"65536"},
		{"9-8"},
		{"1-1003"},
		{"1-two"},
		{"1-2-3"},
	}

	for _, tt := range tests {
		t.Run(tt[0], func(t *testing.T) {
			_, err := parsePorts(tt)
			require.Error(t, err)
		})
	}
}

func TestIdentifyService(t *testing.T) {
	assert.Equal(t, "HTTP", identifyService(80))
	assert.Equal(t, "Redis", identifyService(6379))
	assert.Empty(t, identifyService(65000))
}

func TestParseProcessOutputs(t *testing.T) {
	name, pid := parseLsofOutput("p123\ncnode\nn*:3000\n")
	assert.Equal(t, "node", name)
	assert.Equal(t, "123", pid)

	name, pid = parseSSOutput(`LISTEN 0 4096 *:8080 *:* users:(("nginx",pid=42,fd=6))`)
	assert.Equal(t, "nginx", name)
	assert.Equal(t, "42", pid)

	name, pid = parseNetstatOutput("tcp 0 0 0.0.0.0:8080 0.0.0.0:* LISTEN 77/app", 8080, "linux")
	assert.Equal(t, "app", name)
	assert.Equal(t, "77", pid)

	name, pid = parseNetstatOutput("tcp4 0 0 127.0.0.1:8080 *.* x LISTEN 99", 8080, "darwin")
	assert.Equal(t, "99", name)
	assert.Empty(t, pid)
}

func TestCheckPortClosed(t *testing.T) {
	closed := checkPort(1, 0)
	assert.Equal(t, "closed", closed.Status)
}

func TestCheckPortsConcurrently(t *testing.T) {
	results := checkPortsConcurrently([]int{1, 2}, 0)

	require.Len(t, results, 2)
	assert.Equal(t, 1, results[0].Port)
	assert.Equal(t, 2, results[1].Port)
}

func TestRunNetPortCmdValidationAndOutput(t *testing.T) {
	cmd := getNetPortCmd()
	require.NoError(t, cmd.Flags().Set("output", "json"))

	err := runNetPortCmd(cmd, []string{"1"})
	require.NoError(t, err)

	cmd = getNetPortCmd()
	require.NoError(t, cmd.Flags().Set("output", "xml"))
	err = runNetPortCmd(cmd, []string{"1"})
	require.Error(t, err)

	err = runNetPortCmd(getNetPortCmd(), []string{"bad"})
	require.Error(t, err)

	cmd = getNetPortCmd()
	require.NoError(t, cmd.Flags().Set("output", "yaml"))
	require.NoError(t, runNetPortCmd(cmd, []string{"1"}))

	cmd = getNetPortCmd()
	require.NoError(t, cmd.Flags().Set("output", "table"))
	require.NoError(t, runNetPortCmd(cmd, []string{"1"}))
}

func TestNetCommandsAndTableOutput(t *testing.T) {
	assert.NotNil(t, GetNetCmd())
	assert.NotNil(t, getNetPortCmd())

	output := captureStdout(t, func() {
		printPortResultsTable([]portCheckResult{
			{Port: 80, Status: "open", Service: "HTTP", ProcessName: "server", PID: "10"},
			{Port: 81, Status: "closed"},
		})
	})
	assert.Contains(t, output, "HTTP")
	assert.Contains(t, output, "Total: 2 ports checked")
}

func TestGetProcessUsingPortUnsupportedBranchIsHarmless(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" || runtime.GOOS == "windows" {
		t.Skip("default branch depends on unsupported GOOS")
	}
	name, pid := getProcessUsingPort(1)
	assert.Empty(t, name)
	assert.Empty(t, pid)
}

func TestProcessLookupWithFakeCommands(t *testing.T) {
	prependFakeCommand(t, "lsof", `#!/bin/sh
echo 'p321'
echo 'cfake-lsof'
echo 'n*:1234'
`)
	name, pid := getProcessOnDarwin(1234)
	assert.Equal(t, "fake-lsof", name)
	assert.Equal(t, "321", pid)

	prependFakeCommand(t, "ss", `#!/bin/sh
echo 'LISTEN 0 4096 *:8080 *:* users:(("fake-ss",pid=88,fd=6))'
`)
	name, pid = getProcessOnLinux(8080)
	assert.Equal(t, "fake-ss", name)
	assert.Equal(t, "88", pid)

	prependFakeCommand(t, "netstat", `#!/bin/sh
echo 'TCP 127.0.0.1:9000 0.0.0.0:0 LISTENING 77'
`)
	prependFakeCommand(t, "tasklist", `#!/bin/sh
echo '"fake.exe","77","Console","1","1,000 K"'
`)
	name, pid = getProcessOnWindows(9000)
	assert.Equal(t, "fake.exe", name)
	assert.Equal(t, "77", pid)
	assert.Equal(t, "fake.exe", getWindowsProcessName("77"))

	name, pid = getProcessUsingPort(1234)
	if runtime.GOOS == "darwin" {
		assert.Equal(t, "fake-lsof", name)
		assert.Equal(t, "321", pid)
	} else {
		assert.NotPanics(t, func() { _, _ = name, pid })
	}
}
