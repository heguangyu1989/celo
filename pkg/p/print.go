package p

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

var errStyle = lipgloss.NewStyle().
	Bold(true).
	Background(lipgloss.Color("#f45656"))

var infoStyle = lipgloss.NewStyle()

var successStyle = lipgloss.NewStyle().Bold(true)

func Error(data string) {
	fmt.Println(errStyle.Render(data))
}

func Info(data string) {
	fmt.Println(infoStyle.Render(data))
}

func Success(data string) {
	fmt.Println(successStyle.Render(data))
}

// PrintJSON prints data as formatted JSON
func PrintJSON(data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonBytes))
	return nil
}

// PrintYAML prints data as YAML
func PrintYAML(data interface{}) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Print(string(yamlBytes))
	return nil
}
