// ssg.go
package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Page struct {
	Title       string
	Date        time.Time
	Content     template.HTML
	RawContent  string
	Slug        string
	Tags        []string
	Description string
}

type Site struct {
	Pages     []Page
	Title     string
	BaseURL   string
	BuildTime time.Time
}

// 默认模板常量
const defaultTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
        body { 
            font-family: 'Microsoft YaHei', Arial, sans-serif; 
            max-width: 800px; 
            margin: 0 auto; 
            padding: 20px; 
            line-height: 1.6;
        }
        article { 
            margin: 40px 0; 
            padding: 20px;
            background: #f9f9f9;
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .date { 
            color: #666; 
            font-size: 0.9em;
            margin-bottom: 10px;
        }
        h1, h2, h3 { color: #333; }
        a { color: #0066cc; text-decoration: none; }
        a:hover { text-decoration: underline; }
        code { 
            background: #eee; 
            padding: 2px 5px; 
            border-radius: 3px;
            font-family: 'Courier New', monospace;
        }
        pre code {
            display: block;
            padding: 10px;
            overflow-x: auto;
        }
        .tags {
            margin-top: 10px;
        }
        .tag {
            display: inline-block;
            background: #e0e0e0;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 0.8em;
            margin-right: 5px;
        }
        .footer {
            margin-top: 50px;
            text-align: center;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    
    {{define "index"}}
        {{range .Pages}}
        <article>
            <h2><a href="/{{.Slug}}/">{{.Title}}</a></h2>
            <div class="date">{{.Date.Format "2006-01-02"}}</div>
            <div>{{.Content}}</div>
            {{if .Tags}}
            <div class="tags">
                {{range .Tags}}
                <span class="tag">{{.}}</span>
                {{end}}
            </div>
            {{end}}
        </article>
        {{end}}
        <div class="footer">
            共 {{len .Pages}} 篇文章 | 生成于 {{.BuildTime.Format "2006-01-02 15:04:05"}}
        </div>
    {{end}}
    
    {{define "page"}}
        <article>
            <h1>{{.Title}}</h1>
            <div class="date">{{.Date.Format "2006-01-02"}}</div>
            {{if .Tags}}
            <div class="tags">
                {{range .Tags}}
                <span class="tag">{{.}}</span>
                {{end}}
            </div>
            {{end}}
            <div>{{.Content}}</div>
            <p><a href="/">← 返回首页</a></p>
        </article>
    {{end}}
    
    {{define "list"}}
        <h2>文章列表</h2>
        <ul>
        {{range .Pages}}
            <li><a href="/{{.Slug}}/">{{.Title}}</a> <span class="date">{{.Date.Format "2006-01-02"}}</span></li>
        {{end}}
        </ul>
    {{end}}
</body>
</html>`

func main() {
	// 检查依赖
	if !checkDependencies() {
		fmt.Println("请先安装依赖: go get github.com/russross/blackfriday/v2")
		os.Exit(1)
	}

	// 命令行参数
	inputDir := flag.String("i", "./content", "内容目录 (Markdown文件)")
	outputDir := flag.String("o", "./public", "输出目录 (生成的HTML文件)")
	templateDir := flag.String("t", "./templates", "模板目录 (可选)")
	siteTitle := flag.String("title", "我的静态站点", "站点标题")
	baseURL := flag.String("base", "/", "基础URL")
	flag.Parse()

	// 验证输入目录
	if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
		fmt.Printf("错误: 输入目录 %s 不存在\n", *inputDir)
		createExampleContent(*inputDir)
		fmt.Printf("已创建示例内容目录: %s\n", *inputDir)
	}

	// 创建输出目录
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("错误: 无法创建输出目录: %v\n", err)
		os.Exit(1)
	}

	// 读取所有Markdown文件
	var pages []Page
	err := filepath.Walk(*inputDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		fmt.Printf("处理: %s\n", path)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取文件失败 %s: %v", path, err)
		}

		// 解析前置元数据（简单的---分割）
		parts := strings.SplitN(string(content), "---", 3)
		var page Page

		if len(parts) == 3 {
			// 有前置元数据
			page.RawContent = parts[2]

			// 解析简单的键值对
			lines := strings.Split(parts[1], "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				if strings.Contains(line, ":") {
					kv := strings.SplitN(line, ":", 2)
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])

					switch strings.ToLower(key) {
					case "title":
						page.Title = value
					case "date":
						if t, err := time.Parse("2006-01-02", value); err == nil {
							page.Date = t
						} else {
							page.Date = time.Now()
						}
					case "tags":
						// 支持多种分隔符
						value = strings.Trim(value, "[]\"'")
						tags := strings.FieldsFunc(value, func(r rune) bool {
							return r == ',' || r == ' ' || r == '、'
						})
						for i, tag := range tags {
							tags[i] = strings.TrimSpace(tag)
						}
						page.Tags = tags
					case "description":
						page.Description = value
					}
				}
			}
		} else {
			// 没有前置元数据，整个文件就是内容
			page.RawContent = string(content)
		}

		// 如果标题为空，使用文件名
		if page.Title == "" {
			page.Title = strings.TrimSuffix(filepath.Base(path), ".md")
		}

		// 如果日期为空，使用文件修改时间
		if page.Date.IsZero() {
			page.Date = info.ModTime()
		}

		// 生成slug（URL友好的名称）
		page.Slug = createSlug(page.Title)

		// 转换Markdown到HTML
		html := blackfriday.Run([]byte(page.RawContent),
			blackfriday.WithExtensions(blackfriday.CommonExtensions))
		page.Content = template.HTML(html)

		pages = append(pages, page)
		return nil
	})

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	if len(pages) == 0 {
		fmt.Println("警告: 没有找到任何Markdown文件")
		createExampleContent(*inputDir)
		fmt.Println("已创建示例文章，请重新运行程序")
		os.Exit(0)
	}

	// 按日期排序（最新的在前）
	for i := 0; i < len(pages)-1; i++ {
		for j := i + 1; j < len(pages); j++ {
			if pages[i].Date.Before(pages[j].Date) {
				pages[i], pages[j] = pages[j], pages[i]
			}
		}
	}

	// 解析模板
	var tmpl *template.Template
	var tmplErr error

	if _, err := os.Stat(*templateDir); err == nil {
		// 尝试加载自定义模板
		tmpl, tmplErr = template.ParseGlob(filepath.Join(*templateDir, "*.html"))
	}

	if tmplErr != nil || tmpl == nil {
		// 使用默认模板
		tmpl = template.Must(template.New("index").Parse(defaultTemplate))
	}

	// 构建站点数据
	site := Site{
		Pages:     pages,
		Title:     *siteTitle,
		BaseURL:   *baseURL,
		BuildTime: time.Now(),
	}

	// 生成首页
	indexFile := filepath.Join(*outputDir, "index.html")
	if err := writeTemplate(tmpl, "index", site, indexFile); err != nil {
		fmt.Printf("错误: 生成首页失败: %v\n", err)
	}

	// 生成文章列表页
	listFile := filepath.Join(*outputDir, "list.html")
	if err := writeTemplate(tmpl, "list", site, listFile); err != nil {
		fmt.Printf("错误: 生成列表页失败: %v\n", err)
	}

	// 生成每个文章页面
	for _, page := range pages {
		pageDir := filepath.Join(*outputDir, page.Slug)
		if err := os.MkdirAll(pageDir, 0755); err != nil {
			fmt.Printf("错误: 无法创建目录 %s: %v\n", pageDir, err)
			continue
		}

		pageFile := filepath.Join(pageDir, "index.html")
		if err := writeTemplate(tmpl, "page", page, pageFile); err != nil {
			fmt.Printf("错误: 生成页面 %s 失败: %v\n", page.Slug, err)
		}
	}

	fmt.Printf("\n✅ 完成! 生成了 %d 个页面到 %s\n", len(pages), *outputDir)
	fmt.Printf("📁 首页: %s\n", filepath.Join(*outputDir, "index.html"))
	fmt.Printf("📊 文章列表: %s\n", filepath.Join(*outputDir, "list.html"))
}

func writeTemplate(tmpl *template.Template, name string, data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.ExecuteTemplate(file, name, data)
}

func createSlug(title string) string {
	// 生成URL友好的slug
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, ":", "")
	slug = strings.ReplaceAll(slug, ";", "")
	slug = strings.ReplaceAll(slug, "!", "")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "(", "")
	slug = strings.ReplaceAll(slug, ")", "")
	slug = strings.ReplaceAll(slug, "[", "")
	slug = strings.ReplaceAll(slug, "]", "")
	slug = strings.ReplaceAll(slug, "{", "")
	slug = strings.ReplaceAll(slug, "}", "")

	// 移除连续多个-
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	return strings.Trim(slug, "-")
}

func checkDependencies() bool {
	// 简单检查blackfriday是否可用
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("依赖检查失败:", r)
		}
	}()

	// 尝试创建一个简单的blackfriday实例
	_ = blackfriday.Run([]byte("test"))
	return true
}

func createExampleContent(dir string) {
	// 创建示例内容
	exampleDir := filepath.Join(dir, "posts")
	os.MkdirAll(exampleDir, 0755)

	examples := []struct {
		name    string
		content string
	}{
		{
			name: "welcome.md",
			content: `---
title: 欢迎使用静态站点生成器
date: 2024-01-01
tags: 介绍, 教程
description: 这是一个示例文章
---

# 欢迎使用静态站点生成器

这是一个使用 **Markdown** 编写的示例文章。

## 特性

- 支持 Markdown 语法
- 自动生成漂亮的 HTML
- 支持标签和分类
- 响应式设计

## 代码示例

` + "```go" + `
package main

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

## 列表

- 项目1
- 项目2
- 项目3

[了解更多](https://example.com)
`,
		},
		{
			name: "about.md",
			content: `---
title: 关于本站
date: 2024-01-02
tags: 关于
---

# 关于本站

这是一个使用 Go 语言编写的静态站点生成器。

## 特点

- 简单易用
- 快速生成
- 可自定义模板

> 欢迎使用！
`,
		},
	}

	for _, ex := range examples {
		path := filepath.Join(exampleDir, ex.name)
		if err := os.WriteFile(path, []byte(ex.content), 0644); err != nil {
			fmt.Printf("创建示例文件失败 %s: %v\n", ex.name, err)
		}
	}
}
