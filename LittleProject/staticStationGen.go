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

func main() {
	// 命令行参数
	inputDir := flag.String("i", "./content", "内容目录")
	outputDir := flag.String("o", "./public", "输出目录")
	templateDir := flag.String("t", "./templates", "模板目录")
	siteTitle := flag.String("title", "我的静态站点", "站点标题")
	baseURL := flag.String("base", "/", "基础URL")
	flag.Parse()

	// 创建输出目录
	os.MkdirAll(*outputDir, 0755)

	// 读取所有Markdown文件
	var pages []Page
	filepath.Walk(*inputDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
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
				if strings.Contains(line, ":") {
					kv := strings.SplitN(line, ":", 2)
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])

					switch key {
					case "title":
						page.Title = value
					case "date":
						page.Date, _ = time.Parse("2006-01-02", value)
					case "tags":
						page.Tags = strings.Split(value, ",")
					case "description":
						page.Description = value
					}
				}
			}
		} else {
			// 没有前置元数据，整个文件就是内容
			page.RawContent = string(content)
			page.Title = strings.TrimSuffix(filepath.Base(path), ".md")
		}

		// 生成slug
		page.Slug = strings.TrimSuffix(filepath.Base(path), ".md")

		// 转换Markdown到HTML
		html := blackfriday.Run([]byte(page.RawContent))
		page.Content = template.HTML(html)

		pages = append(pages, page)
		fmt.Printf("处理: %s\n", path)
		return nil
	})

	// 解析模板
	tmpl, err := template.ParseGlob(filepath.Join(*templateDir, "*.html"))
	if err != nil {
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
	indexFile, err := os.Create(filepath.Join(*outputDir, "index.html"))
	if err == nil {
		tmpl.ExecuteTemplate(indexFile, "index", site)
		indexFile.Close()
	}

	// 生成每个页面
	for _, page := range pages {
		pageDir := filepath.Join(*outputDir, page.Slug)
		os.MkdirAll(pageDir, 0755)

		pageFile, err := os.Create(filepath.Join(pageDir, "index.html"))
		if err == nil {
			tmpl.ExecuteTemplate(pageFile, "page", page)
			pageFile.Close()
		}
	}

	fmt.Printf("完成! 生成了 %d 个页面到 %s\n", len(pages), *outputDir)
}

//const defaultTemplate = `<!DOCTYPE html>
//<html>
//<head>
//    <meta charset="UTF-8">
//    <title>{{.Title}}</title>
//    <style>
//        body { font-family: Arial; max-width: 800px; margin: 0 auto; padding: 20px; }
//        article { margin: 40px 0; border-bottom: 1px solid #eee; }
//        .date { color: #666; font-size: 0.9em; }
//    </style>
//</head>
//<body>
//    <h1>{{.Title}}</h1>
//
//    {{define "index"}}
//        {{range .Pages}}
//        <article>
//            <h2><a href="/{{.Slug}}/">{{.Title}}</a></h2>
//            <div class="date">{{.Date.Format "2006-01-02"}}</div>
//            <div>{{.Content}}</div>
//        </article>
//        {{end}}
//    {{end}}
//
//    {{define "page"}}
//        <article>
//            <h1>{{.Title}}</h1>
//            <div class="date">{{.Date.Format "2006-01-02"}}</div>
//            <div>{{.Content}}</div>
//            <p><a href="/">← 返回首页</a></p>
//        </article>
//    {{end}}
//</body>
//</html>`
