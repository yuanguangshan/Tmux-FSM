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
	RootDir        string
	OutputFile     string
	IncludeExts    []string
	IncludeMatches []string
	ExcludeExts    []string
	ExcludeMatches []string
	MaxFileSize    int64
	NoSubdirs      bool
	Verbose        bool
	Version        bool
	ShowStats      bool
}

// FileMetadata 仅存储元数据，不存内容
type FileMetadata struct {
	RelPath   string
	FullPath  string
	Size      int64
	LineCount int
}

// Stats 统计信息
type Stats struct {
	PotentialMatches   int // 符合包含规则的文件数
	ExplicitlyExcluded int // 符合包含规则但被排除规则踢掉的文件数
	FileCount          int // 最终写入的文件数
	TotalSize          int64
	TotalLines         int
	Skipped            int // 完全不匹配规则的文件数
	DirCount           int // 文件夹数量
}

// DirStats 目录统计信息
type DirStats struct {
	Path      string
	FileCount int
	TotalSize int64
	TotalLines int
}

// ExtStats 文件类型统计信息
type ExtStats struct {
	Ext       string
	FileCount int
	TotalSize int64
}

var defaultIgnorePatterns = []string{
	".git", ".idea", ".vscode", ".svn", ".hg",
	"node_modules", "vendor", "dist", "build", "target", "bin", "out", "release", "debug",
	"__pycache__", ".pytest_cache", ".tox", ".coverage", "coverage.xml",
	".DS_Store", ".env", ".venv", "venv", "env",
	"package-lock.json", "yarn.lock", "go.sum", "composer.lock", "Gemfile.lock",
	"*.log", "*.tmp", "*.temp", "*.cache", "*.swp", "*.swo",
	"tags", "TAGS", "*.pid", "*.seed", "*.idx",
	"Pods", "Carthage", "CocoaPods", ".xcassets",
	"obj", "ipch", "*.user", "*.userosscache", "*.sln.docstates",
	"*.VC.db", "*.VC.VC.opendb", "Debug", "Release", "x64", "x86", "arm64",
	"*.aps", "*.ncb", "*.opendb", "*.opensdf", "*.sdf", "*.cachefile", "*.VC.VC.opendb",
	"cmake-build-*", ".gradle", "build", ".sonar", ".scannerwork",
	"*.tgz", "*.tar.gz", "*.zip", "*.rar", "*.7z",
	"logs", "tmp", "temp", "cache", ".history", ".nyc_output",
}

// 语言映射表（全局配置，便于扩展）
var languageMap = map[string]string{
	".go":    "go",
	".js":    "javascript",
	".ts":    "typescript",
	".tsx":   "typescript",
	".jsx":   "javascript",
	".py":    "python",
	".java":  "java",
	".c":     "c",
	".cpp":   "cpp",
	".cc":    "cpp",
	".cxx":   "cpp",
	".h":     "c",
	".hpp":   "cpp",
	".rs":    "rust",
	".rb":    "ruby",
	".php":   "php",
	".cs":    "csharp",
	".swift": "swift",
	".kt":    "kotlin",
	".scala": "scala",
	".r":     "r",
	".sql":   "sql",
	".sh":    "bash",
	".bash":  "bash",
	".zsh":   "bash",
	".fish":  "fish",
	".ps1":   "powershell",
	".md":    "markdown",
	".html":  "html",
	".htm":   "html",
	".css":   "css",
	".scss":  "scss",
	".sass":  "sass",
	".less":  "less",
	".xml":   "xml",
	".json":  "json",
	".yaml":  "yaml",
	".yml":   "yaml",
	".toml":  "toml",
	".ini":   "ini",
	".conf":  "conf",
	".txt":   "text",
}

/*
====================================================
 Main Entry
====================================================
*/

func main() {
	cfg := parseFlags()
	
	// 如果是统计模式，执行统计并退出
	if cfg.ShowStats {
		if err := showProjectStats(cfg); err != nil {
			fmt.Printf("❌ 统计失败: %v\n", err)
			os.Exit(1)
		}
		return
	}
	
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
	var include, match, exclude, excludeMatch string
	var maxKB int64

	flag.StringVar(&cfg.RootDir, "dir", ".", "Root directory to scan")
	flag.StringVar(&cfg.OutputFile, "o", "", "Output markdown file")
	flag.StringVar(&include, "i", "", "Include extensions (e.g. .go,.js)")
	flag.StringVar(&match, "m", "", "Include path keywords (e.g. _test.go)")
	flag.StringVar(&exclude, "x", "", "Exclude extensions (e.g. .exe,.o)")
	flag.StringVar(&excludeMatch, "xm", "", "Exclude path keywords (e.g. vendor/,node_modules/)")
	flag.Int64Var(&maxKB, "max-size", 500, "Max file size in KB")
	flag.BoolVar(&cfg.NoSubdirs, "no-subdirs", false, "Do not scan subdirectories")
	flag.BoolVar(&cfg.NoSubdirs, "ns", false, "Alias for --no-subdirs")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&cfg.Version, "version", false, "Show version")
	flag.BoolVar(&cfg.ShowStats, "s", false, "Show project statistics")

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
		baseName := "project"
		cleanRoot := filepath.Clean(cfg.RootDir)

		if cleanRoot == "." || cleanRoot == string(filepath.Separator) {
			// 如果是当前目录，尝试获取文件夹真实名称
			if abs, err := filepath.Abs(cleanRoot); err == nil {
				baseName = filepath.Base(abs)
			}
		} else {
			// 将路径中的分隔符和点替换为下划线
			baseName = cleanRoot
			baseName = strings.ReplaceAll(baseName, string(filepath.Separator), "_")
			baseName = strings.ReplaceAll(baseName, ".", "_")
			// 清理连续的下划线
			for strings.Contains(baseName, "__") {
				baseName = strings.ReplaceAll(baseName, "__", "_")
			}
			baseName = strings.Trim(baseName, "_")
		}

		date := time.Now().Format("20060102")
		cfg.OutputFile = fmt.Sprintf("%s-%s-docs.md", baseName, date)
	}

	cfg.IncludeExts = normalizeExts(include)
	cfg.IncludeMatches = splitAndTrim(match)
	cfg.ExcludeExts = normalizeExts(exclude)
	cfg.ExcludeMatches = splitAndTrim(excludeMatch)

	// 从配置文件加载额外的忽略规则
	additionalExcludes, additionalExcludeMatches := loadIgnoreFile(cfg.RootDir)
	cfg.ExcludeExts = mergeStringSlices(cfg.ExcludeExts, additionalExcludes)
	cfg.ExcludeMatches = mergeStringSlices(cfg.ExcludeMatches, additionalExcludeMatches)

	cfg.MaxFileSize = maxKB * 1024

	return cfg
}

func splitAndTrim(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// 从配置文件加载忽略规则
func loadIgnoreFile(rootDir string) ([]string, []string) {
	var excludeExts []string
	var excludeMatches []string

	// 尝试多个可能的配置文件名
	possibleFiles := []string{".gen-docs-ignore", ".gdocsignore", ".docs-ignore"}

	for _, filename := range possibleFiles {
		configPath := filepath.Join(rootDir, filename)

		// 检查文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		// 读取配置文件
		content, err := os.ReadFile(configPath)
		if err != nil {
			logf(true, "⚠ 无法读取忽略配置文件 %s: %v", configPath, err)
			continue
		}

		logf(true, "✓ 发现忽略配置文件: %s", configPath)

		// 解析配置文件内容
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			// 跳过空行和注释行
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// 根据行的内容判断是扩展名还是路径匹配
			if strings.HasPrefix(line, ".") {
				// 这是一个扩展名（例如 .log, .tmp）
				excludeExts = append(excludeExts, strings.ToLower(line))
			} else {
				// 这是一个路径匹配模式（例如 vendor/, node_modules/）
				excludeMatches = append(excludeMatches, line)
			}
		}

		if err := scanner.Err(); err != nil {
			logf(true, "⚠ 读取忽略配置文件时出错 %s: %v", configPath, err)
		}

		// 找到并成功解析了一个配置文件，跳出循环
		break
	}

	return excludeExts, excludeMatches
}

// 合并两个字符串切片，避免重复
func mergeStringSlices(base, additional []string) []string {
	// 使用 map 来跟踪已存在的元素，避免重复
	seen := make(map[string]bool)
	var result []string

	// 先添加基础切片中的元素
	for _, item := range base {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	// 再添加附加切片中的元素
	for _, item := range additional {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
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
		fmt.Printf("  Only Ext: %v\n", cfg.IncludeExts)
	}
	if len(cfg.IncludeMatches) > 0 {
		fmt.Printf("  Match   : %v\n", cfg.IncludeMatches)
	}
	if len(cfg.ExcludeExts) > 0 {
		fmt.Printf("  Skip Ext: %v\n", cfg.ExcludeExts)
	}
	if len(cfg.ExcludeMatches) > 0 {
		fmt.Printf("  Skip Key: %v\n", cfg.ExcludeMatches)
	}
	fmt.Println()
}

func printSummary(stats Stats, output string) {
	fmt.Println("\n✔ 完成!")
	fmt.Printf("  符合包含规则 (Potential) : %d\n", stats.PotentialMatches)
	fmt.Printf("  由于排除规则被踢除 (Excluded): %d\n", stats.ExplicitlyExcluded)
	fmt.Printf("  最终写入文件数 (Final)    : %d\n", stats.FileCount)
	fmt.Printf("  总行数 (Total Lines)      : %d\n", stats.TotalLines)
	fmt.Printf("  总物理大小 (Total Size)   : %.2f KB\n", float64(stats.TotalSize)/1024)
	fmt.Printf("  无需处理的无关文件          : %d\n", stats.Skipped)
	fmt.Printf("  输出路径                  : %s\n", output)
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

		// --- 细化过滤逻辑 ---
		// 1. 基础过滤：过大或二进制
		if info.Size() > cfg.MaxFileSize || isBinaryFile(path) {
			stats.Skipped++
			return nil
		}

		// 2. 检查是否符合“包含”意图
		isIncluded := true
		if len(cfg.IncludeExts) > 0 || len(cfg.IncludeMatches) > 0 {
			extMatched := false
			if len(cfg.IncludeExts) > 0 {
				ext := strings.ToLower(filepath.Ext(relPath))
				for _, e := range cfg.IncludeExts {
					if ext == e {
						extMatched = true
						break
					}
				}
			} else {
				extMatched = true // 如果没设后缀白名单，默认后缀通过
			}

			pathMatched := false
			if len(cfg.IncludeMatches) > 0 {
				for _, m := range cfg.IncludeMatches {
					if strings.Contains(relPath, m) {
						pathMatched = true
						break
					}
				}
			} else {
				pathMatched = true // 如果没设关键字匹配，默认路径通过
			}
			isIncluded = extMatched && pathMatched
		}

		if !isIncluded {
			stats.Skipped++
			return nil
		}

		// 3. 符合包含意图 (Potential Match)
		stats.PotentialMatches++

		// 4. 检查是否被“排除”规则拦截
		isExcluded := false
		ext := strings.ToLower(filepath.Ext(relPath))
		for _, e := range cfg.ExcludeExts {
			if ext == e {
				isExcluded = true
				break
			}
		}
		if !isExcluded && len(cfg.ExcludeMatches) > 0 {
			for _, m := range cfg.ExcludeMatches {
				if strings.Contains(relPath, m) {
					isExcluded = true
					break
				}
			}
		}

		if isExcluded {
			stats.ExplicitlyExcluded++
			return nil
		}

		// --- 最终通过 ---
		lineCount, _ := countLines(path)
		files = append(files, FileMetadata{
			RelPath:   relPath,
			FullPath:  path,
			Size:      info.Size(),
			LineCount: lineCount,
		})
		stats.FileCount++
		stats.TotalLines += lineCount
		stats.TotalSize += info.Size()

		logf(cfg.Verbose, "✓ 添加: %s (%d lines)", relPath, lineCount)
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

	// 规则 0: 硬性排除 (关键字排除) - 优先级最高
	if len(cfg.ExcludeMatches) > 0 {
		for _, m := range cfg.ExcludeMatches {
			if strings.Contains(relPath, m) {
				logf(cfg.Verbose, "⊘ 匹配排除关键字 [%s]: %s", m, relPath)
				return true
			}
		}
	}

	// 规则 1: 包含后缀白名单
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

	// 规则 2: 关键字包含匹配
	if len(cfg.IncludeMatches) > 0 {
		found := false
		for _, m := range cfg.IncludeMatches {
			if strings.Contains(relPath, m) {
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
	fmt.Fprintln(w, "<a name=\"toc\"></a>")
	fmt.Fprintln(w, "## 📂 扫描目录")
	for _, file := range files {
		// 生成锚点，方便在 Markdown 中点击跳转
		// 注意：锚点名称在 GitHub 中通常是将空格转为横杠并全小写
		anchor := strings.ReplaceAll(file.RelPath, " ", "-")
		anchor = strings.ReplaceAll(anchor, ".", "")
		anchor = strings.ReplaceAll(anchor, "/", "")
		anchor = strings.ToLower(anchor)

		fmt.Fprintf(w, "- [%s](#📄-%s) (%d lines, %.2f KB)\n", file.RelPath, anchor, file.LineCount, float64(file.Size)/1024)
	}
	fmt.Fprintln(w, "\n---")

	// 流式写入文件内容
	total := len(files)
	for i, file := range files {
		if !cfg.Verbose && (i%10 == 0 || i == total-1) {
			fmt.Printf("\r🚀 写入进度: %d/%d (%.1f%%)", i+1, total, float64(i+1)/float64(total)*100)
		}

		if err := copyFileContent(w, file); err != nil {
			logf(true, "\n⚠ 读取失败 %s: %v", file.RelPath, err)
			continue
		}
	}
	fmt.Println()

	//【补充统计】
	fmt.Fprintln(w, "\n---")
	fmt.Fprintf(w, "### 📊 最终统计汇总\n")
	fmt.Fprintf(w, "- **文件总数:** %d\n", stats.FileCount)
	fmt.Fprintf(w, "- **代码总行数:** %d\n", stats.TotalLines)
	fmt.Fprintf(w, "- **物理总大小:** %.2f KB\n", float64(stats.TotalSize)/1024)

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
	fmt.Fprintf(w, "## 📄 %s\n\n", file.RelPath)
	fmt.Fprintf(w, "````%s\n", lang)

	// 使用 io.Copy 替代 scanner，更安全且不限行长
	if _, err := io.Copy(w, src); err != nil {
		return err
	}

	fmt.Fprintln(w, "\n````")
	fmt.Fprintln(w, "\n[⬆ 回到目录](#toc)")
	return nil
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	// 增加缓冲区以支持超长行
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
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

/*
====================================================
 Project Statistics
====================================================
*/

func showProjectStats(cfg Config) error {
	fmt.Println("📊 正在统计项目信息...")
	fmt.Printf("  Root: %s\n\n", cfg.RootDir)
	
	var files []FileMetadata
	dirMap := make(map[string]*DirStats)
	extMap := make(map[string]*ExtStats)
	var stats Stats
	absOutput, _ := filepath.Abs(cfg.OutputFile)
	
	err := filepath.WalkDir(cfg.RootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		
		relPath, _ := filepath.Rel(cfg.RootDir, path)
		if relPath == "." {
			return nil
		}
		
		// 处理目录
		if d.IsDir() {
			if shouldIgnoreDir(d.Name()) {
				return filepath.SkipDir
			}
			stats.DirCount++
			dirMap[relPath] = &DirStats{Path: relPath}
			return nil
		}
		
		// 排除输出文件
		if absPath, _ := filepath.Abs(path); absPath == absOutput {
			return nil
		}
		
		info, err := d.Info()
		if err != nil {
			return nil
		}
		
		// 过滤二进制和过大文件
		if info.Size() > cfg.MaxFileSize || isBinaryFile(path) {
			return nil
		}
		
		lineCount, _ := countLines(path)
		fileSize := info.Size()
		
		// 统计文件
		files = append(files, FileMetadata{
			RelPath:   relPath,
			FullPath:  path,
			Size:      fileSize,
			LineCount: lineCount,
		})
		stats.FileCount++
		stats.TotalLines += lineCount
		stats.TotalSize += fileSize
		
		// 统计目录
		dir := filepath.Dir(relPath)
		if dir == "." {
			dir = "."
		}
		if dirStats, ok := dirMap[dir]; ok {
			dirStats.FileCount++
			dirStats.TotalSize += fileSize
			dirStats.TotalLines += lineCount
		} else {
			dirMap[dir] = &DirStats{
				Path:       dir,
				FileCount:  1,
				TotalSize:  fileSize,
				TotalLines: lineCount,
			}
		}
		
		// 统计文件类型
		ext := strings.ToLower(filepath.Ext(relPath))
		if ext == "" {
			ext = "(no extension)"
		}
		if extStats, ok := extMap[ext]; ok {
			extStats.FileCount++
			extStats.TotalSize += fileSize
		} else {
			extMap[ext] = &ExtStats{
				Ext:       ext,
				FileCount: 1,
				TotalSize: fileSize,
			}
		}
		
		return nil
	})
	
	if err != nil {
		return err
	}
	
	// 输出统计结果
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("📁 基本统计")
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Printf("  文件夹数量: %d\n", stats.DirCount)
	fmt.Printf("  文件数量  : %d\n", stats.FileCount)
	fmt.Printf("  总行数    : %d\n", stats.TotalLines)
	fmt.Printf("  总大小    : %.2f KB (%.2f MB)\n", 
		float64(stats.TotalSize)/1024, float64(stats.TotalSize)/1024/1024)
	
	// Top 5 最大文件夹
	fmt.Println("\n" + "=" + strings.Repeat("=", 70))
	fmt.Println("📂 Top 5 最大文件夹")
	fmt.Println("=" + strings.Repeat("=", 70))
	
	var dirList []DirStats
	for _, ds := range dirMap {
		if ds.FileCount > 0 {
			dirList = append(dirList, *ds)
		}
	}
	sort.Slice(dirList, func(i, j int) bool {
		return dirList[i].TotalSize > dirList[j].TotalSize
	})
	
	for i := 0; i < 5 && i < len(dirList); i++ {
		ds := dirList[i]
		sizePercent := float64(ds.TotalSize) / float64(stats.TotalSize) * 100
		linesPercent := float64(ds.TotalLines) / float64(stats.TotalLines) * 100
		fmt.Printf("  %d. %s\n", i+1, ds.Path)
		fmt.Printf("     大小: %.2f KB (%.1f%%), 行数: %d (%.1f%%), 文件数: %d\n",
			float64(ds.TotalSize)/1024, sizePercent, ds.TotalLines, linesPercent, ds.FileCount)
	}
	
	// Top 5 最大文件
	fmt.Println("\n" + "=" + strings.Repeat("=", 70))
	fmt.Println("📄 Top 5 最大文件")
	fmt.Println("=" + strings.Repeat("=", 70))
	
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})
	
	for i := 0; i < 5 && i < len(files); i++ {
		f := files[i]
		sizePercent := float64(f.Size) / float64(stats.TotalSize) * 100
		linesPercent := float64(f.LineCount) / float64(stats.TotalLines) * 100
		fmt.Printf("  %d. %s\n", i+1, f.RelPath)
		fmt.Printf("     大小: %.2f KB (%.1f%%), 行数: %d (%.1f%%)\n",
			float64(f.Size)/1024, sizePercent, f.LineCount, linesPercent)
	}
	
	// 按文件类型统计
	fmt.Println("\n" + "=" + strings.Repeat("=", 70))
	fmt.Println("📊 按文件类型统计")
	fmt.Println("=" + strings.Repeat("=", 70))
	
	var extList []ExtStats
	for _, es := range extMap {
		extList = append(extList, *es)
	}
	sort.Slice(extList, func(i, j int) bool {
		return extList[i].TotalSize > extList[j].TotalSize
	})
	
	fmt.Printf("  %-20s %10s %15s %10s\n", "类型", "文件数", "总大小", "占比")
	fmt.Println("  " + strings.Repeat("-", 68))
	for _, es := range extList {
		sizePercent := float64(es.TotalSize) / float64(stats.TotalSize) * 100
		fmt.Printf("  %-20s %10d %12.2f KB %9.1f%%\n",
			es.Ext, es.FileCount, float64(es.TotalSize)/1024, sizePercent)
	}
	
	fmt.Println("\n" + "=" + strings.Repeat("=", 70))
	fmt.Println("✅ 统计完成!")
	fmt.Println("=" + strings.Repeat("=", 70))
	
	return nil
}
