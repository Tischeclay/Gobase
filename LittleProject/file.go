// batch_rename.go
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// 重命名规则配置
type RenameConfig struct {
	Directory     string   // 目标目录
	Pattern       string   // 匹配模式
	Replace       string   // 替换内容
	Prefix        string   // 添加前缀
	Suffix        string   // 添加后缀
	Numbering     bool     // 是否添加序号
	StartNumber   int      // 起始序号
	NumberWidth   int      // 序号宽度 (如 3 -> 001)
	Lowercase     bool     // 转小写
	Uppercase     bool     // 转大写
	RemoveSpaces  bool     // 移除空格
	ReplaceSpaces string   // 替换空格为指定字符
	DryRun        bool     // 预览模式，不实际执行
	Recursive     bool     // 递归处理子目录
	Extensions    []string // 只处理指定扩展名的文件
	Exclude       []string // 排除的文件名模式
}

// 文件信息
type FileInfo struct {
	OldPath string
	NewPath string
	OldName string
	NewName string
	Error   error
}

// 重命名器
type Renamer struct {
	config RenameConfig
	files  []FileInfo
	mu     sync.Mutex
}

func NewRenamer(config RenameConfig) *Renamer {
	return &Renamer{
		config: config,
		files:  make([]FileInfo, 0),
	}
}

// 扫描文件
func (r *Renamer) Scan() error {
	var files []string

	if r.config.Recursive {
		err := filepath.Walk(r.config.Directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		dirFiles, err := ioutil.ReadDir(r.config.Directory)
		if err != nil {
			return err
		}
		for _, file := range dirFiles {
			if !file.IsDir() {
				files = append(files, filepath.Join(r.config.Directory, file.Name()))
			}
		}
	}

	// 过滤文件
	for _, file := range files {
		if r.shouldProcess(file) {
			newName := r.generateNewName(file)
			newPath := filepath.Join(filepath.Dir(file), newName)

			r.mu.Lock()
			r.files = append(r.files, FileInfo{
				OldPath: file,
				NewPath: newPath,
				OldName: filepath.Base(file),
				NewName: newName,
			})
			r.mu.Unlock()
		}
	}

	return nil
}

// 检查是否应该处理该文件
func (r *Renamer) shouldProcess(filePath string) bool {
	fileName := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(fileName))

	// 检查扩展名
	if len(r.config.Extensions) > 0 {
		matched := false
		for _, e := range r.config.Extensions {
			if ext == "."+strings.ToLower(e) || (e == "*" && ext != "") {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查排除模式
	for _, pattern := range r.config.Exclude {
		if matched, _ := regexp.MatchString(pattern, fileName); matched {
			return false
		}
	}

	// 检查匹配模式
	if r.config.Pattern != "" {
		matched, err := regexp.MatchString(r.config.Pattern, fileName)
		if err != nil {
			return false
		}
		if !matched {
			return false
		}
	}

	return true
}

// 生成新文件名
func (r *Renamer) generateNewName(filePath string) string {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	newName := nameWithoutExt

	// 应用模式替换
	if r.config.Pattern != "" && r.config.Replace != "" {
		re := regexp.MustCompile(r.config.Pattern)
		newName = re.ReplaceAllString(newName, r.config.Replace)
	}

	// 移除空格
	if r.config.RemoveSpaces {
		newName = strings.ReplaceAll(newName, " ", "")
	}

	// 替换空格
	if r.config.ReplaceSpaces != "" {
		newName = strings.ReplaceAll(newName, " ", r.config.ReplaceSpaces)
	}

	// 添加前缀
	if r.config.Prefix != "" {
		newName = r.config.Prefix + newName
	}

	// 添加后缀
	if r.config.Suffix != "" {
		newName = newName + r.config.Suffix
	}

	// 大小写转换
	if r.config.Lowercase {
		newName = strings.ToLower(newName)
	} else if r.config.Uppercase {
		newName = strings.ToUpper(newName)
	}

	// 处理重复文件名
	newNameWithExt := newName + ext
	if newNameWithExt != fileName {
		// 检查是否存在同名文件
		newPath := filepath.Join(filepath.Dir(filePath), newNameWithExt)
		if _, err := os.Stat(newPath); err == nil && newPath != filePath {
			// 添加序号避免冲突
			for i := 1; i < 1000; i++ {
				testName := fmt.Sprintf("%s_%d%s", newName, i, ext)
				testPath := filepath.Join(filepath.Dir(filePath), testName)
				if _, err := os.Stat(testPath); os.IsNotExist(err) {
					newNameWithExt = testName
					break
				}
			}
		}
	}

	return newNameWithExt
}

// 添加序号
func (r *Renamer) AddNumbering() {
	// 按原文件名排序
	sort.Slice(r.files, func(i, j int) bool {
		return r.files[i].OldName < r.files[j].OldName
	})

	// 分组处理（按目录）
	dirGroups := make(map[string][]*FileInfo)
	for i := range r.files {
		dir := filepath.Dir(r.files[i].OldPath)
		dirGroups[dir] = append(dirGroups[dir], &r.files[i])
	}

	// 为每个目录的文件添加序号
	for _, group := range dirGroups {
		for idx, file := range group {
			if r.config.Numbering {
				ext := filepath.Ext(file.NewName)
				nameWithoutExt := strings.TrimSuffix(file.NewName, ext)
				number := r.config.StartNumber + idx
				numberStr := strconv.Itoa(number)
				if r.config.NumberWidth > 0 {
					numberStr = fmt.Sprintf("%0*d", r.config.NumberWidth, number)
				}
				file.NewName = fmt.Sprintf("%s_%s%s", nameWithoutExt, numberStr, ext)
				file.NewPath = filepath.Join(filepath.Dir(file.OldPath), file.NewName)
			}
		}
	}
}

// 预览重命名
func (r *Renamer) Preview() {
	fmt.Println("\n📋 重命名预览:")
	fmt.Println(strings.Repeat("=", 80))

	for i, file := range r.files {
		if file.Error != nil {
			fmt.Printf("[%d] ❌ %s\n", i+1, file.Error)
			continue
		}

		if file.OldName == file.NewName {
			fmt.Printf("[%d] ⏭️  %s (无变化)\n", i+1, file.OldName)
		} else {
			fmt.Printf("[%d] 📝 %s\n", i+1, file.OldName)
			fmt.Printf("     → %s\n", file.NewName)
		}
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("总计: %d 个文件\n", len(r.files))
}

// 执行重命名
func (r *Renamer) Execute() error {
	var errors []error

	for _, file := range r.files {
		if file.OldName == file.NewName {
			continue
		}

		if err := os.Rename(file.OldPath, file.NewPath); err != nil {
			errors = append(errors, fmt.Errorf("重命名失败 %s -> %s: %v", file.OldName, file.NewName, err))
		} else {
			fmt.Printf("✅ %s → %s\n", file.OldName, file.NewName)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\n❌ 出现 %d 个错误:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("   %v\n", err)
		}
		return fmt.Errorf("部分文件重命名失败")
	}

	return nil
}

// 生成报告
func (r *Renamer) GenerateReport() string {
	var sb strings.Builder

	sb.WriteString("\n📊 重命名报告\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString(fmt.Sprintf("目录: %s\n", r.config.Directory))
	sb.WriteString(fmt.Sprintf("文件总数: %d\n", len(r.files)))

	changed := 0
	for _, file := range r.files {
		if file.OldName != file.NewName {
			changed++
		}
	}
	sb.WriteString(fmt.Sprintf("重命名: %d\n", changed))
	sb.WriteString(fmt.Sprintf("无变化: %d\n", len(r.files)-changed))

	if len(r.config.Extensions) > 0 {
		sb.WriteString(fmt.Sprintf("扩展名: %v\n", r.config.Extensions))
	}
	if r.config.Pattern != "" {
		sb.WriteString(fmt.Sprintf("匹配模式: %s\n", r.config.Pattern))
		sb.WriteString(fmt.Sprintf("替换为: %s\n", r.config.Replace))
	}
	if r.config.Prefix != "" {
		sb.WriteString(fmt.Sprintf("添加前缀: %s\n", r.config.Prefix))
	}
	if r.config.Suffix != "" {
		sb.WriteString(fmt.Sprintf("添加后缀: %s\n", r.config.Suffix))
	}
	if r.config.Numbering {
		sb.WriteString(fmt.Sprintf("添加序号: 从 %d 开始\n", r.config.StartNumber))
	}

	return sb.String()
}

// ==================== 交互式界面 ====================

func interactiveMode() {
	fmt.Println("\n🎨 批量文件重命名工具 - 交互模式")
	fmt.Println(strings.Repeat("=", 50))

	config := RenameConfig{}

	// 获取目录
	fmt.Print("请输入目录路径 (默认当前目录): ")
	var dir string
	fmt.Scanln(&dir)
	if dir == "" {
		dir, _ = os.Getwd()
	}
	config.Directory = dir

	// 获取选项
	fmt.Print("是否递归处理子目录? (y/n, 默认n): ")
	var recursive string
	fmt.Scanln(&recursive)
	config.Recursive = recursive == "y" || recursive == "Y"

	fmt.Print("只处理特定扩展名? (多个用逗号分隔, 如: jpg,png, 留空处理所有): ")
	var extStr string
	fmt.Scanln(&extStr)
	if extStr != "" {
		config.Extensions = strings.Split(extStr, ",")
		for i := range config.Extensions {
			config.Extensions[i] = strings.TrimSpace(config.Extensions[i])
		}
	}

	fmt.Print("添加前缀 (留空跳过): ")
	fmt.Scanln(&config.Prefix)

	fmt.Print("添加后缀 (留空跳过): ")
	fmt.Scanln(&config.Suffix)

	fmt.Print("是否添加序号? (y/n): ")
	var addNumber string
	fmt.Scanln(&addNumber)
	if addNumber == "y" || addNumber == "Y" {
		config.Numbering = true
		fmt.Print("起始序号 (默认1): ")
		var startNum int
		fmt.Scanln(&startNum)
		config.StartNumber = startNum
		if config.StartNumber == 0 {
			config.StartNumber = 1
		}
		fmt.Print("序号宽度 (如 3=001, 默认0=不补零): ")
		fmt.Scanln(&config.NumberWidth)
	}

	fmt.Print("是否转小写? (y/n): ")
	var toLower string
	fmt.Scanln(&toLower)
	config.Lowercase = toLower == "y" || toLower == "Y"

	fmt.Print("是否转大写? (y/n): ")
	var toUpper string
	fmt.Scanln(&toUpper)
	config.Uppercase = toUpper == "y" || toUpper == "Y"

	fmt.Print("是否移除空格? (y/n): ")
	var removeSpaces string
	fmt.Scanln(&removeSpaces)
	config.RemoveSpaces = removeSpaces == "y" || removeSpaces == "Y"

	fmt.Print("替换空格为? (留空不移除): ")
	fmt.Scanln(&config.ReplaceSpaces)

	fmt.Print("匹配模式 (正则表达式, 留空跳过): ")
	fmt.Scanln(&config.Pattern)
	if config.Pattern != "" {
		fmt.Print("替换为: ")
		fmt.Scanln(&config.Replace)
	}

	fmt.Print("\n预览模式? (y/n, 预览后确认执行): ")
	var dryRun string
	fmt.Scanln(&dryRun)
	config.DryRun = dryRun == "y" || dryRun == "Y"

	// 执行重命名
	renamer := NewRenamer(config)

	fmt.Println("\n正在扫描文件...")
	if err := renamer.Scan(); err != nil {
		log.Fatal("扫描失败:", err)
	}

	if config.Numbering {
		renamer.AddNumbering()
	}

	// 显示预览
	renamer.Preview()

	if config.DryRun {
		fmt.Println("\n🔍 预览模式，未实际执行重命名")
		fmt.Println(renamer.GenerateReport())
		return
	}

	// 确认执行
	fmt.Print("\n确认执行重命名? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		fmt.Println("已取消")
		return
	}

	// 执行重命名
	fmt.Println("\n正在执行重命名...")
	if err := renamer.Execute(); err != nil {
		fmt.Println("执行失败:", err)
	} else {
		fmt.Println("\n✅ 重命名完成！")
		fmt.Println(renamer.GenerateReport())
	}
}

// ==================== 命令行模式 ====================

func main() {
	// 命令行参数
	var (
		dir           string
		prefix        string
		suffix        string
		pattern       string
		replace       string
		extensions    string
		exclude       string
		numbering     bool
		startNumber   int
		numberWidth   int
		lowercase     bool
		uppercase     bool
		removeSpaces  bool
		replaceSpaces string
		recursive     bool
		dryRun        bool
		interactive   bool
	)

	flag.StringVar(&dir, "dir", ".", "目标目录")
	flag.StringVar(&prefix, "prefix", "", "添加前缀")
	flag.StringVar(&suffix, "suffix", "", "添加后缀")
	flag.StringVar(&pattern, "pattern", "", "匹配模式 (正则表达式)")
	flag.StringVar(&replace, "replace", "", "替换内容")
	flag.StringVar(&extensions, "ext", "", "只处理指定扩展名 (逗号分隔)")
	flag.StringVar(&exclude, "exclude", "", "排除的文件名模式 (正则表达式)")
	flag.BoolVar(&numbering, "number", false, "添加序号")
	flag.IntVar(&startNumber, "start", 1, "起始序号")
	flag.IntVar(&numberWidth, "width", 0, "序号宽度 (0=不补零)")
	flag.BoolVar(&lowercase, "lower", false, "转小写")
	flag.BoolVar(&uppercase, "upper", false, "转大写")
	flag.BoolVar(&removeSpaces, "rmspace", false, "移除空格")
	flag.StringVar(&replaceSpaces, "rpspace", "", "替换空格为指定字符")
	flag.BoolVar(&recursive, "r", false, "递归处理子目录")
	flag.BoolVar(&dryRun, "dry", false, "预览模式，不实际执行")
	flag.BoolVar(&interactive, "i", false, "交互模式")
	flag.Parse()

	// 交互模式
	if interactive {
		interactiveMode()
		return
	}

	// 命令行模式
	config := RenameConfig{
		Directory:     dir,
		Prefix:        prefix,
		Suffix:        suffix,
		Pattern:       pattern,
		Replace:       replace,
		Numbering:     numbering,
		StartNumber:   startNumber,
		NumberWidth:   numberWidth,
		Lowercase:     lowercase,
		Uppercase:     uppercase,
		RemoveSpaces:  removeSpaces,
		ReplaceSpaces: replaceSpaces,
		Recursive:     recursive,
		DryRun:        dryRun,
	}

	if extensions != "" {
		config.Extensions = strings.Split(extensions, ",")
		for i := range config.Extensions {
			config.Extensions[i] = strings.TrimSpace(config.Extensions[i])
		}
	}

	if exclude != "" {
		config.Exclude = []string{exclude}
	}

	// 创建重命名器
	renamer := NewRenamer(config)

	// 扫描文件
	fmt.Println("正在扫描文件...")
	if err := renamer.Scan(); err != nil {
		log.Fatal("扫描失败:", err)
	}

	// 添加序号
	if numbering {
		renamer.AddNumbering()
	}

	// 显示预览
	renamer.Preview()

	// 预览模式
	if dryRun {
		fmt.Println("\n🔍 预览模式，未实际执行重命名")
		fmt.Println(renamer.GenerateReport())
		return
	}

	// 确认执行
	fmt.Print("\n确认执行重命名? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" {
		fmt.Println("已取消")
		return
	}

	// 执行重命名
	fmt.Println("\n正在执行重命名...")
	if err := renamer.Execute(); err != nil {
		fmt.Println("执行失败:", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ 重命名完成！")
	fmt.Println(renamer.GenerateReport())
}
