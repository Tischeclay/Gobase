// jsonpretty.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// 颜色定义
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

type JSONProcessor struct {
	inputFile  string
	outputFile string
	indent     int
	minify     bool
	color      bool
	validate   bool
	sortKeys   bool
	escapeHTML bool
	prefix     string
	lineWidth  int
	noColor    bool
}

type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
	Size   int      `json:"size"`
	Lines  int      `json:"lines"`
}

func NewJSONProcessor() *JSONProcessor {
	return &JSONProcessor{
		indent:     2,
		lineWidth:  80,
		escapeHTML: true,
	}
}

// 处理JSON数据
func (jp *JSONProcessor) Process(data []byte) ([]byte, error) {
	// 解析JSON
	var v interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&v); err != nil {
		return nil, fmt.Errorf("无效的JSON: %v", err)
	}

	// 只验证不输出
	if jp.validate {
		return jp.validateJSON(data)
	}

	// 设置编码选项
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(jp.escapeHTML)

	if jp.sortKeys {
		encoder.SetIndent(jp.prefix, strings.Repeat(" ", jp.indent))
	}

	// 输出处理
	var out []byte
	var err error

	if jp.minify {
		// 压缩JSON
		err = encoder.Encode(v)
		if err == nil {
			out = buf.Bytes()
		}
	} else {
		// 美化JSON
		if jp.sortKeys {
			// 使用带缩进的编码
			err = encoder.Encode(v)
			if err == nil {
				out = buf.Bytes()
			}
		} else {
			// 使用MarshalIndent
			out, err = json.MarshalIndent(v, jp.prefix, strings.Repeat(" ", jp.indent))
		}
	}

	if err != nil {
		return nil, err
	}

	// 去除末尾的换行符
	out = bytes.TrimRight(out, "\n")

	// 添加颜色
	if jp.color && !jp.minify && !jp.noColor {
		out = jp.addColor(out)
	}

	// 格式化行宽（如果有需要）
	if jp.lineWidth > 0 && !jp.minify && len(out) > jp.lineWidth {
		out = jp.formatLineWidth(out)
	}

	return out, nil
}

// 验证JSON
func (jp *JSONProcessor) validateJSON(data []byte) ([]byte, error) {
	result := ValidationResult{
		Valid: true,
		Size:  len(data),
		Lines: bytes.Count(data, []byte{'\n'}) + 1,
	}

	var v interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&v); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())

		// 尝试定位错误位置
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			lines := bytes.Split(data, []byte{'\n'})
			pos := 0
			for i, line := range lines {
				if pos+len(line) >= int(syntaxErr.Offset) {
					lineNum := i + 1
					colNum := int(syntaxErr.Offset) - pos
					result.Errors = append(result.Errors,
						fmt.Sprintf("位置: 第%d行, 第%d列", lineNum, colNum))

					// 显示错误行
					if lineNum <= len(lines) {
						contextLine := string(lines[lineNum-1])
						result.Errors = append(result.Errors,
							fmt.Sprintf("上下文: %s", contextLine))

						// 显示标记
						marker := strings.Repeat(" ", colNum-1) + "^"
						result.Errors = append(result.Errors, marker)
					}
					break
				}
				pos += len(line) + 1
			}
		}
	}

	// 返回结果
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.Encode(result)

	return buf.Bytes(), nil
}

// 添加颜色高亮
func (jp *JSONProcessor) addColor(data []byte) []byte {
	text := string(data)

	// 定义正则表达式模式
	patterns := []struct {
		pattern string
		color   string
	}{
		// 字符串值
		{`"([^"\\]|\\.)*"`, colorGreen},
		// 数字
		{`\b-?\d+(\.\d+)?([eE][+-]?\d+)?\b`, colorYellow},
		// 布尔值
		{`\b(true|false)\b`, colorBlue},
		// null
		{`\bnull\b`, colorRed},
		// 键名
		{`"([^"\\]|\\.)*"(?=\s*:)`, colorCyan},
		// 冒号
		{`:`, colorWhite + colorBold},
		// 括号
		{`[\[\]{}]`, colorPurple + colorBold},
		// 逗号
		{`,`, colorWhite},
	}

	// 应用颜色
	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		text = re.ReplaceAllStringFunc(text, func(match string) string {
			// 不要给空字符串加颜色
			if match == "" {
				return match
			}
			return p.color + match + colorReset
		})
	}

	return []byte(text)
}

// 格式化行宽
func (jp *JSONProcessor) formatLineWidth(data []byte) []byte {
	// 简单实现：在逗号后添加换行
	text := string(data)
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if len(line) > jp.lineWidth {
			// 在适当位置分割长行
			parts := jp.splitLongLine(line)
			lines[i] = strings.Join(parts, "\n"+strings.Repeat(" ", jp.indent))
		}
	}

	return []byte(strings.Join(lines, "\n"))
}

// 分割长行
func (jp *JSONProcessor) splitLongLine(line string) []string {
	var parts []string
	var current strings.Builder

	for _, r := range line {
		current.WriteRune(r)
		if current.Len() >= jp.lineWidth {
			// 在逗号或括号后分割
			if r == ',' || r == '{' || r == '[' {
				parts = append(parts, current.String())
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// 从文件读取
func (jp *JSONProcessor) ReadFromFile(filename string) ([]byte, error) {
	if filename == "" || filename == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(filename)
}

// 写入到文件
func (jp *JSONProcessor) WriteToFile(filename string, data []byte) error {
	if filename == "" || filename == "-" {
		fmt.Println(string(data))
		return nil
	}
	return os.WriteFile(filename, data, 0644)
}

// 处理文件
func (jp *JSONProcessor) ProcessFile(filename string) error {
	// 读取数据
	data, err := jp.ReadFromFile(filename)
	if err != nil {
		return fmt.Errorf("读取失败: %v", err)
	}

	// 处理JSON
	result, err := jp.Process(data)
	if err != nil {
		return err
	}

	// 写入结果
	return jp.WriteToFile(jp.outputFile, result)
}

// 获取统计信息
func (jp *JSONProcessor) GetStats(data []byte) map[string]interface{} {
	stats := make(map[string]interface{})

	stats["size"] = len(data)
	stats["lines"] = bytes.Count(data, []byte{'\n'}) + 1

	// 统计JSON结构
	var v interface{}
	if err := json.Unmarshal(data, &v); err == nil {
		stats["type"] = fmt.Sprintf("%T", v)

		// 计算深度
		stats["depth"] = jp.calcDepth(v, 1)

		// 统计键值对数量
		if obj, ok := v.(map[string]interface{}); ok {
			stats["keys"] = len(obj)
		}
		if arr, ok := v.([]interface{}); ok {
			stats["items"] = len(arr)
		}
	}

	return stats
}

// 计算JSON深度
func (jp *JSONProcessor) calcDepth(v interface{}, current int) int {
	maxDepth := current

	switch val := v.(type) {
	case map[string]interface{}:
		for _, child := range val {
			if depth := jp.calcDepth(child, current+1); depth > maxDepth {
				maxDepth = depth
			}
		}
	case []interface{}:
		for _, child := range val {
			if depth := jp.calcDepth(child, current+1); depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth
}

func main() {
	jp := NewJSONProcessor()

	// 定义命令行参数
	input := flag.String("i", "", "输入文件 (默认标准输入)")
	output := flag.String("o", "", "输出文件 (默认标准输出)")
	indent := flag.Int("indent", 2, "缩进空格数")
	minify := flag.Bool("m", false, "压缩JSON")
	color := flag.Bool("c", false, "彩色输出")
	noColor := flag.Bool("no-color", false, "禁用颜色")
	validate := flag.Bool("v", false, "验证JSON格式")
	sortKeys := flag.Bool("sort", false, "排序键名")
	escapeHTML := flag.Bool("escape", true, "转义HTML字符")
	lineWidth := flag.Int("width", 0, "最大行宽 (0=不限制)")
	prefix := flag.String("prefix", "", "每行前缀")
	stats := flag.Bool("stats", false, "显示统计信息")

	flag.Parse()

	// 设置参数
	jp.inputFile = *input
	jp.outputFile = *output
	jp.indent = *indent
	jp.minify = *minify
	jp.color = *color
	jp.noColor = *noColor
	jp.validate = *validate
	jp.sortKeys = *sortKeys
	jp.escapeHTML = *escapeHTML
	jp.lineWidth = *lineWidth
	jp.prefix = *prefix

	// 如果没有指定输入文件，使用命令行参数
	if jp.inputFile == "" && flag.NArg() > 0 {
		jp.inputFile = flag.Arg(0)
	}

	// 处理JSON
	data, err := jp.ReadFromFile(jp.inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 显示统计信息
	if *stats {
		statsData := jp.GetStats(data)
		fmt.Fprintf(os.Stderr, "📊 统计信息:\n")
		fmt.Fprintf(os.Stderr, "  大小: %d 字节\n", statsData["size"])
		fmt.Fprintf(os.Stderr, "  行数: %d\n", statsData["lines"])
		if t, ok := statsData["type"]; ok {
			fmt.Fprintf(os.Stderr, "  类型: %v\n", t)
		}
		if d, ok := statsData["depth"]; ok {
			fmt.Fprintf(os.Stderr, "  深度: %v\n", d)
		}
		if k, ok := statsData["keys"]; ok {
			fmt.Fprintf(os.Stderr, "  键数: %v\n", k)
		}
		if items, ok := statsData["items"]; ok {
			fmt.Fprintf(os.Stderr, "  元素数: %v\n", items)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// 处理JSON
	result, err := jp.Process(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	if err := jp.WriteToFile(jp.outputFile, result); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
