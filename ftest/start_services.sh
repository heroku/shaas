#!/bin/bash

# Build the shaas binary
echo "Building the shaas binary..."
go build -o shaas ./shaas.go
if [ $? -ne 0 ]; then
  echo "Failed to build the shaas binary. Exiting."
  exit 1
fi
echo "Shaas binary built successfully."

echo "Starting services..."

# Start each service and capture its PID
nohup ./shaas --port 5001 --basic-auth user:pass > /dev/null 2>&1 & # Start service on port 5001
PID1=$!
echo "Service on port 5001 started with PID: $PID1"

nohup ./shaas --port 5002 --readonly > /dev/null 2>&1 &  # Start service on port 5002 (readonly)
PID2=$!
echo "Service on port 5002 started with PID: $PID2"

nohup ./shaas --port 5003 > /dev/null 2>&1 & # Start service on port 5003
PID3=$!
echo "Service on port 5003 started with PID: $PID3"

# Save PIDs to a file
echo "$PID1" > service_pids.txt
echo "$PID2" >> service_pids.txt
echo "$PID3" >> service_pids.txt

echo "Services started. PIDs saved to service_pids.txt."
