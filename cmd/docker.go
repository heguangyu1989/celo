package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/heguangyu1989/celo/pkg/p"
	"github.com/heguangyu1989/celo/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GetDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docker",
		Short: "Docker utilities",
		Long:  "Various utilities for Docker image operations.",
	}

	cmd.AddCommand(getDockerCheckCmd())
	return cmd
}

func getDockerCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [image...]",
		Short: "Check if Docker images exist in registry",
		Long: `Check if Docker images exist in the registry using 'docker manifest inspect'.

Supports checking multiple images at once.
Image format: [registry/]repository:tag or [registry/]repository (defaults to 'latest')`,
		Example: `  # Check single image
  celo docker check nginx:latest

  # Check multiple images
  celo docker check nginx:latest redis:alpine

  # Check with custom registry
  celo docker check registry.example.com/myapp:v1.0`,
		RunE: runDockerCheckCmd,
	}

	cmd.Flags().String("output", "table", "output format: json, yaml, table")
	return cmd
}

type imageCheckResult struct {
	Image   string `json:"image" yaml:"image"`
	Exists  bool   `json:"exists" yaml:"exists"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

func runDockerCheckCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	output, _ := cmd.Flags().GetString("output")

	// Check if docker command is available
	if _, err := exec.LookPath("docker"); err != nil {
		p.Error("Docker command not found. Please install Docker.")
		return fmt.Errorf("docker not found: %w", err)
	}

	results := make([]imageCheckResult, 0, len(args))

	for _, image := range args {
		result := checkImageExists(image)
		results = append(results, result)
	}

	// Output results
	switch output {
	case "json":
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))

	case "yaml":
		data, err := yaml.Marshal(results)
		if err != nil {
			return err
		}
		fmt.Println(string(data))

	case "table":
		printCheckResultsTable(results)

	default:
		p.Error(fmt.Sprintf("Unsupported output format: %s", output))
		return fmt.Errorf("unsupported output format: %s", output)
	}

	return nil
}

func checkImageExists(image string) imageCheckResult {
	result := imageCheckResult{
		Image: image,
	}

	// Normalize image name
	image = normalizeImageName(image)

	// Run docker manifest inspect command
	// Using --insecure flag to support both http and https registries
	execCmd := exec.Command("docker", "manifest", "inspect", image)
	output, err := execCmd.CombinedOutput()

	if err != nil {
		// Command failed, image likely doesn't exist or not accessible
		result.Exists = false
		outputStr := strings.TrimSpace(string(output))
		if outputStr != "" {
			// Extract error message from output
			// Common errors: "manifest unknown", "no such manifest", "unauthorized"
			result.Message = extractErrorMessage(outputStr)
		} else {
			result.Message = "Image not found or registry not accessible"
		}
		return result
	}

	// Command succeeded, image exists
	result.Exists = true
	result.Message = "Image exists"

	return result
}

// normalizeImageName ensures image has a tag
// If no tag is specified, defaults to "latest"
func normalizeImageName(image string) string {
	// Check if image already has a tag
	if strings.Contains(image, ":") {
		// Check if it's not a port number (e.g., registry:5000/image)
		parts := strings.Split(image, "/")
		lastPart := parts[len(parts)-1]
		if strings.Contains(lastPart, ":") {
			return image
		}
	}
	// Add :latest tag
	return image + ":latest"
}

// extractErrorMessage extracts a user-friendly error message from docker output
func extractErrorMessage(output string) string {
	output = strings.TrimSpace(output)

	// Common error patterns
	switch {
	case strings.Contains(output, "manifest unknown"):
		return "Manifest unknown (image not found)"
	case strings.Contains(output, "no such manifest"):
		return "Manifest not found"
	case strings.Contains(output, "unauthorized"):
		return "Unauthorized (check credentials)"
	case strings.Contains(output, "connection refused"):
		return "Registry unreachable"
	default:
		// Return first line of output
		lines := strings.Split(output, "\n")
		if len(lines) > 0 && lines[0] != "" {
			return lines[0]
		}
		return "Check failed"
	}
}

func printCheckResultsTable(results []imageCheckResult) {
	rows := make([]table.Row, len(results))
	maxImageLen := 10 // minimum width for "Image"
	maxMsgLen := 7    // minimum width for "Message"

	for i, r := range results {
		status := "✓ Yes"
		if !r.Exists {
			status = "✗ No"
		}

		rows[i] = table.Row{
			r.Image,
			status,
			r.Message,
		}

		if len(r.Image) > maxImageLen {
			maxImageLen = len(r.Image)
		}
		if len(r.Message) > maxMsgLen {
			maxMsgLen = len(r.Message)
		}
	}

	columns := []table.Column{
		{Title: "Image", Width: utils.MaxInt(maxImageLen, 20)},
		{Title: "Exists", Width: 6},
		{Title: "Message", Width: utils.MaxInt(maxMsgLen, 20)},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(results)+1),
	)

	fmt.Println()
	fmt.Println(t.View())
}
