package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	// 解析命令行参数
	var (
		dir      string
		pattern  string
		replace  string
		prefix   string
		suffix   string
		startNum int
		dryRun   bool
		ext      string
	)

	flag.StringVar(&dir, "dir", ".", "目标目录路径")
	flag.StringVar(&pattern, "pattern", "", "匹配模式（正则表达式）")
	flag.StringVar(&replace, "replace", "", "替换字符串")
	flag.StringVar(&prefix, "prefix", "", "添加前缀")
	flag.StringVar(&suffix, "suffix", "", "添加后缀")
	flag.IntVar(&startNum, "start", 1, "起始编号")
	flag.BoolVar(&dryRun, "dry-run", false, "试运行，不实际重命名")
	flag.StringVar(&ext, "ext", ".docx", "文件扩展名（支持 .doc .docx）")
	flag.Parse()

	// 支持的Word文档扩展名
	wordExts := []string{".docx", ".doc"}
	if ext != "" && !contains(wordExts, strings.ToLower(ext)) {
		fmt.Printf("警告：扩展名 %s 可能不是标准Word文档格式\n", ext)
	}

	fmt.Printf("开始处理目录: %s\n", dir)

	// 获取Word文档文件
	files, err := getWordFiles(dir, wordExts)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("未找到Word文档")
		return
	}

	fmt.Printf("找到 %d 个Word文档:\n", len(files))

	// 批量重命名
	renamedCount := 0
	for i, oldPath := range files {
		newName := generateNewName(filepath.Base(oldPath), pattern, replace, prefix, suffix, startNum+i)

		// 保持原扩展名
		ext := filepath.Ext(oldPath)
		newName = strings.TrimSuffix(newName, filepath.Ext(newName)) + ext

		newPath := filepath.Join(filepath.Dir(oldPath), newName)

		// 检查新文件名是否已存在
		if _, err := os.Stat(newPath); err == nil {
			fmt.Printf("⚠️  跳过: %s -> %s (文件已存在)\n",
				filepath.Base(oldPath), newName)
			continue
		}

		if oldPath == newPath {
			fmt.Printf("✓ 保持: %s\n", filepath.Base(oldPath))
			continue
		}

		if dryRun {
			fmt.Printf("📋 预览: %s -> %s\n",
				filepath.Base(oldPath), newName)
		} else {
			err := os.Rename(oldPath, newPath)
			if err != nil {
				fmt.Printf("❌ 错误重命名 %s: %v\n", oldPath, err)
			} else {
				fmt.Printf("✅ 重命名: %s -> %s\n",
					filepath.Base(oldPath), newName)
				renamedCount++
			}
		}
	}

	if dryRun {
		fmt.Printf("\n📋 试运行完成，将重命名 %d 个文件\n", len(files))
	} else {
		fmt.Printf("\n🎉 完成! 成功重命名 %d/%d 个文件\n", renamedCount, len(files))
	}
}

// 获取Word文档文件
func getWordFiles(dir string, exts []string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != dir {
			return filepath.SkipDir // 只处理当前目录，不递归子目录
		}

		if !info.IsDir() {
			fileExt := strings.ToLower(filepath.Ext(path))
			for _, ext := range exts {
				if fileExt == ext {
					files = append(files, path)
					break
				}
			}
		}

		return nil
	})

	return files, err
}

// 生成新文件名
func generateNewName(oldName, pattern, replace, prefix, suffix string, num int) string {
	newName := oldName

	// 移除扩展名
	newName = strings.TrimSuffix(newName, filepath.Ext(newName))

	// 正则替换
	if pattern != "" && replace != "" {
		re, err := regexp.Compile(pattern)
		if err == nil {
			newName = re.ReplaceAllString(newName, replace)
		}
	}

	// 添加前缀
	if prefix != "" {
		newName = prefix + newName
	}

	// 添加后缀
	if suffix != "" {
		newName = newName + suffix
	}

	// 如果进行了替换操作，添加序号
	if pattern != "" && replace != "" || prefix != "" || suffix != "" {
		newName = fmt.Sprintf("%s_%03d", newName, num)
	}

	return newName
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
