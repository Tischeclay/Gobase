// jsonpretty.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type JSONProcessor struct {
	input    string
	output   string
	indent   int
	minify   bool
	color    bool
	validate bool
}

func NewJSONProcessor() *JSONProcessor {
	return &JSONProcessor{}
}

func (jp *JSONProcessor) processFile(inputFile string) ([]byte, error) {
	var data []byte
	var err error

	if inputFile == "-" {
		// 从标准输入读取
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
	} else {
		// 从文件读取
		data, err = os.ReadFile(inputFile)
		if err != nil {
			return nil, err
		}
	}

	return jp.process(data)
}

func (jp *JSONProcessor) process(data []byte) ([]byte, error) {
	// 解析JSON
	var v interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&v); err != nil {
		return nil, fmt.Errorf("无效的JSON: %v", err)
	}

	// 只验证不输出
	if jp.validate {
		return []byte("JSON 格式有效"), nil
	}

	// 输出处理
	var out []byte
	var err error

	if jp.minify {
		// 压缩
		out, err = json.Marshal(v)
	} else {
		// 美化
		out, err = json.MarshalIndent(v, "", strings.Repeat(" ", jp.indent))
	}

	if err != nil {
		return nil, err
	}

	// 添加颜色（简单实现）
	if jp.color && !jp.minify {
		out = jp.addColor(out)
	}

	return out, nil
}

func (jp *JSONProcessor) addColor(data []byte) []byte {
	// 简单颜色实现
	str := string(data)

	// 字符串值 - 绿色
	str = highlightRegex(str, `"[^"]*"`, "\x1b[32m$0\x1b[0m")

	// 数字 - 黄色
	str = highlightRegex(str, `\b\d+\b`, "\x1b[33m$0\x1b[0m")

	// 布尔值 - 蓝色
	str = highlightRegex(str, `\b(true|false)\b`, "\x1b[34m$0\x1b[0m")

	// null - 红色
	str = highlightRegex(str, `\bnull\b`, "\x1b[31m$0\x1b[0m")

	return []byte(str)
}

func highlightRegex(s, pattern, replacement string) string {
	// 简化的正则替换
	return s // 实际需要实现正则
}

func (jp *JSONProcessor) writeOutput(data []byte) error {
	if jp.output == "" {
		// 输出到标准输出
		fmt.Println(string(data))
		return nil
	}

	// 写入文件
	return os.WriteFile(jp.output, data, 0644)
}

func main() {
	jp := NewJSONProcessor()

	// 定义命令行参数
	input := flag.String("i", "", "输入文件 (默认为标准输入)")
	output := flag.String("o", "", "输出文件 (默认为标准输出)")
	indent := flag.Int("indent", 2, "缩进空格数")
	minify := flag.Bool("m", false, "压缩JSON (移除空白)")
	color := flag.Bool("c", false, "彩色输出")
	validate := flag.Bool("v", false, "只验证JSON格式")

	flag.Parse()

	jp.input = *input
	jp.output = *output
	jp.indent = *indent
	jp.minify = *minify
	jp.color = *color
	jp.validate = *validate

	// 如果没有指定输入文件，使用标准输入
	if jp.input == "" && flag.NArg() > 0 {
		jp.input = flag.Arg(0)
	}

	// 处理JSON
	var data []byte
	var err error

	if jp.input == "" {
		// 从标准输入读取
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
			os.Exit(1)
		}
		data, err = jp.process(data)
	} else {
		data, err = jp.processFile(jp.input)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "处理失败: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	if err := jp.writeOutput(data); err != nil {
		fmt.Fprintf(os.Stderr, "写入输出失败: %v\n", err)
		os.Exit(1)
	}
}
