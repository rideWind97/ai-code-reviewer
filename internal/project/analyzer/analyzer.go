package analyzer

import (
	"fmt"
	"sort"
	"strings"
)

// File 是项目快照中的一个文件。
type File struct {
	Path    string // 文件路径。
	Content string // 文件内容。
}

// Snapshot 是 Project Analyzer 的输入。
type Snapshot struct {
	Files []File // 仓库文件快照；MVP 不读取磁盘，方便测试和后续接 repo cache。
}

// Result 是项目分析结果。
type Result struct {
	Language    string   // 主要语言。
	EntryPoints []string // 入口文件，例如 cmd/*/main.go。
	TestFiles   int      // 测试文件数量。
	Packages    []string // Go package 目录列表。
}

// Analyze 生成项目分析摘要。
//
// 当前实现是确定性静态分析骨架，不调用模型；后续可以补充函数索引、复用目录和风险路径。
func Analyze(snapshot Snapshot) Result {
	result := Result{Language: "Unknown"}
	packages := make(map[string]struct{})
	for _, file := range snapshot.Files {
		if file.Path == "go.mod" {
			result.Language = "Go"
		}
		if strings.HasPrefix(file.Path, "cmd/") && strings.HasSuffix(file.Path, "/main.go") {
			result.EntryPoints = append(result.EntryPoints, file.Path)
		}
		if strings.HasSuffix(file.Path, "_test.go") {
			result.TestFiles++
		}
		if strings.HasSuffix(file.Path, ".go") {
			dir := directory(file.Path)
			packages[dir] = struct{}{}
		}
	}
	for pkg := range packages {
		result.Packages = append(result.Packages, pkg)
	}
	sort.Strings(result.EntryPoints)
	sort.Strings(result.Packages)
	return result
}

// RenderMarkdown 把分析结果渲染为项目分析文件片段。
func (r Result) RenderMarkdown() string {
	return fmt.Sprintf("# Project Analysis\n\n- Language: %s\n- Entry points: %s\n- Test files: %d\n- Packages: %s\n",
		r.Language,
		strings.Join(r.EntryPoints, ", "),
		r.TestFiles,
		strings.Join(r.Packages, ", "),
	)
}

// directory 返回文件所在目录。
func directory(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return "."
	}
	return path[:idx]
}
