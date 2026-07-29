package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBenchHeaders(t *testing.T) {
	headers, err := parseBenchHeaders([]string{
		"Content-Type: application/json",
		"Authorization: Bearer abc",
		"X-Empty:",
	})

	require.NoError(t, err)
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "Bearer abc", headers["Authorization"])
	assert.Equal(t, "", headers["X-Empty"])
}

func TestParseBenchHeadersErrors(t *testing.T) {
	_, err := parseBenchHeaders([]string{"no-colon"})
	require.Error(t, err)

	_, err = parseBenchHeaders([]string{": value"})
	require.Error(t, err)
}

func TestBuildBenchConfig(t *testing.T) {
	// Body given, no method -> defaults to POST.
	cmd := getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("data", `{"a":1}`))
	require.NoError(t, cmd.Flags().Set("header", "Content-Type: application/json"))
	config, err := buildBenchConfig(cmd, "https://example.com")
	require.NoError(t, err)
	assert.Equal(t, "POST", config.Method)
	assert.Equal(t, `{"a":1}`, config.Body)
	assert.Equal(t, 10, config.Requests) // defaults to concurrency
	assert.Equal(t, "application/json", config.Headers["Content-Type"])

	// No body, no method -> GET; requests honoured.
	cmd = getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("requests", "25"))
	require.NoError(t, cmd.Flags().Set("concurrency", "5"))
	config, err = buildBenchConfig(cmd, "https://example.com")
	require.NoError(t, err)
	assert.Equal(t, "GET", config.Method)
	assert.Equal(t, 25, config.Requests)
	assert.Equal(t, 5, config.Concurrency)
}

func TestBuildBenchConfigDataFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"from":"file"}`), 0644))

	cmd := getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("data-file", path))
	config, err := buildBenchConfig(cmd, "https://example.com")
	require.NoError(t, err)
	assert.Equal(t, `{"from":"file"}`, config.Body)
	assert.Equal(t, "POST", config.Method)
}

func TestBuildBenchConfigErrors(t *testing.T) {
	cmd := getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("header", "bad-header"))
	_, err := buildBenchConfig(cmd, "https://example.com")
	require.Error(t, err)

	cmd = getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("concurrency", "0"))
	_, err = buildBenchConfig(cmd, "https://example.com")
	require.Error(t, err)

	cmd = getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("data-file", filepath.Join(t.TempDir(), "missing.json")))
	_, err = buildBenchConfig(cmd, "https://example.com")
	require.Error(t, err)

	cmd = getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("timeout", "-1"))
	_, err = buildBenchConfig(cmd, "https://example.com")
	require.Error(t, err)
}

func TestRunBenchAgainstServer(t *testing.T) {
	var hits int32
	mu := make(chan struct{}, 1)
	mu <- struct{}{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-mu
		hits++
		mu <- struct{}{}
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	config := benchConfig{
		URL:         server.URL,
		Method:      "POST",
		Headers:     map[string]string{"Content-Type": "application/json"},
		Body:        `{"ping":1}`,
		Requests:    12,
		Concurrency: 4,
		Timeout:     5,
	}

	results, elapsed := runBench(config)
	require.Len(t, results, 12)
	assert.Equal(t, int32(12), hits)
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(0))

	for i, r := range results {
		assert.Equal(t, i, r.Index)
		assert.Equal(t, http.StatusOK, r.StatusCode)
		assert.Equal(t, `{"ok":true}`, r.Body)
		assert.Equal(t, len(`{"ok":true}`), r.Size)
		assert.Empty(t, r.Error)
	}

	summary := summarize(results, elapsed)
	assert.Equal(t, 12, summary.Total)
	assert.Equal(t, 12, summary.Success)
	assert.Equal(t, 0, summary.Failed)
}

func TestRunBenchError(t *testing.T) {
	config := benchConfig{
		URL:         "http://127.0.0.1:1", // refused
		Method:      "GET",
		Requests:    3,
		Concurrency: 3,
		Timeout:     1,
	}

	results, elapsed := runBench(config)
	require.Len(t, results, 3)
	for _, r := range results {
		assert.NotEmpty(t, r.Error)
		assert.Equal(t, 0, r.StatusCode)
	}

	summary := summarize(results, elapsed)
	assert.Equal(t, 3, summary.Failed)
	assert.Equal(t, 0, summary.Success)
}

func TestSummarizeEmpty(t *testing.T) {
	summary := summarize(nil, time.Second)
	assert.Equal(t, 0, summary.Total)
	assert.Equal(t, int64(0), summary.AvgMS)
}

func TestSummarizeStats(t *testing.T) {
	results := []benchResult{
		{Index: 0, StatusCode: 200, LatencyMS: 10},
		{Index: 1, StatusCode: 200, LatencyMS: 20},
		{Index: 2, StatusCode: 500, LatencyMS: 30},
		{Index: 3, StatusCode: 0, LatencyMS: 40, Error: "boom"},
	}

	summary := summarize(results, 2*time.Second)
	assert.Equal(t, 4, summary.Total)
	assert.Equal(t, 2, summary.Success)
	assert.Equal(t, 2, summary.Failed)
	assert.Equal(t, int64(10), summary.MinMS)
	assert.Equal(t, int64(40), summary.MaxMS)
	assert.Equal(t, int64(25), summary.AvgMS)
	assert.InDelta(t, 2.0, summary.QPS, 0.001)
}

func TestPercentile(t *testing.T) {
	sorted := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	assert.Equal(t, int64(5), percentile(sorted, 50))
	assert.Equal(t, int64(9), percentile(sorted, 90))
	assert.Equal(t, int64(10), percentile(sorted, 99))
	assert.Equal(t, int64(0), percentile(nil, 50))
}

func TestMaybeStripBodies(t *testing.T) {
	report := benchReport{
		Results: []benchResult{{Index: 0, Body: "hello"}},
	}

	stripped := maybeStripBodies(report, false)
	assert.Empty(t, stripped.Results[0].Body)
	// Original is untouched.
	assert.Equal(t, "hello", report.Results[0].Body)

	kept := maybeStripBodies(report, true)
	assert.Equal(t, "hello", kept.Results[0].Body)
}

func TestWriteBenchReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	report := benchReport{
		Summary: benchSummary{Total: 1, Success: 1},
		Results: []benchResult{{Index: 0, StatusCode: 200, Body: "hi"}},
	}

	require.NoError(t, writeBenchReport(path, report))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"status_code": 200`)
	assert.Contains(t, string(data), `"body": "hi"`)
}

func TestPrintBenchTable(t *testing.T) {
	results := []benchResult{
		{Index: 0, StatusCode: 200, LatencyMS: 12, Size: 5, Body: "line1\nline2"},
		{Index: 1, StatusCode: 0, LatencyMS: 3, Error: "connection refused"},
	}
	summary := summarize(results, time.Second)

	output := captureStdout(t, func() {
		printBenchTable(results, summary, true)
	})
	assert.Contains(t, output, "Status")
	assert.Contains(t, output, "connection refused")
	assert.Contains(t, output, "Requests: 2 total")

	// Without bodies the Body column is absent.
	output = captureStdout(t, func() {
		printBenchTable(results, summary, false)
	})
	assert.NotContains(t, output, "line1 line2")
}

func TestRunBenchHTTPCmdEndToEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	reportPath := filepath.Join(t.TempDir(), "out.json")

	cmd := getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("requests", "4"))
	require.NoError(t, cmd.Flags().Set("concurrency", "2"))
	require.NoError(t, cmd.Flags().Set("output-file", reportPath))

	output := captureStdout(t, func() {
		require.NoError(t, runBenchHTTPCmd(cmd, []string{server.URL}))
	})
	assert.Contains(t, output, "Requests: 4 total, 4 success")
	assert.FileExists(t, reportPath)

	// json / yaml formats.
	cmd = getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("requests", "2"))
	require.NoError(t, cmd.Flags().Set("output", "json"))
	require.NoError(t, runBenchHTTPCmd(cmd, []string{server.URL}))

	cmd = getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("requests", "2"))
	require.NoError(t, cmd.Flags().Set("output", "yaml"))
	require.NoError(t, runBenchHTTPCmd(cmd, []string{server.URL}))

	// Unsupported format errors.
	cmd = getBenchHTTPCmd()
	require.NoError(t, cmd.Flags().Set("requests", "1"))
	require.NoError(t, cmd.Flags().Set("output", "xml"))
	require.Error(t, runBenchHTTPCmd(cmd, []string{server.URL}))

	// No URL -> help, no error.
	require.NoError(t, runBenchHTTPCmd(getBenchHTTPCmd(), []string{}))
}

func TestBenchCommandsWiring(t *testing.T) {
	assert.NotNil(t, GetBenchCmd())
	assert.NotNil(t, getBenchHTTPCmd())
}

func TestTruncateAndOneLine(t *testing.T) {
	assert.Equal(t, "a b c", oneLine("a\nb\rc"))
	assert.Equal(t, "hello", truncateText("hello", 10))
	assert.Equal(t, "hel…", truncateText("hello world", 4))
	assert.Equal(t, "h", truncateText("hello", 1))
}
