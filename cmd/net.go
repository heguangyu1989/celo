package cmd

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/table"
	"github.com/heguangyu1989/celo/pkg/p"
	"github.com/heguangyu1989/celo/pkg/utils"
	"github.com/spf13/cobra"
)

func GetNetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "net",
		Short: "Network utilities",
		Long:  "Various utilities for network diagnostics and port checking.",
	}

	cmd.AddCommand(getNetPortCmd())
	return cmd
}

func getNetPortCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "port [port|start-end]",
		Short: "Check TCP port status",
		Long: `Check TCP port listening status on localhost.

Supports checking single port or port range.
Will display port status and attempt to identify the service/process using the port.`,
		Example: `  # Check single port
  celo net port 8080

  # Check port range
  celo net port 8000-8100

  # Check multiple specific ports
  celo net port 22 80 443 3306 8080`,
		RunE: runNetPortCmd,
	}

	cmd.Flags().String("output", "table", "output format: table, json, yaml")
	cmd.Flags().Int("timeout", 2, "connection timeout in seconds")
	return cmd
}

type portCheckResult struct {
	Port        int    `json:"port" yaml:"port"`
	Status      string `json:"status" yaml:"status"`
	Service     string `json:"service,omitempty" yaml:"service,omitempty"`
	ProcessName string `json:"process_name,omitempty" yaml:"process_name,omitempty"`
	PID         string `json:"pid,omitempty" yaml:"pid,omitempty"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

func runNetPortCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	output, _ := cmd.Flags().GetString("output")
	timeout, _ := cmd.Flags().GetInt("timeout")

	// Parse all ports from arguments
	ports, err := parsePorts(args)
	if err != nil {
		p.Error(fmt.Sprintf("Invalid port specification: %v", err))
		return err
	}

	if len(ports) == 0 {
		p.Error("No valid ports to check")
		return fmt.Errorf("no valid ports")
	}

	// Check ports concurrently
	results := checkPortsConcurrently(ports, timeout)

	// Output results
	switch output {
	case "json":
		return p.PrintJSON(results)
	case "yaml":
		return p.PrintYAML(results)
	case "table":
		printPortResultsTable(results)
	default:
		p.Error(fmt.Sprintf("Unsupported output format: %s", output))
		return fmt.Errorf("unsupported output format: %s", output)
	}

	return nil
}

// parsePorts parses port arguments which can be:
// - Single port: "8080"
// - Port range: "8000-8100"
// - Multiple ports: "22 80 443"
func parsePorts(args []string) ([]int, error) {
	var ports []int
	seen := make(map[int]bool)

	for _, arg := range args {
		// Check if it's a range (contains "-")
		if strings.Contains(arg, "-") {
			parts := strings.Split(arg, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid port range: %s", arg)
			}

			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start port: %s", parts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end port: %s", parts[1])
			}

			if start > end {
				return nil, fmt.Errorf("start port %d is greater than end port %d", start, end)
			}

			if start < 1 || end > 65535 {
				return nil, fmt.Errorf("port range must be between 1 and 65535")
			}

			// Limit range size to prevent abuse
			if end-start > 1000 {
				return nil, fmt.Errorf("port range too large (max 1000 ports)")
			}

			for p := start; p <= end; p++ {
				if !seen[p] {
					ports = append(ports, p)
					seen[p] = true
				}
			}
		} else {
			// Single port
			port, err := strconv.Atoi(strings.TrimSpace(arg))
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", arg)
			}

			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("port %d is out of range (1-65535)", port)
			}

			if !seen[port] {
				ports = append(ports, port)
				seen[port] = true
			}
		}
	}

	return ports, nil
}

func checkPortsConcurrently(ports []int, timeout int) []portCheckResult {
	results := make([]portCheckResult, len(ports))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 50) // Limit concurrent checks

	for i, port := range ports {
		wg.Add(1)
		go func(index, p int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results[index] = checkPort(p, timeout)
		}(i, port)
	}

	wg.Wait()
	return results
}

func checkPort(port int, timeout int) portCheckResult {
	result := portCheckResult{
		Port:   port,
		Status: "closed",
	}

	// Try to connect to the port using standard library
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, utils.SecondsToDuration(timeout))

	if err != nil {
		// Port is closed or filtered
		result.Status = "closed"
		return result
	}

	// Port is open
	conn.Close()
	result.Status = "open"

	// Try to identify the service
	result.Service = identifyService(port)

	// Try to get process information
	processName, pid := getProcessUsingPort(port)
	result.ProcessName = processName
	result.PID = pid

	return result
}

// identifyService returns well-known service name for common ports
func identifyService(port int) string {
	wellKnownPorts := map[int]string{
		20:    "FTP-DATA",
		21:    "FTP",
		22:    "SSH",
		23:    "Telnet",
		25:    "SMTP",
		53:    "DNS",
		80:    "HTTP",
		110:   "POP3",
		143:   "IMAP",
		443:   "HTTPS",
		445:   "SMB",
		465:   "SMTPS",
		587:   "SMTP-Submission",
		631:   "IPP",
		993:   "IMAPS",
		995:   "POP3S",
		1433:  "MSSQL",
		1521:  "Oracle",
		3306:  "MySQL",
		3389:  "RDP",
		5432:  "PostgreSQL",
		5900:  "VNC",
		6379:  "Redis",
		8080:  "HTTP-Proxy",
		8443:  "HTTPS-Alt",
		9200:  "Elasticsearch",
		27017: "MongoDB",
	}

	if service, ok := wellKnownPorts[port]; ok {
		return service
	}
	return ""
}

// getProcessUsingPort attempts to find the process using the specified port
// This uses platform-specific commands for maximum compatibility
func getProcessUsingPort(port int) (processName, pid string) {
	switch runtime.GOOS {
	case "darwin":
		return getProcessOnDarwin(port)
	case "linux":
		return getProcessOnLinux(port)
	case "windows":
		return getProcessOnWindows(port)
	default:
		return "", ""
	}
}

func getProcessOnDarwin(port int) (string, string) {
	// Try lsof first (most common on macOS)
	cmd := exec.Command("lsof", "-nP", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-Fpcn")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return parseLsofOutput(string(output))
	}

	// Fallback to netstat
	cmd = exec.Command("netstat", "-anv", "-p", "tcp")
	output, err = cmd.Output()
	if err == nil {
		return parseNetstatOutput(string(output), port, "darwin")
	}

	return "", ""
}

func getProcessOnLinux(port int) (string, string) {
	// Try ss command first (modern replacement for netstat)
	cmd := exec.Command("ss", "-tlnp", "sport", "=", ":"+strconv.Itoa(port))
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		name, pid := parseSSOutput(string(output))
		if name != "" || pid != "" {
			return name, pid
		}
	}

	// Try netstat
	cmd = exec.Command("netstat", "-tlnp")
	output, err = cmd.Output()
	if err == nil {
		name, pid := parseNetstatOutput(string(output), port, "linux")
		if name != "" || pid != "" {
			return name, pid
		}
	}

	// Try lsof
	cmd = exec.Command("lsof", "-nP", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		return parseLsofOutput(string(output))
	}

	return "", ""
}

func getProcessOnWindows(port int) (string, string) {
	// Use netstat with -o to get PID
	cmd := exec.Command("netstat", "-ano", "-p", "tcp")
	output, err := cmd.Output()
	if err != nil {
		return "", ""
	}

	lines := strings.Split(string(output), "\n")
	portStr := ":" + strconv.Itoa(port)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			// Check if local address contains our port
			if strings.HasSuffix(fields[1], portStr) && fields[3] == "LISTENING" {
				pid := fields[4]
				// Get process name from PID
				name := getWindowsProcessName(pid)
				return name, pid
			}
		}
	}

	return "", ""
}

func getWindowsProcessName(pid string) string {
	cmd := exec.Command("tasklist", "/FI", "PID eq "+pid, "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse CSV output: "process.exe","1234",...
	line := strings.TrimSpace(string(output))
	parts := strings.Split(line, "\"")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func parseLsofOutput(output string) (string, string) {
	var pid, name string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		prefix := line[0]
		value := line[1:]

		switch prefix {
		case 'p':
			pid = value
		case 'c':
			name = value
		case 'n':
			// Connection info, we can stop here
			if name != "" && pid != "" {
				return name, pid
			}
		}
	}

	return name, pid
}

func parseSSOutput(output string) (string, string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Skip header
		if strings.Contains(line, "State") || strings.Contains(line, "Recv-Q") {
			continue
		}

		// Look for process info in parentheses
		if idx := strings.Index(line, "users:"); idx != -1 {
			// Format: users:(("process",pid=1234,fd=5))
			info := line[idx:]
			// Extract process name
			if start := strings.Index(info, "\""); start != -1 {
				if end := strings.Index(info[start+1:], "\""); end != -1 {
					name := info[start+1 : start+1+end]
					// Extract PID
					if pidStart := strings.Index(info, "pid="); pidStart != -1 {
						pidEnd := strings.Index(info[pidStart:], ",")
						if pidEnd == -1 {
							pidEnd = strings.Index(info[pidStart:], ")")
						}
						if pidEnd != -1 {
							pid := info[pidStart+4 : pidStart+pidEnd]
							return name, pid
						}
					}
					return name, ""
				}
			}
		}
	}
	return "", ""
}

func parseNetstatOutput(output string, port int, platform string) (string, string) {
	lines := strings.Split(output, "\n")
	portStr := ":" + strconv.Itoa(port)

	for _, line := range lines {
		fields := strings.Fields(line)

		switch platform {
		case "darwin":
			// macOS netstat format: Proto Recv-Q Send-Q Local Address Foreign Address (state) pid
			if len(fields) >= 8 {
				if strings.HasSuffix(fields[3], portStr) && strings.Contains(fields[6], "LISTEN") {
					return fields[7], ""
				}
			}
		case "linux":
			// Linux netstat -tlnp format: Proto Recv-Q Send-Q Local Address Foreign Address State PID/Program
			if len(fields) >= 7 {
				if strings.HasSuffix(fields[3], portStr) && fields[5] == "LISTEN" {
					// PID/Program name format: "1234/process" or "-"
					pidProg := fields[6]
					if idx := strings.Index(pidProg, "/"); idx != -1 {
						return pidProg[idx+1:], pidProg[:idx]
					}
				}
			}
		}
	}

	return "", ""
}

func printPortResultsTable(results []portCheckResult) {
	rows := make([]table.Row, len(results))
	maxServiceLen := 7
	maxProcessLen := 7

	for i, r := range results {
		status := "✗ Closed"
		if r.Status == "open" {
			status = "✓ Open"
		}

		processInfo := r.ProcessName
		if r.PID != "" {
			if processInfo != "" {
				processInfo = fmt.Sprintf("%s (PID: %s)", processInfo, r.PID)
			} else {
				processInfo = fmt.Sprintf("PID: %s", r.PID)
			}
		}

		rows[i] = table.Row{
			strconv.Itoa(r.Port),
			status,
			r.Service,
			processInfo,
		}

		if len(r.Service) > maxServiceLen {
			maxServiceLen = len(r.Service)
		}
		if len(processInfo) > maxProcessLen {
			maxProcessLen = len(processInfo)
		}
	}

	columns := []table.Column{
		{Title: "Port", Width: 6},
		{Title: "Status", Width: 10},
		{Title: "Service", Width: utils.MaxInt(maxServiceLen, 12)},
		{Title: "Process", Width: utils.MaxInt(maxProcessLen, 20)},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(results)+1),
	)

	fmt.Println()
	fmt.Println(t.View())

	// Summary
	openCount := 0
	for _, r := range results {
		if r.Status == "open" {
			openCount++
		}
	}

	fmt.Printf("\nTotal: %d ports checked, %d open, %d closed\n",
		len(results), openCount, len(results)-openCount)
}
