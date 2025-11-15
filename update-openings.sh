#!/bin/bash

# Script to download chess opening database from lichess-org/chess-openings
# Usage: ./update-openings.sh [commit-hash]

set -e

# Show help
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    echo "Chess Opening Database Updater"
    echo ""
    echo "Usage: $0 [commit-hash]"
    echo ""
    echo "Downloads chess opening TSV files from lichess-org/chess-openings repository."
    echo ""
    echo "Arguments:"
    echo "  commit-hash    Git commit hash to download from (optional)"
    echo "                 Default: 3379fcd843d3aa24425a0aef51dc00cfe28a8071"
    echo ""
    echo "Examples:"
    echo "  $0                                              # Use default commit"
    echo "  $0 3379fcd843d3aa24425a0aef51dc00cfe28a8071    # Use specific commit"
    echo ""
    echo "Files downloaded: a.tsv, b.tsv, c.tsv, d.tsv, e.tsv"
    echo "Output directory: server/assets/openings"
    exit 0
fi

# Configuration
REPO="lichess-org/chess-openings"
COMMIT="${1:-3379fcd843d3aa24425a0aef51dc00cfe28a8071}"  # Default commit or use argument
BASE_URL="https://raw.githubusercontent.com/${REPO}/${COMMIT}"
OUTPUT_DIR="server/assets/openings"

# TSV files to download
FILES=("a.tsv" "b.tsv" "c.tsv" "d.tsv" "e.tsv")

echo "================================================"
echo "Chess Opening Database Updater"
echo "================================================"
echo "Repository: ${REPO}"
echo "Commit:     ${COMMIT}"
echo "Output:     ${OUTPUT_DIR}"
echo "================================================"
echo ""

# Create output directory if it doesn't exist
mkdir -p "${OUTPUT_DIR}"

# Download each file
for file in "${FILES[@]}"; do
    url="${BASE_URL}/${file}"
    output="${OUTPUT_DIR}/${file}"
    
    echo "Downloading ${file}..."
    if curl -f -L -o "${output}" "${url}"; then
        size=$(du -h "${output}" | cut -f1)
        echo "  ✓ Downloaded ${file} (${size})"
    else
        echo "  ✗ Failed to download ${file}"
        exit 1
    fi
done

echo ""
echo "================================================"
echo "Summary"
echo "================================================"

# Show statistics
total_size=$(du -sh "${OUTPUT_DIR}" | cut -f1)
file_count=$(ls -1 "${OUTPUT_DIR}"/*.tsv 2>/dev/null | wc -l)

echo "Files downloaded: ${file_count}"
echo "Total size:       ${total_size}"
echo ""

# Count total openings (subtract header lines)
if command -v wc &> /dev/null; then
    total_lines=0
    for file in "${FILES[@]}"; do
        lines=$(wc -l < "${OUTPUT_DIR}/${file}")
        total_lines=$((total_lines + lines - 1))  # Subtract header
    done
    echo "Total openings:   ~${total_lines}"
fi

echo ""
echo "✓ Opening database updated successfully!"
echo ""
echo "To verify the update, run:"
echo "  go test -v -run TestOpeningBook ./server"
