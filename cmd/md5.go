package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/heguangyu1989/celo/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GetMD5Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "md5",
		Short: " calculate a message-digest fingerprint (checksum) for a file or string",
		RunE:  runMD5Cmd,
	}
	cmd.Flags().String("output", "json", "output format: json, yaml, table")
	return cmd
}

type calType string

const (
	calTypeString calType = "string"
	calTypeFile   calType = "file"
	calTypeDir    calType = "dir"
)

type calData struct {
	Name string  `json:"name"`
	Type calType `json:"type"`
	MD5  string  `json:"md5"`
}

func (c *calData) Init() {
	data := make([]byte, 0)
	switch c.Type {
	case calTypeString:
		data = []byte(c.Name)
	case calTypeFile:
		data, _ = os.ReadFile(c.Name)
	}
	if len(data) > 0 {
		m := md5.New()
		m.Write(data)
		c.MD5 = hex.EncodeToString(m.Sum(nil))
	}
}

func runMD5Cmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	dataList := make([]calData, 0)
	for _, arg := range args {
		it := calData{
			Name: arg,
		}
		f, err := os.Stat(arg)
		if os.IsNotExist(err) {
			it.Type = calTypeString
		} else {
			if f.IsDir() {
				it.Type = calTypeDir
			} else {
				it.Type = calTypeFile
			}
		}
		it.Init()
		dataList = append(dataList, it)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		_data, err := json.MarshalIndent(dataList, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(_data))
	case "yaml":
		_data, _ := yaml.Marshal(dataList)
		fmt.Println(string(_data))
	case "table":
		maxNameLen := 0
		maxMD5Len := 0
		maxTypeLen := 0
		rows := make([]table.Row, 0)
		for _, it := range dataList {
			rows = append(rows, table.Row{
				it.Name,
				string(it.Type),
				it.MD5,
			})
			maxNameLen = utils.MaxInt(maxNameLen, len(it.Name))
			maxMD5Len = utils.MaxInt(maxMD5Len, len(it.MD5))
			maxTypeLen = utils.MaxInt(maxTypeLen, len(it.Type))
		}

		columns := []table.Column{
			{Title: "Name", Width: maxNameLen},
			{Title: "Type", Width: maxTypeLen},
			{Title: "MD5", Width: maxMD5Len},
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithHeight(7),
		)
		fmt.Println(t.View())
	}
	return nil
}
