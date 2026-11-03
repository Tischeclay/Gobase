package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Device 对应 JSON 中的一台设备。
// json 标签指定字段对应的 JSON 键名。
type Device struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Power  float64 `json:"power"`
	Online bool    `json:"online"`
}

// SiteData 对应整个 JSON 文件。
type SiteData struct {
	Site    string   `json:"site"`
	Devices []Device `json:"devices"`
}

func readJSON(path string) (SiteData, error) {
	var result SiteData

	// 1. 读取文件，得到字节切片。
	data, err := os.ReadFile(path)
	if err != nil {
		return result, fmt.Errorf("读取文件失败：%w", err)
	}

	// 2. 将 JSON 解析到结构体。
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("解析 JSON 失败：%w", err)
	}

	return result, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "用法：go run main.go devices.json")
		os.Exit(1)
	}

	result, err := readJSON(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误：", err)
		os.Exit(1)
	}

	fmt.Println("站点：", result.Site)
	fmt.Printf("设备数量：%d\n\n", len(result.Devices))

	// 3. 遍历设备，打印信息。
	for _, device := range result.Devices {
		status := "离线"
		if device.Online {
			status = "在线"
		}

		fmt.Printf(
			"编号：%s | 名称：%s | 功率：%.2f kW | 状态：%s\n",
			device.ID,
			device.Name,
			device.Power,
			status,
		)
	}
}
