#!/bin/bash

# Screenshot MCP Server - Example Usage
# This script demonstrates various ways to use the screenshot server

SERVER_URL="http://localhost:8080"

echo "Screenshot MCP Server - Example Usage"
echo "======================================"

# Check if server is running
echo "1. Checking server health..."
curl -s "${SERVER_URL}/health" | jq '.' || echo "Server not running. Please start it first with: make run-server"

echo ""
echo "2. Taking screenshot by window title..."
curl -X POST "${SERVER_URL}/v1/screenshot" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "title",
    "target": "Notepad",
    "format": "png",
    "quality": 95
  }' | jq '.success, .width, .height, .format'

echo ""
echo "3. Taking screenshot using GET method..."
curl -s "${SERVER_URL}/v1/screenshot?method=title&target=Calculator&format=png" | jq '.success, .metadata.processing_time'

echo ""
echo "4. Listing Chrome instances..."
curl -s "${SERVER_URL}/v1/chrome/instances" | jq '.count, .instances[0].pid // "No Chrome instances found"'

echo ""
echo "5. Listing Chrome tabs..."
curl -s "${SERVER_URL}/v1/chrome/tabs" | jq '.count, .tabs[0].title // "No tabs found"'

echo ""
echo "6. Using MCP JSON-RPC to take screenshot..."
curl -X POST "${SERVER_URL}/rpc" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "screenshot.capture",
    "params": {
      "method": "title",
      "target": "Visual Studio Code",
      "format": "png",
      "include_cursor": false
    },
    "id": 1
  }' | jq '.result.success // .error'

echo ""
echo "7. Discovering Chrome instances via MCP..."
curl -X POST "${SERVER_URL}/rpc" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "chrome.instances",
    "params": {},
    "id": 2
  }' | jq '.result.count // .error'

echo ""
echo "8. Getting Chrome tabs via MCP..."
curl -X POST "${SERVER_URL}/rpc" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "chrome.tabs",
    "params": {},
    "id": 3
  }' | jq '.result.count // .error'

echo ""
echo "Example usage completed!"
echo ""
echo "For more examples, check out:"
echo "  - CLI usage: ./bin/mcpctl.exe --help"
echo "  - API documentation: ${SERVER_URL}/docs"
echo "  - Chrome tab screenshots: Use the tab ID from chrome.tabs response"