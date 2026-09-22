import subprocess
import json

def run():
    p = subprocess.Popen(['D:/projects/sys-mcp/sys-mcp.exe'], stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True)
    
    # Initialize
    init_req = {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}
    p.stdin.write(json.dumps(init_req) + "\n")
    p.stdin.flush()
    print("Init response:", p.stdout.readline().strip())
    
    # Initialized notification
    init_notif = {"jsonrpc": "2.0", "method": "notifications/initialized"}
    p.stdin.write(json.dumps(init_notif) + "\n")
    p.stdin.flush()
    
    # Call tool get_memory_usage
    tool_req = {"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "get_memory_usage", "arguments": {}}}
    p.stdin.write(json.dumps(tool_req) + "\n")
    p.stdin.flush()
    
    print("Memory response:", p.stdout.readline().strip())
    
    # Call tool get_network_info
    tool_req2 = {"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "get_network_info", "arguments": {}}}
    p.stdin.write(json.dumps(tool_req2) + "\n")
    p.stdin.flush()
    
    print("Network response:", p.stdout.readline().strip())
    
    p.kill()

if __name__ == "__main__":
    run()
