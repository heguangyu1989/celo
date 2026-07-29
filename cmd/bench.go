package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/go-resty/resty/v2"
	"github.com/heguangyu1989/celo/pkg/p"
	"github.com/heguangyu1989/celo/pkg/utils"
	"github.com/spf13/cobra"
)

// benchConfig holds the parameters for an HTTP concurrency test.
type benchConfig struct {
	URL         string
	Method      string
	Headers     map[string]string
	Body        string
	Requests    int
	Concurrency int
	Timeout     int
}

// benchResult records the outcome of a single HTTP request.
type benchResult struct {
	Index      int    `json:"index" yaml:"index"`
	StatusCode int    `json:"status_code" yaml:"status_code"`
	LatencyMS  int64  `json:"latency_ms" yaml:"latency_ms"`
	Size       int    `json:"size" yaml:"size"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
	Body       string `json:"body,omitempty" yaml:"body,omitempty"`
}

// benchSummary aggregates statistics across all requests.
type benchSummary struct {
	Total   int     `json:"total" yaml:"total"`
	Success int     `json:"success" yaml:"success"`
	Failed  int     `json:"failed" yaml:"failed"`
	TotalMS int64   `json:"total_ms" yaml:"total_ms"`
	QPS     float64 `json:"qps" yaml:"qps"`
	MinMS   int64   `json:"min_ms" yaml:"min_ms"`
	MaxMS   int64   `json:"max_ms" yaml:"max_ms"`
	AvgMS   int64   `json:"avg_ms" yaml:"avg_ms"`
	P50MS   int64   `json:"p50_ms" yaml:"p50_ms"`
	P90MS   int64   `json:"p90_ms" yaml:"p90_ms"`
	P99MS   int64   `json:"p99_ms" yaml:"p99_ms"`
}

// benchReport is the full report written to file / printed as json|yaml.
type benchReport struct {
	Summary benchSummary  `json:"summary" yaml:"summary"`
	Results []benchResult `json:"results" yaml:"results"`
}

func GetBenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "Benchmark utilities",
		Long:  "Utilities for load and concurrency testing.",
	}

	cmd.AddCommand(getBenchHTTPCmd())
	return cmd
}

func getBenchHTTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http URL",
		Short: "Concurrent HTTP request test",
		Long: `Fire concurrent HTTP requests against a URL and report per-request
results plus aggregate statistics.

Set the method, headers and request body, run N total requests spread
across C concurrent goroutines, print the results as a table (or
json/yaml) and optionally write the full report (including every
response body) to a file.`,
		Example: `  # 100 requests, 20 in parallel
  celo bench http https://example.com -n 100 -c 20

  # POST with headers and body, save the full report
  celo bench http https://api.moonshot.cn/v1/chat/completions \
    -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $MOONSHOT_API_KEY" \
    -d '{"model":"kimi-k3-private-preview","messages":[{"role":"user","content":"你是？"}]}' \
    -n 50 -c 10 -o report.json

  # Read the request body from a file
  celo bench http https://example.com/api -X POST --data-file body.json -n 30 -c 5`,
		Args: cobra.MaximumNArgs(1),
		RunE: runBenchHTTPCmd,
	}

	cmd.Flags().StringP("method", "X", "", "HTTP method (default GET, or POST when a body is given)")
	cmd.Flags().StringArrayP("header", "H", nil, "request header 'Key: Value' (repeatable)")
	cmd.Flags().StringP("data", "d", "", "request body")
	cmd.Flags().String("data-file", "", "read request body from file")
	cmd.Flags().IntP("requests", "n", 0, "total number of requests (default = concurrency)")
	cmd.Flags().IntP("concurrency", "c", 10, "number of concurrent goroutines")
	cmd.Flags().Int("timeout", 30, "per-request timeout in seconds")
	cmd.Flags().StringP("output-file", "o", "", "write the full JSON report (with bodies) to this file")
	cmd.Flags().String("output", "table", "stdout format: table, json, yaml")
	cmd.Flags().Bool("with-body", false, "include response bodies in table/stdout output")
	return cmd
}

func runBenchHTTPCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	config, err := buildBenchConfig(cmd, args[0])
	if err != nil {
		p.Error(err.Error())
		return err
	}

	results, elapsed := runBench(config)
	summary := summarize(results, elapsed)
	report := benchReport{Summary: summary, Results: results}

	if outputFile, _ := cmd.Flags().GetString("output-file"); outputFile != "" {
		if err := writeBenchReport(outputFile, report); err != nil {
			p.Error(fmt.Sprintf("Failed to write report: %v", err))
			return err
		}
		p.Success(fmt.Sprintf("Report written to %s", outputFile))
	}

	withBody, _ := cmd.Flags().GetBool("with-body")
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return p.PrintJSON(maybeStripBodies(report, withBody))
	case "yaml":
		return p.PrintYAML(maybeStripBodies(report, withBody))
	case "table":
		printBenchTable(results, summary, withBody)
	default:
		p.Error(fmt.Sprintf("Unsupported output format: %s", output))
		return fmt.Errorf("unsupported output format: %s", output)
	}

	return nil
}

func buildBenchConfig(cmd *cobra.Command, url string) (benchConfig, error) {
	method, _ := cmd.Flags().GetString("method")
	headerArgs, _ := cmd.Flags().GetStringArray("header")
	data, _ := cmd.Flags().GetString("data")
	dataFile, _ := cmd.Flags().GetString("data-file")
	requests, _ := cmd.Flags().GetInt("requests")
	concurrency, _ := cmd.Flags().GetInt("concurrency")
	timeout, _ := cmd.Flags().GetInt("timeout")

	headers, err := parseBenchHeaders(headerArgs)
	if err != nil {
		return benchConfig{}, err
	}

	body := data
	if dataFile != "" {
		content, err := os.ReadFile(dataFile)
		if err != nil {
			return benchConfig{}, fmt.Errorf("failed to read data file: %w", err)
		}
		body = string(content)
	}

	if concurrency < 1 {
		return benchConfig{}, fmt.Errorf("concurrency must be at least 1")
	}
	if requests < 1 {
		requests = concurrency
	}
	if timeout < 0 {
		return benchConfig{}, fmt.Errorf("timeout must not be negative")
	}

	if method == "" {
		method = "GET"
		if body != "" {
			method = "POST"
		}
	}

	return benchConfig{
		URL:         url,
		Method:      strings.ToUpper(method),
		Headers:     headers,
		Body:        body,
		Requests:    requests,
		Concurrency: concurrency,
		Timeout:     timeout,
	}, nil
}

// parseBenchHeaders turns "Key: Value" strings into a header map.
func parseBenchHeaders(args []string) (map[string]string, error) {
	headers := make(map[string]string)
	for _, arg := range args {
		rawKey, rawValue, ok := strings.Cut(arg, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header (expected 'Key: Value'): %s", arg)
		}
		key := strings.TrimSpace(rawKey)
		if key == "" {
			return nil, fmt.Errorf("invalid header (empty key): %s", arg)
		}
		headers[key] = strings.TrimSpace(rawValue)
	}
	return headers, nil
}

// runBench dispatches config.Requests requests across a pool of
// config.Concurrency worker goroutines and returns per-request results
// (ordered by index) plus the total wall-clock duration.
func runBench(config benchConfig) ([]benchResult, time.Duration) {
	client := resty.New().SetTimeout(utils.SecondsToDuration(config.Timeout))

	results := make([]benchResult, config.Requests)
	jobs := make(chan int, config.Requests)
	for i := range config.Requests {
		jobs <- i
	}
	close(jobs)

	workers := utils.MinInt(config.Concurrency, config.Requests)
	var wg sync.WaitGroup
	start := time.Now()

	for range workers {
		wg.Go(func() {
			for idx := range jobs {
				results[idx] = doBenchRequest(client, config, idx)
			}
		})
	}

	wg.Wait()
	return results, time.Since(start)
}

func doBenchRequest(client *resty.Client, config benchConfig, index int) benchResult {
	req := client.R()
	if len(config.Headers) > 0 {
		req.SetHeaders(config.Headers)
	}
	if config.Body != "" {
		req.SetBody(config.Body)
	}

	start := time.Now()
	resp, err := req.Execute(config.Method, config.URL)
	result := benchResult{Index: index, LatencyMS: time.Since(start).Milliseconds()}

	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.StatusCode = resp.StatusCode()
	body := resp.Body()
	result.Size = len(body)
	result.Body = string(body)
	return result
}

func summarize(results []benchResult, elapsed time.Duration) benchSummary {
	summary := benchSummary{Total: len(results), TotalMS: elapsed.Milliseconds()}
	if len(results) == 0 {
		return summary
	}

	latencies := make([]int64, 0, len(results))
	var sum int64
	for _, r := range results {
		if r.Error == "" && r.StatusCode >= 200 && r.StatusCode < 400 {
			summary.Success++
		} else {
			summary.Failed++
		}
		latencies = append(latencies, r.LatencyMS)
		sum += r.LatencyMS
	}

	slices.Sort(latencies)
	summary.MinMS = latencies[0]
	summary.MaxMS = latencies[len(latencies)-1]
	summary.AvgMS = sum / int64(len(latencies))
	summary.P50MS = percentile(latencies, 50)
	summary.P90MS = percentile(latencies, 90)
	summary.P99MS = percentile(latencies, 99)
	if elapsed > 0 {
		summary.QPS = float64(len(results)) / elapsed.Seconds()
	}
	return summary
}

// percentile returns the p-th percentile (0-100) of an ascending-sorted
// slice using the nearest-rank method.
func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(math.Ceil(float64(p)/100*float64(len(sorted)))) - 1
	rank = max(rank, 0)
	if rank >= len(sorted) {
		rank = len(sorted) - 1
	}
	return sorted[rank]
}

// maybeStripBodies drops response bodies from a copy of the report when
// withBody is false, keeping stdout readable while the on-disk report
// still carries everything.
func maybeStripBodies(report benchReport, withBody bool) benchReport {
	if withBody {
		return report
	}
	stripped := make([]benchResult, len(report.Results))
	for i, r := range report.Results {
		r.Body = ""
		stripped[i] = r
	}
	return benchReport{Summary: report.Summary, Results: stripped}
}

func writeBenchReport(path string, report benchReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func printBenchTable(results []benchResult, summary benchSummary, withBody bool) {
	rows := make([]table.Row, len(results))
	maxErrLen := 5
	for i, r := range results {
		status := strconv.Itoa(r.StatusCode)
		if r.StatusCode == 0 {
			status = "-"
		}

		row := table.Row{
			strconv.Itoa(r.Index),
			status,
			strconv.FormatInt(r.LatencyMS, 10),
			strconv.Itoa(r.Size),
		}
		if withBody {
			row = append(row, truncateText(oneLine(r.Body), 60))
		}
		row = append(row, truncateText(oneLine(r.Error), 40))
		rows[i] = row

		if len(r.Error) > maxErrLen {
			maxErrLen = len(r.Error)
		}
	}

	columns := []table.Column{
		{Title: "#", Width: utils.MaxInt(len(strconv.Itoa(len(results))), 3)},
		{Title: "Status", Width: 6},
		{Title: "Latency(ms)", Width: 11},
		{Title: "Size", Width: 8},
	}
	if withBody {
		columns = append(columns, table.Column{Title: "Body", Width: 60})
	}
	columns = append(columns, table.Column{Title: "Error", Width: utils.MinInt(utils.MaxInt(maxErrLen, 5), 40)})

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(results)+1),
	)

	fmt.Println()
	fmt.Println(t.View())
	printBenchSummary(summary)
}

func printBenchSummary(s benchSummary) {
	fmt.Printf("\nRequests: %d total, %d success, %d failed\n", s.Total, s.Success, s.Failed)
	fmt.Printf("Duration: %dms, QPS: %.2f\n", s.TotalMS, s.QPS)
	fmt.Printf("Latency:  min %dms, avg %dms, p50 %dms, p90 %dms, p99 %dms, max %dms\n",
		s.MinMS, s.AvgMS, s.P50MS, s.P90MS, s.P99MS, s.MaxMS)
}

// oneLine collapses a multi-line string into a single trimmed line.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

// truncateText shortens s to at most max runes, appending an ellipsis
// when it had to cut.
func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}
