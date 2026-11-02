package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Reading struct {
	DeviceID string    `json:"device_id"`
	Time     time.Time `json:"time"`
	Power    float64   `json:"power"`
}

// 模拟采集一条数据。
func generateReading(deviceID string) Reading {
	return Reading{
		DeviceID: deviceID,
		Time:     time.Now(),
		Power:    100 + rand.Float64()*50,
	}
}

// 将数据保存为 JSON 文件。
// 如果文件已存在，会覆盖原内容。
func saveReadings(path string, readings []Reading) error {
	data, err := json.MarshalIndent(readings, "", "  ")
	if err != nil {
		return fmt.Errorf("转换为 JSON：%w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入文件：%w", err)
	}

	return nil
}

// 从 JSON 文件读取数据。
func loadReadings(path string) ([]Reading, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件：%w", err)
	}

	var readings []Reading
	if err := json.Unmarshal(data, &readings); err != nil {
		return nil, fmt.Errorf("解析 JSON：%w", err)
	}

	return readings, nil
}

// 统计平均功率和最大功率。
func printSummary(readings []Reading) {
	if len(readings) == 0 {
		fmt.Println("没有可统计的数据")
		return
	}

	var total float64
	maxPower := readings[0].Power

	for _, reading := range readings {
		total += reading.Power

		if reading.Power > maxPower {
			maxPower = reading.Power
		}
	}

	fmt.Printf("\n读取到 %d 条数据\n", len(readings))
	fmt.Printf("平均功率：%.2f kW\n", total/float64(len(readings)))
	fmt.Printf("最大功率：%.2f kW\n", maxPower)
}

func run() error {
	const path = "readings.json"

	// read 模式只读取已有文件，不重新采集。
	if len(os.Args) == 2 && os.Args[1] == "read" {
		readings, err := loadReadings(path)
		if err != nil {
			return err
		}
		printSummary(readings)
		return nil
	}

	if len(os.Args) != 1 {
		return fmt.Errorf("用法：go run main.go [read]")
	}

	readings := make([]Reading, 0, 5)

	for i := 0; i < 5; i++ {
		reading := generateReading("meter-001")
		readings = append(readings, reading)

		fmt.Printf(
			"%s | %s | %.2f kW\n",
			reading.Time.Format("15:04:05"),
			reading.DeviceID,
			reading.Power,
		)

		if i < 4 {
			time.Sleep(time.Second)
		}
	}

	if err := saveReadings(path, readings); err != nil {
		return err
	}

	fmt.Println("数据已保存到", path)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "运行失败：", err)
		os.Exit(1)
	}
}
