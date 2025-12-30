package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/heguangyu1989/celo/pkg/p"
	"github.com/heguangyu1989/celo/pkg/utils"
	"github.com/spf13/cobra"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
	rootPath string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		err := os.Remove(filepath.Join(m.rootPath, ".env"))
		if err != nil {
			panic(err)
		}
		err = os.Symlink(filepath.Join(m.rootPath, m.choice), filepath.Join(m.rootPath, ".env"))
		if err != nil {
			panic(err)
		}
		return quitTextStyle.Render("create link .env ->" + m.choice)
	}
	if m.quitting {
		return quitTextStyle.Render("nothing changed")
	}
	return "\n" + m.list.View()
}

func GetEnvListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "File-based environment variable switching",
		RunE:  runEnvListCmd,
	}
	return cmd
}

func runEnvListCmd(cmd *cobra.Command, args []string) error {
	var rootPath string
	var err error
	if len(args) != 0 {
		rootPath = args[0]
	} else {
		rootPath, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	f, err := os.Stat(rootPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", rootPath)
	} else {
		if !f.IsDir() {
			return fmt.Errorf("%s is not a directory", rootPath)
		}
	}
	envFiles, err := utils.FindAllEnvFiles(rootPath)
	if err != nil {
		return fmt.Errorf("could not find all env files: %w", err)
	}
	if len(envFiles) == 0 {
		p.Error("no env files found")
		return nil
	}
	items := make([]list.Item, 0, len(envFiles))
	for _, it := range envFiles {
		items = append(items, item(it))
	}

	title := fmt.Sprintf("Env files(%s):", rootPath)
	exitEnv := filepath.Join(rootPath, ".env")
	f, err = os.Lstat(exitEnv)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			target, err := os.Readlink(exitEnv)
			if err != nil {
				return fmt.Errorf("could not read symlink: %w", err)
			}
			title += fmt.Sprintf("\n .env -> %s", target)
		} else {
			p.Error(fmt.Sprintf("root path %s .env must be a symlink", rootPath))
			return nil
		}
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{list: l, rootPath: rootPath}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return nil
}
