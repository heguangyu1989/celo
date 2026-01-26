package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/heguangyu1989/celo/pkg/p"
	"github.com/heguangyu1989/celo/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GetPasswordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "generate secure random passwords",
		Long:  "Generate secure random passwords with customizable length and character sets.",
		Example: `  # Generate a default 16-character password (lowercase + uppercase + digits)
  celo password

  # Generate a 20-character password
  celo password --length 20

  # Generate a password with all character types
  celo password --length 20 --upper --lower --digits --special

  # Generate 5 passwords
  celo password --count 5

  # Generate passwords with custom character set
  celo password --custom "abc123ABC!@#" --length 12`,
		RunE: runPasswordCmd,
	}

	cmd.Flags().Int("length", 16, "password length")
	cmd.Flags().Bool("upper", true, "include uppercase letters")
	cmd.Flags().Bool("lower", true, "include lowercase letters")
	cmd.Flags().Bool("digits", true, "include digits")
	cmd.Flags().Bool("special", false, "include special characters")
	cmd.Flags().String("custom", "", "custom character set (overrides other character type flags)")
	cmd.Flags().Int("count", 1, "number of passwords to generate")
	cmd.Flags().String("output", "table", "output format: json, yaml, table")

	return cmd
}

type passwordResult struct {
	Password string `json:"password" yaml:"password"`
	Index    int    `json:"index" yaml:"index"`
}

func runPasswordCmd(cmd *cobra.Command, args []string) error {
	length, _ := cmd.Flags().GetInt("length")
	useUpper, _ := cmd.Flags().GetBool("upper")
	useLower, _ := cmd.Flags().GetBool("lower")
	useDigits, _ := cmd.Flags().GetBool("digits")
	useSpecial, _ := cmd.Flags().GetBool("special")
	customChars, _ := cmd.Flags().GetString("custom")
	count, _ := cmd.Flags().GetInt("count")
	output, _ := cmd.Flags().GetString("output")

	// Validate parameters
	if length <= 0 {
		p.Error("Password length must be positive")
		return fmt.Errorf("invalid length: %d", length)
	}

	if count <= 0 {
		p.Error("Count must be positive")
		return fmt.Errorf("invalid count: %d", count)
	}

	// Build password options
	opts := utils.PasswordOptions{
		Length:      length,
		UseUpper:    useUpper,
		UseLower:    useLower,
		UseDigits:   useDigits,
		UseSpecial:  useSpecial,
		CustomChars: customChars,
	}

	// Generate passwords
	passwords, err := utils.GeneratePasswords(count, opts)
	if err != nil {
		p.Error(fmt.Sprintf("Failed to generate password: %v", err))
		return err
	}

	// Prepare results
	results := make([]passwordResult, count)
	for i, password := range passwords {
		results[i] = passwordResult{
			Index:    i + 1,
			Password: password,
		}
	}

	// Output results
	switch output {
	case "json":
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			p.Error(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return err
		}
		fmt.Println(string(data))

	case "yaml":
		data, err := yaml.Marshal(results)
		if err != nil {
			p.Error(fmt.Sprintf("Failed to marshal YAML: %v", err))
			return err
		}
		fmt.Println(string(data))

	case "table":
		if count == 1 {
			// For single password, just output the password
			p.Success(fmt.Sprintf("Generated password: %s", passwords[0]))
		} else {
			// For multiple passwords, show as table
			rows := make([]table.Row, count)
			maxLen := 0
			for i, password := range passwords {
				if len(password) > maxLen {
					maxLen = len(password)
				}
				rows[i] = table.Row{
					fmt.Sprintf("%d", i+1),
					password,
				}
			}

			columns := []table.Column{
				{Title: "#", Width: len(fmt.Sprintf("%d", count))},
				{Title: "Password", Width: utils.MaxInt(maxLen, 8)},
			}

			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithHeight(utils.MinInt(count+1, 20)), // Limit table height
			)
			fmt.Println()
			fmt.Println(t.View())
		}

	default:
		p.Error(fmt.Sprintf("Unsupported output format: %s", output))
		return fmt.Errorf("unsupported output format: %s", output)
	}

	// Show character set info for table output
	if output == "table" {
		fmt.Println()
		p.Info("Character set used:")
		var charTypes []string
		if customChars != "" {
			charTypes = append(charTypes, fmt.Sprintf("Custom (%d chars)", len(customChars)))
		} else {
			if useLower {
				charTypes = append(charTypes, "lowercase")
			}
			if useUpper {
				charTypes = append(charTypes, "uppercase")
			}
			if useDigits {
				charTypes = append(charTypes, "digits")
			}
			if useSpecial {
				charTypes = append(charTypes, "special")
			}
		}
		p.Info(fmt.Sprintf("  Types: %s", strings.Join(charTypes, ", ")))
		p.Info(fmt.Sprintf("  Length: %d", length))
		if count > 1 {
			p.Info(fmt.Sprintf("  Count: %d", count))
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(GetPasswordCmd())
}