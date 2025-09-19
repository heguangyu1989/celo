# Celo 快速开始指南

## 项目简介

Celo 是一个高效的命令行工具，专注于提升开发效率。它提供多种实用功能，包括 MD5 校验、GitLab 合并请求创建、构建信息查看等。

## 系统要求

- Go 1.25.1 或更高版本
- Git（用于GitLab功能）

## 安装

### 从源码编译

```bash
git clone https://github.com/heguangyu1989/celo.git
cd celo
go build -o celo .
```

### 安装到系统

```bash
# 将编译后的二进制文件移动到PATH中的目录
sudo mv celo /usr/local/bin/
```

## 配置

创建配置文件 `~/.celo.yaml` 或 `~/.celo.json`，包含GitLab访问令牌：

```yaml
# ~/.celo.yaml
gitlab_token: "your_gitlab_personal_access_token"
```

## 命令使用

### 基本命令

```bash
# 查看帮助
celo --help

# 查看构建信息
celo info
```

### MD5 校验

```bash
# 基本用法
celo md5 file1.txt file2.txt

# 输出为JSON格式
celo md5 file1.txt file2.txt --output json

# 输出为YAML格式
celo md5 file1.txt file2.txt --output yaml

# 输出为表格格式
celo md5 file1.txt file2.txt --output table

# 计算字符串MD5
celo md5 "hello world"
```

支持三种类型的输入：
- `string`: 直接计算字符串的MD5
- `file`: 计算文件的MD5

### GitLab 合并请求创建

```bash
celo merge \
  --src feature-branch \
  --dst main \
  --title "新功能: 添加用户管理模块" \
  --tags bugfix,feature,enhancement
```

**参数说明：**
- `--src`: 源分支名称（必需）
- `--dst`: 目标分支名称（必需）
- `--title`: 合并请求标题（必需）
- `--tags`: 标签数组（可选）

### 配置文件选项

```bash
# 指定配置文件路径
celo --config /path/to/config.yaml md5 file.txt

# 使用配置文件
celo merge --src dev --dst main --title "测试" --tags bug
```



## GitLab 设置

### 获取 Personal Access Token

1. 登录您的GitLab账户
2. 进入 Settings → Access Tokens
3. 创建新令牌，确保具有 `api` 权限
4. 复制令牌并配置到配置文件中

### 自动检测GitLab项目

工具会自动从Git URL中提取GitLab项目ID，支持以下格式：
- `git@gitlab.com:user/project.git`
- `https://gitlab.com/user/project.git`