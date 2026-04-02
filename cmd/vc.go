package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func GetVCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vc",
		Short: "Manage VSCode Server services",
		Long:  "Manage VSCode Server processes and clean up folders on remote servers.",
	}

	cmd.AddCommand(getVCSkillAllCmd())
	cmd.AddCommand(getVCCleanCmd())

	return cmd
}

// getVCSkillAllCmd 杀死所有vscode-server进程
func getVCSkillAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "skill-all",
		Short: "Kill all VSCode Server processes",
		Long:  "Find and forcefully kill all running vscode-server processes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 查找所有vscode-server进程
			psCmd := exec.Command("ps", "aux")
			grepCmd := exec.Command("grep", "vscode-server")

			// 管道连接
			pipe, err := psCmd.StdoutPipe()
			if err != nil {
				return fmt.Errorf("创建管道失败: %v", err)
			}
			grepCmd.Stdin = pipe

			// 启动命令
			if err := psCmd.Start(); err != nil {
				return fmt.Errorf("执行ps命令失败: %v", err)
			}

			// 获取grep输出
			output, err := grepCmd.Output()
			if err != nil && err.Error() != "exit status 1" {
				return fmt.Errorf("执行grep命令失败: %v", err)
			}

			// 等待ps命令结束
			_ = psCmd.Wait()

			if len(output) == 0 {
				fmt.Println("未发现运行的vscode-server进程")
				return nil
			}

			lines := strings.Split(string(output), "\n")
			var pids []string
			for _, line := range lines {
				if strings.TrimSpace(line) == "" || strings.Contains(line, "grep") {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					pid := fields[1]
					if pid != "" && pid != " " {
						pids = append(pids, pid)
					}
				}
			}

			if len(pids) == 0 {
				fmt.Println("未发现运行的vscode-server进程")
				return nil
			}

			fmt.Printf("发现 %d 个vscode-server进程\n", len(pids))
			for _, pid := range pids {
				killCmd := exec.Command("kill", "-9", pid)
				if err := killCmd.Run(); err != nil {
					fmt.Printf("  杀死进程 %s 失败: %v\n", pid, err)
				} else {
					fmt.Printf("  ✓ 已杀死进程 %s\n", pid)
				}
			}

			return nil
		},
	}
}

// getVCCleanCmd 清理.vscode-server文件夹
func getVCCleanCmd() *cobra.Command {
	var (
		keepVersions int
		autoConfirm  bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean .vscode-server folders",
		Long:  "Scan and clean old versions, logs, and cache files in .vscode-server. Keeps the latest 3 versions by default.",
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("获取用户主目录失败: %v", err)
			}

			vscodeDir := filepath.Join(homeDir, ".vscode-server")
			if _, err := os.Stat(vscodeDir); os.IsNotExist(err) {
				return fmt.Errorf(".vscode-server目录不存在: %s", vscodeDir)
			}

			fmt.Printf("开始扫描: %s\n", vscodeDir)
			fmt.Println(strings.Repeat("=", 40))

			// 统计清理前后大小
			beforeSize, err := getDirSize(vscodeDir)
			if err != nil {
				beforeSize = 0
			}

			// 分析目录
			analysis, err := analyzeVSCodeDir(vscodeDir, keepVersions)
			if err != nil {
				return err
			}

			// 显示分析结果
			if err := displayAnalysis(analysis); err != nil {
				return err
			}

			// 确认清理
			if !autoConfirm {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("\n确认清理? [y/N] ")
				resp, _ := reader.ReadString('\n')
				resp = strings.TrimSpace(strings.ToLower(resp))
				if resp != "y" && resp != "yes" {
					fmt.Println("已取消清理")
					return nil
				}
			}

			// 执行清理
			cleanedSize, err := performCleanup(analysis, keepVersions)
			if err != nil {
				return err
			}

			// 显示结果
			afterSize, _ := getDirSize(vscodeDir)
			fmt.Println("\n" + strings.Repeat("=", 40))
			fmt.Printf("清理前: %s\n", formatSize(beforeSize))
			fmt.Printf("清理后: %s\n", formatSize(afterSize))
			fmt.Printf("释放空间: %s\n", formatSize(cleanedSize))

			return nil
		},
	}

	cmd.Flags().IntVarP(&keepVersions, "keep", "k", 3, "保留的CLI版本数量")
	cmd.Flags().BoolVarP(&autoConfirm, "yes", "y", false, "自动确认，无需交互")

	return cmd
}

// analysisResult 分析结果
type analysisResult struct {
	TotalSize     int64
	Items         []cleanupItem
	CLIVersionDir string
	LRUPath       string
}

// cleanupItem 待清理项目
type cleanupItem struct {
	Path        string
	Size        int64
	Description string
}

// analyzeVSCodeDir 分析.vscode-server目录
func analyzeVSCodeDir(vscodeDir string, keepCount int) (*analysisResult, error) {
	result := &analysisResult{
		CLIVersionDir: filepath.Join(vscodeDir, "cli", "servers"),
		LRUPath:       filepath.Join(vscodeDir, "cli", "servers", "lru.json"),
	}

	// 1. 分析CLI版本
	cliItems, err := analyzeCLIVersions(result.CLIVersionDir, result.LRUPath, keepCount)
	if err != nil {
		return nil, err
	}
	result.Items = append(result.Items, cliItems...)

	// 2. 分析缓存和日志
	cacheItems := analyzeCacheAndLogs(vscodeDir)
	result.Items = append(result.Items, cacheItems...)

	// 3. 分析临时文件
	tmpItems := analyzeTempFiles(vscodeDir)
	result.Items = append(result.Items, tmpItems...)

	// 计算总大小
	for _, item := range result.Items {
		result.TotalSize += item.Size
	}

	return result, nil
}

// analyzeCLIVersions 分析CLI版本
type lruData []string

func analyzeCLIVersions(cliDir, lruPath string, keepCount int) ([]cleanupItem, error) {
	var items []cleanupItem

	// 读取LRU文件
	lruFile, err := os.Open(lruPath)
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil
		}
		return nil, fmt.Errorf("读取LRU文件失败: %v", err)
	}
	defer func() {
		_ = lruFile.Close()
	}()

	data, err := io.ReadAll(lruFile)
	if err != nil {
		return nil, fmt.Errorf("读取LRU内容失败: %v", err)
	}

	var versions lruData
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, fmt.Errorf("解析LRU文件失败: %v", err)
	}

	// 检查cli目录是否存在
	if _, err := os.Stat(cliDir); os.IsNotExist(err) {
		return items, nil
	}

	// 遍历所有版本目录
	entries, err := os.ReadDir(cliDir)
	if err != nil {
		return nil, fmt.Errorf("读取CLI目录失败: %v", err)
	}

	// 找出要保留的版本
	keepVersions := make(map[string]bool)
	for i, version := range versions {
		if i < keepCount {
			keepVersions[version] = true
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "Stable-") || strings.HasSuffix(name, ".staging") {
			continue
		}

		// 检查是否在保留列表中
		if keepVersions[name] {
			continue
		}

		// 添加到待清理列表
		versionPath := filepath.Join(cliDir, name)
		size, _ := getDirSize(versionPath)
		if size > 0 {
			items = append(items, cleanupItem{
				Path:        versionPath,
				Size:        size,
				Description: fmt.Sprintf("CLI版本: %s", name),
			})
		}

		// 检查对应的.staging目录
		stagingPath := versionPath + ".staging"
		if _, err := os.Stat(stagingPath); err == nil {
			stagingSize, _ := getDirSize(stagingPath)
			if stagingSize > 0 {
				items = append(items, cleanupItem{
					Path:        stagingPath,
					Size:        stagingSize,
					Description: fmt.Sprintf("Staging: %s.staging", name),
				})
			}
		}
	}

	return items, nil
}

// analyzeCacheAndLogs 分析缓存和日志
func analyzeCacheAndLogs(vscodeDir string) []cleanupItem {
	var items []cleanupItem

	// 扩展安装包缓存
	cachePath := filepath.Join(vscodeDir, "data", "CachedExtensionVSIXs")
	if size, _ := getDirSize(cachePath); size > 0 {
		items = append(items, cleanupItem{
			Path:        cachePath,
			Size:        size,
			Description: "扩展安装包缓存",
		})
	}

	// 日志目录（>10MB）
	logsPath := filepath.Join(vscodeDir, "data", "logs")
	if size, _ := getDirSize(logsPath); size > 10*1024*1024 {
		items = append(items, cleanupItem{
			Path:        logsPath,
			Size:        size,
			Description: "日志目录",
		})
	}

	// CLI旧日志（>100KB）
	pattern := filepath.Join(vscodeDir, ".cli.*.log")
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if info.Size() > 100*1024 {
			items = append(items, cleanupItem{
				Path:        match,
				Size:        info.Size(),
				Description: fmt.Sprintf("CLI日志: %s", info.Name()),
			})
		}
	}

	return items
}

// analyzeTempFiles 分析临时文件
func analyzeTempFiles(vscodeDir string) []cleanupItem {
	var items []cleanupItem

	// 查找.tmp和.temp文件
	patterns := []string{
		filepath.Join(vscodeDir, "*.tmp"),
		filepath.Join(vscodeDir, "*.temp"),
		filepath.Join(vscodeDir, "*", "*.tmp"),
		filepath.Join(vscodeDir, "*", "*.temp"),
		filepath.Join(vscodeDir, "*", "*", "*.tmp"),
		filepath.Join(vscodeDir, "*", "*", "*.temp"),
	}

	seen := make(map[string]bool)
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if seen[match] {
				continue
			}
			seen[match] = true

			info, err := os.Stat(match)
			if err != nil || info.Size() == 0 {
				continue
			}

			relPath, _ := filepath.Rel(vscodeDir, match)
			items = append(items, cleanupItem{
				Path:        match,
				Size:        info.Size(),
				Description: fmt.Sprintf("临时文件: %s", relPath),
			})
		}
	}

	return items
}

// displayAnalysis 显示分析结果
func displayAnalysis(analysis *analysisResult) error {
	if len(analysis.Items) == 0 {
		fmt.Println("未发现需要清理的项目")
		return nil
	}

	// 分类显示
	fmt.Printf("发现 %d 个可清理项目:\n", len(analysis.Items))
	fmt.Println(strings.Repeat("-", 40))

	// 按类别分组
	groups := make(map[string][]cleanupItem)
	for _, item := range analysis.Items {
		if strings.Contains(item.Description, "CLI版本:") || strings.Contains(item.Description, "Staging:") {
			groups["CLI版本"] = append(groups["CLI版本"], item)
		} else if strings.Contains(item.Description, "扩展安装包") {
			groups["扩展缓存"] = append(groups["扩展缓存"], item)
		} else if strings.Contains(item.Description, "日志目录") {
			groups["日志文件"] = append(groups["日志文件"], item)
		} else if strings.Contains(item.Description, "CLI日志:") {
			groups["CLI日志"] = append(groups["CLI日志"], item)
		} else if strings.Contains(item.Description, "临时文件") {
			groups["临时文件"] = append(groups["临时文件"], item)
		} else {
			groups["其他"] = append(groups["其他"], item)
		}
	}

	// 按顺序显示类别
	categories := []string{"CLI版本", "扩展缓存", "日志文件", "CLI日志", "临时文件", "其他"}
	idx := 1
	for _, category := range categories {
		items, ok := groups[category]
		if !ok || len(items) == 0 {
			continue
		}

		fmt.Printf("\n%s:\n", category)
		for _, item := range items {
			fmt.Printf("  %2d. %s  %s\n", idx, formatSize(item.Size), item.Description)
			idx++
		}
	}

	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("总计可释放: %s\n", formatSize(analysis.TotalSize))

	return nil
}

// performCleanup 执行清理
func performCleanup(analysis *analysisResult, keepCount int) (int64, error) {
	var cleanedSize int64

	fmt.Println("\n开始清理...")
	fmt.Println(strings.Repeat("-", 40))

	for i, item := range analysis.Items {
		// 跳过LRU备份文件
		if strings.Contains(item.Path, "lru.json.backup") {
			continue
		}

		if err := os.RemoveAll(item.Path); err != nil {
			fmt.Printf("  ✗ %s (删除失败: %v)\n", item.Description, err)
		} else {
			fmt.Printf("  ✓ %s\n", item.Description)
			cleanedSize += item.Size
		}

		// 清理LRU文件（保留最新的keepCount个版本）
		if i == len(analysis.Items)-1 && strings.Contains(item.Description, "CLI版本") {
			if file, err := os.ReadFile(analysis.LRUPath); err == nil {
				var versions lruData
				if json.Unmarshal(file, &versions) == nil {
					if len(versions) > keepCount {
						newLRU := versions[:keepCount]
						if data, err := json.Marshal(newLRU); err == nil {
							// 备份原文件
							backupPath := analysis.LRUPath + ".backup"
							if err := os.WriteFile(backupPath, file, 0644); err != nil {
								fmt.Printf("  ⚠ 备份LRU文件失败: %v\n", err)
							}
							if err := os.WriteFile(analysis.LRUPath, data, 0644); err != nil {
								fmt.Printf("  ⚠ 更新LRU文件失败: %v\n", err)
							}
						}
					}
				}
			}
		}
	}

	return cleanedSize, nil
}

// getDirSize 获取目录大小
func getDirSize(path string) (int64, error) {
	var size int64

	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	if !info.IsDir() {
		return info.Size(), nil
	}

	if err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return size, nil
}

// formatSize 格式化文件大小
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	// 排序文件大小和单位，从大到小
	sizes := []struct {
		size int64
		unit string
	}{
		{GB, "GB"},
		{MB, "MB"},
		{KB, "KB"},
		{1, "bytes"},
	}

	for _, s := range sizes {
		if bytes >= s.size {
			if s.size == 1 {
				return strconv.FormatInt(bytes, 10) + " bytes"
			}
			value := float64(bytes) / float64(s.size)
			// 根据单位决定保留小数位数
			if s.unit == "GB" {
				return strconv.FormatFloat(value, 'f', 1, 64) + " GB"
			}
			return strconv.FormatFloat(value, 'f', 0, 64) + " " + s.unit
		}
	}

	return "0 bytes"
}

func init() {
	// 在root.go中添加vc命令
}
