import subprocess
import json

def run():
    p = subprocess.Popen(['D:/projects/sys-mcp/sys-mcp.exe'], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

    # Initialize
    init_req = {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}
    p.stdin.write(json.dumps(init_req) + "\n")
    p.stdin.flush()
    p.stdout.readline()  # consume init response

    # Initialized notification
    init_notif = {"jsonrpc": "2.0", "method": "notifications/initialized"}
    p.stdin.write(json.dumps(init_notif) + "\n")
    p.stdin.flush()

    # Call get_system_info for temps
    tool_req = {"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "get_system_info", "arguments": {}}}
    p.stdin.write(json.dumps(tool_req) + "\n")
    p.stdin.flush()
    resp = p.stdout.readline().strip()
    data = json.loads(resp)
    if 'result' in data and 'content' in data['result']:
        for block in data['result']['content']:
            if 'text' in block:
                print(block['text'])
    else:
        print(resp)

    p.kill()

if __name__ == "__main__":
    run()
