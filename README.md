# Sys MCP

A blazing-fast, native Go Model Context Protocol (MCP) server for deep system inspection. 

Provides real-time AI agents (like Antigravity or Claude) with comprehensive access to your machine's hardware metrics.

## Features

- **CPU Monitoring**: Per-core and total usage with real-time sliding window accuracy.
- **Memory & Swap**: Tracks both physical RAM and swap usage.
- **Network I/O**: Lists active interfaces, IP addresses, and total MBs sent/received.
- **Disk Tracking**: Automatically surveys all mounted partitions for capacity and usage.
- **Top Processes**: Ranks the top 10 running processes by memory usage.
- **GPU Stats**: Hooks into `nvidia-smi` to report VRAM, temperature, and utilization (NVIDIA only).

## Usage

1. Compile the native executable:
   ```bash
   go build -o sys-mcp.exe
   ```

2. Point your MCP client to the compiled executable. For example, in Antigravity's `mcp_config.json`:
   ```json
   {
     "mcpServers": {
       "sys-mcp": {
         "command": "C:/absolute/path/to/sys-mcp.exe",
         "args": []
       }
     }
   }
   ```

3. Ask your AI to read your system state!

## Built With

* Go 1.22+
* [mcp-go](https://github.com/mark3labs/mcp-go) - Go implementation of the Model Context Protocol
* [gopsutil](https://github.com/shirou/gopsutil) - Python's psutil ported to Go
