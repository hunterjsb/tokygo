#!/bin/bash

# Stop the tokygo server if it's running

echo "Stopping tokygo server..."

# Find and kill the process
if pgrep -f "tokygo" > /dev/null; then
    pkill -f "tokygo"
    echo "Stopped existing tokygo process"

    # Wait for process to stop
    sleep 2

    # Force kill if still running
    if pgrep -f "tokygo" > /dev/null; then
        pkill -9 -f "tokygo"
        echo "Force killed tokygo process"
    fi
else
    echo "No tokygo process found"
fi

exit 0
