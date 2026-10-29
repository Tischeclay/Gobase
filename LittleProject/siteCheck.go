package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

type Result struct {
	URL      string
	Status   int
	Duration time.Duration
	Err      error
}

func check(client *http.Client, rawURL string) Result {
	result := Result{URL: rawURL}

	// 未指定协议时，默认使用 HTTPS。
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		result.Err = fmt.Errorf("网址格式错误：%w", err)
		return result
	}
	if parsed.Hostname() == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		result.Err = fmt.Errorf("需要有效的 HTTP 或 HTTPS 网址")
		return result
	}

	start := time.Now()
	resp, err := client.Get(rawURL)
	result.Duration = time.Since(start)

	if err != nil {
		result.Err = err
		return result
	}
	defer resp.Body.Close()

	result.Status = resp.StatusCode

	// 最多读取 4 KiB，避免下载整个网页。
	// 耗时仅统计到收到响应头为止。
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	return result
}

func run() int {
	timeout := flag.Duration("timeout", 5*time.Second, "每个请求的超时时间")
	concurrency := flag.Int("c", 5, "最大并发数")

	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"用法：go run sitecheck.go [选项] 网址1 网址2 ...\n\n",
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return 1
	}
	if *timeout <= 0 || *concurrency <= 0 {
		fmt.Fprintln(os.Stderr, "超时时间和并发数必须大于 0")
		return 1
	}

	client := &http.Client{
		Timeout: *timeout,
	}
	defer client.CloseIdleConnections()

	urls := flag.Args()
	results := make([]Result, len(urls))

	workers := *concurrency
	if workers > len(urls) {
		workers = len(urls)
	}

	jobs := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for index := range jobs {
				results[index] = check(client, urls[index])
			}
		}()
	}

	for index := range urls {
		jobs <- index
	}
	close(jobs)
	wg.Wait()

	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "网址\t状态码\t耗时\t结果")

	failed := 0
	for _, result := range results {
		if result.Err != nil {
			failed++
			fmt.Fprintf(
				writer,
				"%s\t-\t%s\t请求失败：%v\n",
				result.URL,
				result.Duration.Round(time.Millisecond),
				result.Err,
			)
			continue
		}

		message := "正常"
		if result.Status < 200 || result.Status >= 400 {
			message = "异常"
			failed++
		}

		fmt.Fprintf(
			writer,
			"%s\t%d\t%s\t%s\n",
			result.URL,
			result.Status,
			result.Duration.Round(time.Millisecond),
			message,
		)
	}
	writer.Flush()

	fmt.Printf("\n检查完成：共 %d 个，正常 %d 个，异常 %d 个\n",
		len(results), len(results)-failed, failed)

	if failed > 0 {
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
