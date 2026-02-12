package monitor

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"performance/dingtalk"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type Config struct {
	DingTalkAccessToken  string   `json:"dingtalk_access_token"`
	AtMobiles            []string `json:"at_mobiles"`
	CheckIntervalSeconds int      `json:"check_interval_seconds"`
	EnableCheck          bool     `json:"enable_check"`
	CpuThreshold         float64  `json:"cpu_threshold"`
	MemThreshold         float64  `json:"mem_threshold"`
	DiskThreshold        float64  `json:"disk_threshold"`
	AlertStartTime       string   `json:"alert_start_time"`
	AlertEndTime         string   `json:"alert_end_time"`
}

var config Config

func LoadConfig() error {
	file, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &config)
}

func Start() {
	// Load Configuration
	if err := LoadConfig(); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	if !config.EnableCheck {
		fmt.Println("Monitoring is disabled in config.")
		return
	}

	// Initialize DingTalk Access Token
	dingtalk.AccessToken = config.DingTalkAccessToken

	fmt.Println("Starting System Monitor...")
	fmt.Printf("Interval: %d seconds\n", config.CheckIntervalSeconds)
	fmt.Printf("Alert Time Window: %s - %s\n", config.AlertStartTime, config.AlertEndTime)
	fmt.Printf("Thresholds - CPU: %.2f%%, Mem: %.2f%%, Disk: %.2f%%\n", config.CpuThreshold, config.MemThreshold, config.DiskThreshold)

	ticker := time.NewTicker(time.Duration(config.CheckIntervalSeconds) * time.Second)
	defer ticker.Stop()

	// Run immediately once
	checkAndAlert()

	for range ticker.C {
		checkAndAlert()
	}
}

func isTimeInWindow(start, end string) bool {
	if start == "" || end == "" {
		return true // If not configured, assume always run
	}
	now := time.Now()
	// Parse HM
	s, err1 := time.Parse("15:04", start)
	e, err2 := time.Parse("15:04", end)
	if err1 != nil || err2 != nil {
		fmt.Printf("Error parsing time window: %v, %v\n", err1, err2)
		return true // Default to true on error
	}

	// Set dates to today for comparison
	current := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)
	startT := time.Date(0, 1, 1, s.Hour(), s.Minute(), 0, 0, time.UTC)
	endT := time.Date(0, 1, 1, e.Hour(), e.Minute(), 0, 0, time.UTC)

	if startT.After(endT) { // Crosses midnight
		return current.After(startT) || current.Before(endT) || current.Equal(startT) || current.Equal(endT)
	}
	return (current.After(startT) || current.Equal(startT)) && (current.Before(endT) || current.Equal(endT))
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "Unknown"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "Unknown"
}

func checkAndAlert() {
	if !isTimeInWindow(config.AlertStartTime, config.AlertEndTime) {
		fmt.Println("Current time is outside alerting window. Skipping check.")
		return
	}

	var alertMsg string

	// CPU Usage
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		fmt.Printf("Error getting CPU percent: %v\n", err)
	} else if len(percent) > 0 {
		val := percent[0]
		fmt.Printf("CPU Usage: %.2f%%\n", val)
		if val > config.CpuThreshold {
			alertMsg += fmt.Sprintf("CPU Usage High: %.2f%% (Threshold: %.2f%%)\n", val, config.CpuThreshold)
		}
	}

	// Memory Usage
	v, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("Error getting virtual memory: %v\n", err)
	} else {
		totalGB := float64(v.Total) / 1024 / 1024 / 1024
		usedGB := float64(v.Used) / 1024 / 1024 / 1024
		availableGB := float64(v.Available) / 1024 / 1024 / 1024

		fmt.Printf("Memory Usage: %.2f%% (Total: %.2f GB, Used: %.2f GB, Available: %.2f GB)\n", v.UsedPercent, totalGB, usedGB, availableGB)
		if v.UsedPercent > config.MemThreshold {
			alertMsg += fmt.Sprintf("Memory Usage High: %.2f%% (Threshold: %.2f%%, Total: %.2f GB, Used: %.2f GB, Available: %.2f GB)\n", v.UsedPercent, config.MemThreshold, totalGB, usedGB, availableGB)
		}
	}

	// Disk Usage
	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:"
	}
	usage, err := disk.Usage(path)
	if err != nil {
		fmt.Printf("Error getting disk usage: %v\n", err)
	} else {
		totalGB := float64(usage.Total) / 1024 / 1024 / 1024
		usedGB := float64(usage.Used) / 1024 / 1024 / 1024
		freeGB := float64(usage.Free) / 1024 / 1024 / 1024

		fmt.Printf("Disk Usage (%s): %.2f%% (Total: %.2f GB, Used: %.2f GB, Free: %.2f GB)\n", path, usage.UsedPercent, totalGB, usedGB, freeGB)
		if usage.UsedPercent > config.DiskThreshold {
			alertMsg += fmt.Sprintf("Disk Usage High (%s): %.2f%% (Threshold: %.2f%%, Total: %.2f GB, Used: %.2f GB, Free: %.2f GB)\n", path, usage.UsedPercent, config.DiskThreshold, totalGB, usedGB, freeGB)
		}
	}

	// Send Alert if needed
	if alertMsg != "" {
		nowStr := time.Now().Format("2006-01-02 15:04:05")
		ip := getLocalIP()
		header := fmt.Sprintf("Time: %s\nIP: %s\n", nowStr, ip)
		finalMsg := header + alertMsg

		fmt.Println("Sending Alert...")
		err := dingtalk.SendDingTalkWarning(finalMsg, config.AtMobiles, false)
		if err != nil {
			fmt.Printf("Failed to send alert: %v\n", err)
		} else {
			fmt.Println("Alert sent successfully.")
		}
	}
	fmt.Println("----------------------------------------")
}
