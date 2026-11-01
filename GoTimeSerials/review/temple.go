package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Reading 表示设备的一次采集结果。
type Reading struct {
	DeviceID string
	Time     time.Time
	Power    float64 // 功率，单位 kW
}

// generateReading 模拟采集一次设备数据。
func generateReading(deviceID string) Reading {
	// 模拟功率在 [100, 150) kW 之间变化。
	power := 100 + rand.Float64()*50

	return Reading{
		DeviceID: deviceID,
		Time:     time.Now(),
		Power:    power,
	}
}

func main() {
	deviceID := "meter-001"
	count := 10
	var totalPower float64

	fmt.Printf("开始采集，设备：%s\n", deviceID)

	for i := 0; i < count; i++ {
		reading := generateReading(deviceID)
		totalPower += reading.Power

		fmt.Printf(
			"第 %02d 条 | %s | 设备：%s | 功率：%.2f kW\n",
			i+1,
			reading.Time.Format("2006-01-02 15:04:05"),
			reading.DeviceID,
			reading.Power,
		)

		// 最后一条采集完成后，无须继续等待。
		if i < count-1 {
			time.Sleep(time.Second)
		}
	}

	average := totalPower / float64(count)
	fmt.Printf("\n采集完成，共 %d 条，平均功率：%.2f kW\n", count, average)
}
