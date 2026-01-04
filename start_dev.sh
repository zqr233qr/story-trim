#!/bin/bash

# 杀死旧的进程 (可选，慎用)
# pkill -f "cmd/server/main.go"

echo "🚀 Starting Backend Server..."
go run cmd/server/main.go > server.log 2>&1 &
SERVER_PID=$!
echo "Backend PID: $SERVER_PID"

echo "🚀 Starting Frontend..."
cd web && npm run dev &
WEB_PID=$!
echo "Frontend PID: $WEB_PID"

echo "✅ Both services started."
echo "Press Ctrl+C to stop."

trap "kill $SERVER_PID $WEB_PID; exit" INT TERM

wait