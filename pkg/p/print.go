package p

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
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
