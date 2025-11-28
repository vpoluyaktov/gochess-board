#!/bin/bash
# Cleanup script for engine log files that may have been created before the fix

# Remove log files from repository root
cd "$(dirname "$0")/.." || exit 1

echo "Cleaning up engine log files from repository root..."
rm -f game.* log.* 2>/dev/null

# Count remaining files
count=$(ls -1 game.* log.* 2>/dev/null | wc -l)

if [ "$count" -eq 0 ]; then
    echo "✓ All engine log files cleaned up"
else
    echo "⚠ Warning: $count log files still present"
fi

# Show temp directory location
temp_dir="/tmp/gochess-board-engines"
if [ -d "$temp_dir" ]; then
    log_count=$(ls -1 "$temp_dir"/game.* "$temp_dir"/log.* 2>/dev/null | wc -l)
    echo "Engine logs are now being written to: $temp_dir"
    echo "  ($log_count log files currently in temp directory)"
else
    echo "Engine temp directory will be created at: $temp_dir"
fi
