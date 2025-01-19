#!/bin/bash
echo "Stopping services..."
if [ -f service_pids.txt ]; then
  while read -r pid; do
    echo "Stopping PID: $pid"
    kill "$pid" 2>/dev/null
  done < service_pids.txt
  rm -f service_pids.txt
  echo "Services stopped."
else
  echo "No running services found."
fi
