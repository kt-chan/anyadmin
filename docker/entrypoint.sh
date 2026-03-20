#!/bin/bash
# entrypoint.sh - Start script for AnyAdmin Docker container

set -e

# --- Configuration ---
LOG_DIR="/home/anyadmin/logs"
BACKEND_LOG="${LOG_DIR}/backend.log"
FRONTEND_LOG="${LOG_DIR}/frontend.log"

# Ensure log files exist and are writable
# (Running as 'anyadmin' user, so we just touch them)
touch "${BACKEND_LOG}" "${FRONTEND_LOG}"

echo ">>> Starting AnyAdmin Services..."

# --- Start Backend ---
echo ">>> Starting Backend (Go)..."
cd /home/anyadmin/app/backend
# Ensure backend binary is executable
chmod +x anyadmin-server
./anyadmin-server > "${BACKEND_LOG}" 2>&1 &
BACKEND_PID=$!
echo "Backend started with PID ${BACKEND_PID}"

# --- Start Frontend ---
echo ">>> Starting Frontend (Node.js)..."
cd /home/anyadmin/app/frontend
node app.js > "${FRONTEND_LOG}" 2>&1 &
FRONTEND_PID=$!
echo "Frontend started with PID ${FRONTEND_PID}"

# --- Logging & Health Monitoring ---
echo ">>> Monitoring services. Logs are being redirected to ${LOG_DIR}."

# Function to handle shutdown signals
cleanup() {
    echo ">>> Shutting down services..."
    kill -TERM "${BACKEND_PID}" 2>/dev/null
    kill -TERM "${FRONTEND_PID}" 2>/dev/null
    wait "${BACKEND_PID}" "${FRONTEND_PID}"
    exit 0
}

trap cleanup SIGINT SIGTERM

# Output logs to stdout for Docker logging
echo ">>> Tail -f logs to keep container alive and show output..."
tail -f "${BACKEND_LOG}" "${FRONTEND_LOG}" &
TAIL_PID=$!

# Wait for any process to exit
wait -n "${BACKEND_PID}" "${FRONTEND_PID}"

echo ">>> One of the services exited. Shutting down."
cleanup
