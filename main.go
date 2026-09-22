package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	s := server.NewMCPServer("SystemInfo-Fast", "2.0.0")

	// --- CPU Tool ---
	cpuTool := mcp.NewTool("get_cpu_info",
		mcp.WithDescription("Returns detailed CPU usage, cores, and model info."),
	)
	s.AddTool(cpuTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		info, err := cpu.Info()
		var result strings.Builder
		if err == nil && len(info) > 0 {
			result.WriteString(fmt.Sprintf("Model: %s\n", info[0].ModelName))
			result.WriteString(fmt.Sprintf("Cores: %d\n", info[0].Cores))
		}
		
		percentages, err := cpu.Percent(time.Second, true)
		if err == nil {
			result.WriteString("CPU Usage per core:\n")
			var total float64
			for i, p := range percentages {
				result.WriteString(fmt.Sprintf(" Core %d: %.2f%%\n", i, p))
				total += p
			}
			if len(percentages) > 0 {
				result.WriteString(fmt.Sprintf("\nTotal CPU Usage: %.2f%%\n", total/float64(len(percentages))))
			}
		}

		if result.Len() == 0 {
			return mcp.NewToolResultError("No CPU data available"), nil
		}
		return mcp.NewToolResultText(result.String()), nil
	})

	// --- Memory Tool ---
	memTool := mcp.NewTool("get_memory_usage",
		mcp.WithDescription("Returns current RAM and Swap usage info."),
	)
	s.AddTool(memTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var result strings.Builder
		v, err := mem.VirtualMemory()
		if err == nil {
			gbTotal := float64(v.Total) / (1024 * 1024 * 1024)
			gbUsed := float64(v.Used) / (1024 * 1024 * 1024)
			result.WriteString(fmt.Sprintf("RAM Used: %.2f GB / Total: %.2f GB (%.2f%%)\n", gbUsed, gbTotal, v.UsedPercent))
		}
		
		s, err := mem.SwapMemory()
		if err == nil {
			gbTotal := float64(s.Total) / (1024 * 1024 * 1024)
			gbUsed := float64(s.Used) / (1024 * 1024 * 1024)
			result.WriteString(fmt.Sprintf("Swap Used: %.2f GB / Total: %.2f GB (%.2f%%)\n", gbUsed, gbTotal, s.UsedPercent))
		}

		if result.Len() == 0 {
			return mcp.NewToolResultError("No Memory data available"), nil
		}
		return mcp.NewToolResultText(result.String()), nil
	})

	// --- System / Host Tool ---
	infoTool := mcp.NewTool("get_system_info",
		mcp.WithDescription("Returns basic system information (OS, Uptime, Temperature if supported)."),
	)
	s.AddTool(infoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		hInfo, err := host.Info()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := fmt.Sprintf("OS: %v (%v %v)\nUptime: %v seconds\nHostname: %v\n", hInfo.OS, hInfo.Platform, hInfo.PlatformVersion, hInfo.Uptime, hInfo.Hostname)

		temps, err := host.SensorsTemperatures()
		if err == nil && len(temps) > 0 {
			result += "\nTemperatures:\n"
			for _, t := range temps {
				result += fmt.Sprintf(" - %v: %.2f°C\n", t.SensorKey, t.Temperature)
			}
		} else {
			result += "\nTemperatures: Not supported natively on this OS by gopsutil without external tools."
		}

		return mcp.NewToolResultText(result), nil
	})

	// --- Disk Tool ---
	diskTool := mcp.NewTool("get_disk_info",
		mcp.WithDescription("Returns information about disk partitions and usage."),
	)
	s.AddTool(diskTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		partitions, err := disk.Partitions(false)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var result strings.Builder
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				continue
			}
			gbTotal := float64(usage.Total) / (1024 * 1024 * 1024)
			gbUsed := float64(usage.Used) / (1024 * 1024 * 1024)
			result.WriteString(fmt.Sprintf("Partition %s (%s): Used: %.2f GB / Total: %.2f GB (%.2f%%)\n", p.Mountpoint, p.Fstype, gbUsed, gbTotal, usage.UsedPercent))
		}

		if result.Len() == 0 {
			return mcp.NewToolResultError("No disk data available"), nil
		}
		return mcp.NewToolResultText(result.String()), nil
	})

	// --- Network Tool ---
	netTool := mcp.NewTool("get_network_info",
		mcp.WithDescription("Returns network interfaces and IO statistics."),
	)
	s.AddTool(netTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		interfaces, err := net.Interfaces()
		var result strings.Builder
		if err == nil {
			result.WriteString("Interfaces:\n")
			for _, iface := range interfaces {
				if len(iface.Addrs) > 0 {
					var addrs []string
					for _, a := range iface.Addrs {
						addrs = append(addrs, a.Addr)
					}
					result.WriteString(fmt.Sprintf(" - %s: %s\n", iface.Name, strings.Join(addrs, ", ")))
				}
			}
		}

		ioStats, err := net.IOCounters(false)
		if err == nil && len(ioStats) > 0 {
			result.WriteString("\nGlobal IO Stats:\n")
			stat := ioStats[0]
			mbSent := float64(stat.BytesSent) / (1024 * 1024)
			mbRecv := float64(stat.BytesRecv) / (1024 * 1024)
			result.WriteString(fmt.Sprintf(" Sent: %.2f MB\n Recv: %.2f MB\n", mbSent, mbRecv))
		}

		if result.Len() == 0 {
			return mcp.NewToolResultError("No network data available"), nil
		}
		return mcp.NewToolResultText(result.String()), nil
	})

	// --- Processes Tool ---
	procTool := mcp.NewTool("get_top_processes",
		mcp.WithDescription("Returns the top 10 processes by memory usage."),
	)
	s.AddTool(procTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		procs, err := process.Processes()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		type ProcInfo struct {
			PID int32
			Name string
			MemPercent float32
			CPUPercent float64
		}

		var procList []ProcInfo
		for _, p := range procs {
			name, err := p.Name()
			if err != nil || name == "" {
				continue
			}
			memP, _ := p.MemoryPercent()
			cpuP, _ := p.CPUPercent()
			
			procList = append(procList, ProcInfo{
				PID: p.Pid,
				Name: name,
				MemPercent: memP,
				CPUPercent: cpuP,
			})
		}

		sort.Slice(procList, func(i, j int) bool {
			return procList[i].MemPercent > procList[j].MemPercent
		})

		var result strings.Builder
		result.WriteString("Top 10 Processes by Memory:\n")
		result.WriteString(fmt.Sprintf("%-8s %-25s %-10s %-10s\n", "PID", "NAME", "MEM%", "CPU%"))
		limit := 10
		if len(procList) < limit {
			limit = len(procList)
		}
		for i := 0; i < limit; i++ {
			p := procList[i]
			// Trim name if it's too long
			name := p.Name
			if len(name) > 23 {
				name = name[:20] + "..."
			}
			result.WriteString(fmt.Sprintf("%-8d %-25s %-10.2f %-10.2f\n", p.PID, name, p.MemPercent, p.CPUPercent))
		}

		return mcp.NewToolResultText(result.String()), nil
	})

	// --- GPU Tool ---
	gpuTool := mcp.NewTool("get_gpu_info",
		mcp.WithDescription("Attempts to fetch GPU info using nvidia-smi."),
	)
	s.AddTool(gpuTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Try nvidia-smi
		cmd := exec.Command("nvidia-smi", "--query-gpu=name,temperature.gpu,utilization.gpu,utilization.memory,memory.total,memory.free,memory.used", "--format=csv,noheader")
		out, err := cmd.Output()
		if err != nil {
			return mcp.NewToolResultText("GPU Info not available via nvidia-smi (Make sure it's installed or you have an NVIDIA GPU)."), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("NVIDIA GPU Info (Name, Temp, Util, Mem Util, Mem Total, Mem Free, Mem Used):\n%s", string(out))), nil
	})

	// Run
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
