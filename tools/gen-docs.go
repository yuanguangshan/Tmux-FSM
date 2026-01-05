package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

/*
====================================================
 Configuration & Globals
====================================================
*/

const versionStr = "v2.0.0"

// Config 集中管理配置
type Config struct {
	RootDir     string
	OutputFile  string
	IncludeExts []string
	ExcludeExts []string
	MaxFileSize int64
	NoSubdirs   bool
	Verbose     bool
	Version     bool
}

// FileMetadata 仅存储元数据，不存内容
type FileMetadata struct {
	RelPath  string
	FullPath string
	Size     int64
}

// Stats 统计信息
type Stats struct {
	FileCount int
	TotalSize int64
	Skipped   int
}

var defaultIgnorePatterns = []string{
	".git", ".idea", ".vscode",
	"node_modules", "vendor", "dist", "build", "target", "bin",
	"__pycache__", ".DS_Store",
	"package-lock.json", "yarn.lock", "go.sum",
}

// 语言映射表（全局配置，便于扩展）
var languageMap = map[string]string{
	".go":   "go",
	".js":   "javascript",
	".ts":   "typescript",
	".tsx":  "typescript",
	".jsx":  "javascript",
	".py":   "python",
	".java": "java",
	".c":    "c",
	".cpp":  "cpp",
	".cc":   "cpp",
	".cxx":  "cpp",
	".h":    "c",
	".hpp":  "cpp",
	".rs":   "rust",
	".rb":   "ruby",
	".php":  "php",
	".cs":   "csharp",
	".swift": "swift",
	".kt":   "kotlin",
	".scala": "scala",
	".r":    "r",
	".sql":  "sql",
	".sh":   "bash",
	".bash": "bash",
	".zsh":  "bash",
	".fish": "fish",
	".ps1":  "powershell",
	".md":   "markdown",
	".html": "html",
	".htm":  "html",
	".css":  "css",
	".scss": "scss",
	".sass": "sass",
	".less": "less",
	".xml":  "xml",
	".json": "json",
	".yaml": "yaml",
	".yml":  "yaml",
	".toml": "toml",
	".ini":  "ini",
	".conf": "conf",
	".txt":  "text",
}

/*
====================================================
 Main Entry
====================================================
*/

func main() {
	cfg := parseFlags()
	printStartupInfo(cfg)

	// Phase 1: 扫描文件结构
	fmt.Println("⏳ 正在扫描文件结构...")
	files, stats, err := scanDirectory(cfg)
	if err != nil {
		fmt.Printf("❌ 扫描失败: %v\n", err)
		os.Exit(1)
	}

	// Phase 2: 流式写入
	fmt.Printf("💾 正在写入文档 [文件数: %d]...\n", len(files))
	if err := writeMarkdownStream(cfg, files, stats); err != nil {
		fmt.Printf("❌ 写入失败: %v\n", err)
		os.Exit(1)
	}

	printSummary(stats, cfg.OutputFile)
}

/*
====================================================
 Flag Parsing
====================================================
*/

func parseFlags() Config {
	var cfg Config
	var include, exclude string
	var maxKB int64

	flag.StringVar(&cfg.RootDir, "dir", ".", "Root directory to scan")
	flag.StringVar(&cfg.OutputFile, "o", "", "Output markdown file")
	flag.StringVar(&include, "i", "", "Include extensions (e.g. .go,.js)")
	flag.StringVar(&exclude, "x", "", "Exclude extensions")
	flag.Int64Var(&maxKB, "max-size", 500, "Max file size in KB")
	flag.BoolVar(&cfg.NoSubdirs, "no-subdirs", false, "Do not scan subdirectories")
	flag.BoolVar(&cfg.NoSubdirs, "ns", false, "Alias for --no-subdirs")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&cfg.Version, "version", false, "Show version")

	flag.Parse()

	if cfg.Version {
		fmt.Printf("gen-docs %s\n", versionStr)
		os.Exit(0)
	}

	// 支持位置参数
	if args := flag.Args(); len(args) > 0 {
		cfg.RootDir = args[0]
	}

	// 自动生成输出文件名
	if cfg.OutputFile == "" {
		base := filepath.Base(cfg.RootDir)
		if base == "." || base == string(filepath.Separator) {
			base = "project"
		}
		date := time.Now().Format("20060102")
		cfg.OutputFile = fmt.Sprintf("%s-%s-docs.md", base, date)
	}

	cfg.IncludeExts = normalizeExts(include)
	cfg.ExcludeExts = normalizeExts(exclude)
	cfg.MaxFileSize = maxKB * 1024

	return cfg
}

/*
====================================================
 Startup & Summary
====================================================
*/

func printStartupInfo(cfg Config) {
	fmt.Println("▶ Gen-Docs Started")
	fmt.Printf("  Root: %s\n", cfg.RootDir)
	fmt.Printf("  Out : %s\n", cfg.OutputFile)
	fmt.Printf("  Max : %d KB\n", cfg.MaxFileSize/1024)
	if len(cfg.IncludeExts) > 0 {
		fmt.Printf("  Only: %v\n", cfg.IncludeExts)
	}
	if len(cfg.ExcludeExts) > 0 {
		fmt.Printf("  Skip: %v\n", cfg.ExcludeExts)
	}
	fmt.Println()
}

func printSummary(stats Stats, output string) {
	fmt.Println("\n✔ 完成!")
	fmt.Printf("  文件数  : %d\n", stats.FileCount)
	fmt.Printf("  已跳过  : %d\n", stats.Skipped)
	fmt.Printf("  总大小  : %.2f KB\n", float64(stats.TotalSize)/1024)
	fmt.Printf("  输出路径: %s\n", output)
}

/*
====================================================
 Directory Scanning
====================================================
*/

func scanDirectory(cfg Config) ([]FileMetadata, Stats, error) {
	var files []FileMetadata
	var stats Stats

	absOutput, _ := filepath.Abs(cfg.OutputFile)

	err := filepath.WalkDir(cfg.RootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logf(cfg.Verbose, "⚠ 无法访问: %s", path)
			stats.Skipped++
			return nil
		}

		relPath, _ := filepath.Rel(cfg.RootDir, path)
		if relPath == "." {
			return nil
		}

		// 处理目录
		if d.IsDir() {
			if cfg.NoSubdirs && relPath != "." {
				return filepath.SkipDir
			}
			if shouldIgnoreDir(d.Name()) {
				logf(cfg.Verbose, "⊘ 跳过目录: %s", relPath)
				return filepath.SkipDir
			}
			return nil
		}

		// 排除输出文件自身
		if absPath, _ := filepath.Abs(path); absPath == absOutput {
			return nil
		}

		// 获取文件信息
		info, err := d.Info()
		if err != nil {
			return nil
		}

		// 应用过滤规则
		if shouldIgnoreFile(relPath, info.Size(), cfg) {
			stats.Skipped++
			return nil
		}

		// 二进制检测
		if isBinaryFile(path) {
			logf(cfg.Verbose, "⊘ 二进制文件: %s", relPath)
			stats.Skipped++
			return nil
		}

		// 加入列表
		files = append(files, FileMetadata{
			RelPath:  relPath,
			FullPath: path,
			Size:     info.Size(),
		})
		stats.FileCount++
		stats.TotalSize += info.Size()

		logf(cfg.Verbose, "✓ 添加: %s", relPath)

		return nil
	})

	// 排序保证输出一致性
	sort.Slice(files, func(i, j int) bool {
		return files[i].RelPath < files[j].RelPath
	})

	return files, stats, err
}

/*
====================================================
 Ignore Rules
====================================================
*/

func shouldIgnoreDir(name string) bool {
	if strings.HasPrefix(name, ".") && name != "." {
		return true
	}
	for _, pattern := range defaultIgnorePatterns {
		if name == pattern {
			return true
		}
	}
	return false
}

func shouldIgnoreFile(relPath string, size int64, cfg Config) bool {
	// 大小限制
	if size > cfg.MaxFileSize {
		logf(cfg.Verbose, "⊘ 文件过大: %s", relPath)
		return true
	}

	ext := strings.ToLower(filepath.Ext(relPath))

	// 排除规则优先
	for _, e := range cfg.ExcludeExts {
		if ext == e {
			return true
		}
	}

	// 包含规则（白名单）
	if len(cfg.IncludeExts) > 0 {
		found := false
		for _, i := range cfg.IncludeExts {
			if ext == i {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// 路径包含忽略模式
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range parts {
		for _, pattern := range defaultIgnorePatterns {
			if part == pattern {
				return true
			}
		}
	}

	return false
}

/*
====================================================
 File Utilities
====================================================
*/

func normalizeExts(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var exts []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		exts = append(exts, p)
	}
	return exts
}

func isBinaryFile(path string) bool {
	// 快速路径：压缩文件
	if strings.Contains(path, ".min.") {
		return true
	}

	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()

	// 只读前 512 字节
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	buf = buf[:n]

	// NULL 字节检测
	for _, b := range buf {
		if b == 0 {
			return true
		}
	}

	// UTF-8 有效性检测
	return !utf8.Valid(buf)
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if lang, ok := languageMap[ext]; ok {
		return lang
	}
	return "text"
}

/*
====================================================
 Markdown Output
====================================================
*/

func writeMarkdownStream(cfg Config, files []FileMetadata, stats Stats) error {
	f, err := os.Create(cfg.OutputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriterSize(f, 64*1024)

	// 写入头部
	fmt.Fprintln(w, "# Project Documentation")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- **Generated at:** %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "- **Root Dir:** `%s`\n", cfg.RootDir)
	fmt.Fprintf(w, "- **File Count:** %d\n", stats.FileCount)
	fmt.Fprintf(w, "- **Total Size:** %.2f KB\n", float64(stats.TotalSize)/1024)
	fmt.Fprintln(w)

	// 写入目录
	fmt.Fprintln(w, "## 📂 File List")
	for _, file := range files {
		fmt.Fprintf(w, "- `%s` (%.2f KB)\n", file.RelPath, float64(file.Size)/1024)
	}
	fmt.Fprintln(w, "\n---")

	// 流式写入文件内容
	total := len(files)
	for i, file := range files {
		if !cfg.Verbose && (i%10 == 0 || i == total-1) {
			fmt.Printf("\r🚀 进度: %d/%d (%.1f%%)", i+1, total, float64(i+1)/float64(total)*100)
		}

		if err := copyFileContent(w, file); err != nil {
			logf(true, "\n⚠ 读取失败 %s: %v", file.RelPath, err)
			continue
		}
	}
	fmt.Println()

	// 【改进1】显式 Flush 并捕获错误
	return w.Flush()
}

func copyFileContent(w *bufio.Writer, file FileMetadata) error {
	src, err := os.Open(file.FullPath)
	if err != nil {
		return err
	}
	defer src.Close()

	lang := detectLanguage(file.RelPath)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "## 📄 `%s`\n\n", file.RelPath)
	
	// 【改进2】使用更安全的代码块分隔符（4个反引号）
	// 这样即使源代码中包含 ``` 也不会破坏格式
	fmt.Fprintf(w, "````%s\n", lang)

	if _, err := io.Copy(w, src); err != nil {
		return err
	}

	fmt.Fprintln(w, "\n````")
	return nil
}

/*
====================================================
 Logging
====================================================
*/

func logf(verbose bool, format string, a ...any) {
	if verbose {
		fmt.Printf(format+"\n", a...)
	}
}
