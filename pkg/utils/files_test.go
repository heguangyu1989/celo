package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir := t.TempDir()
	
	// 测试用例1: 文件存在的情况
	existingFile := filepath.Join(tempDir, "existing_file.txt")
	file, err := os.Create(existingFile)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	file.Close()
	
	if !FileExists(existingFile) {
		t.Errorf("FileExists(%q) 返回 false，期望返回 true", existingFile)
	}
	
	// 测试用例2: 文件不存在的情况
	nonExistingFile := filepath.Join(tempDir, "non_existing_file.txt")
	if FileExists(nonExistingFile) {
		t.Errorf("FileExists(%q) 返回 true，期望返回 false", nonExistingFile)
	}
	
	// 测试用例3: 路径是目录的情况
	if FileExists(tempDir) {
		t.Errorf("FileExists(%q) 返回 true，期望返回 false（因为是目录）", tempDir)
	}
	
	// 测试用例4: 空路径
	if FileExists("") {
		t.Errorf("FileExists(\"\") 返回 true，期望返回 false")
	}
}