# Opening Database Update Script

## Overview

The `update-openings.sh` script provides an easy way to download and update the chess opening database from the [lichess-org/chess-openings](https://github.com/lichess-org/chess-openings) repository.

## Usage

### Basic Usage

```bash
# Download using the default commit
./update-openings.sh
```

### Specify a Commit

```bash
# Download from a specific commit
./update-openings.sh 3379fcd843d3aa24425a0aef51dc00cfe28a8071
```

### Get Help

```bash
./update-openings.sh --help
```

## What It Does

1. **Downloads 5 TSV files** from the lichess-org/chess-openings repository:
   - `a.tsv` - Openings starting with 'A' ECO codes
   - `b.tsv` - Openings starting with 'B' ECO codes
   - `c.tsv` - Openings starting with 'C' ECO codes
   - `d.tsv` - Openings starting with 'D' ECO codes
   - `e.tsv` - Openings starting with 'E' ECO codes

2. **Saves files** to `server/assets/openings/` directory

3. **Shows statistics**:
   - Number of files downloaded
   - Total size
   - Approximate number of openings

4. **Suggests verification** by running tests

## Output Example

```
================================================
Chess Opening Database Updater
================================================
Repository: lichess-org/chess-openings
Commit:     3379fcd843d3aa24425a0aef51dc00cfe28a8071
Output:     server/assets/openings
================================================

Downloading a.tsv...
  ✓ Downloaded a.tsv (64K)
Downloading b.tsv...
  ✓ Downloaded b.tsv (76K)
Downloading c.tsv...
  ✓ Downloaded c.tsv (128K)
Downloading d.tsv...
  ✓ Downloaded d.tsv (60K)
Downloading e.tsv...
  ✓ Downloaded e.tsv (40K)

================================================
Summary
================================================
Files downloaded: 5
Total size:       368K

Total openings:   ~3594

✓ Opening database updated successfully!

To verify the update, run:
  go test -v -run TestOpeningBook ./server
```

## After Updating

### 1. Verify the Update

```bash
go test -v -run TestOpeningBook ./server
```

### 2. Rebuild the Application

```bash
go build -o gochess-board .
```

The opening database is embedded in the binary using Go's `embed` package, so you need to rebuild after updating.

### 3. Run the Application

```bash
./gochess-board
```

The new opening database will be loaded at startup.

## Finding the Latest Commit

To find the latest commit hash from the lichess-org/chess-openings repository:

1. Visit: https://github.com/lichess-org/chess-openings
2. Click on the commit history
3. Copy the full commit hash (40 characters)
4. Use it with the script: `./update-openings.sh <commit-hash>`

## Troubleshooting

### Script Fails to Download

- **Check internet connection**: The script requires internet access
- **Verify commit hash**: Make sure the commit hash is valid
- **GitHub rate limiting**: If you get 403 errors, wait a few minutes

### Files Not Found After Update

- **Check directory**: Files should be in `server/assets/openings/`
- **Permissions**: Make sure the script has write permissions
- **Run from root**: Always run the script from the repository root

### Application Doesn't Load New Data

- **Rebuild required**: The data is embedded, so you must rebuild
- **Check logs**: Look at `gochess.log` for loading errors
- **Verify files**: Run `ls -lh server/assets/openings/` to check files exist

## Technical Details

### File Format

The TSV files have three columns:
- **eco**: ECO code (e.g., "C50")
- **name**: Opening name (e.g., "Italian Game")
- **pgn**: PGN notation of moves (e.g., "1. e4 e5 2. Nf3 Nc6 3. Bc4")

### Data Structure

The application loads these files into an in-memory trie structure:
- **Loading time**: ~7-8 seconds at startup
- **Memory usage**: ~1-2 MB
- **Lookup time**: Microseconds
- **Total openings**: 3,594
- **Max depth**: 36 moves

## License

The opening data from lichess-org/chess-openings is in the Public Domain.
