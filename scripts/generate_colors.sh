#!/bin/bash

# Generate colors.sh from src/colors.go
# This ensures both files use the same color definitions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COLORS_GO="$PROJECT_ROOT/src/colors.go"
COLORS_SH="$SCRIPT_DIR/colors.sh"

if [ ! -f "$COLORS_GO" ]; then
    echo "Error: $COLORS_GO not found" >&2
    exit 1
fi

cat > "$COLORS_SH" << 'HEADER'
#!/bin/bash

# Auto-generated from src/colors.go
# DO NOT EDIT MANUALLY - Run 'make generate-colors' or './scripts/generate_colors.sh'

HEADER

# Extract constants from colors.go
# Match lines like: 	Red     = "\033[31m"
grep -E '^\s+(Reset|Black|Red|Green|Yellow|Blue|Magenta|Cyan|White|Gray|BrightRed|BrightGreen|BrightYellow|BrightBlue|BrightMagenta|BrightCyan|BrightWhite|Bold|Dim|Italic|Underline|Blink|Reverse|Strike)\s+=' "$COLORS_GO" | \
    sed -E 's/^[[:space:]]+([A-Za-z][A-Za-z0-9]*)[[:space:]]+=[[:space:]]+"([^"]+)".*/\1=\2/' | \
    while IFS='=' read -r name value; do
        case "$name" in
            Reset)
                bash_name="NC"
                ;;
            BrightRed)
                bash_name="BRIGHT_RED"
                ;;
            BrightGreen)
                bash_name="BRIGHT_GREEN"
                ;;
            BrightYellow)
                bash_name="BRIGHT_YELLOW"
                ;;
            BrightBlue)
                bash_name="BRIGHT_BLUE"
                ;;
            BrightMagenta)
                bash_name="BRIGHT_MAGENTA"
                ;;
            BrightCyan)
                bash_name="BRIGHT_CYAN"
                ;;
            BrightWhite)
                bash_name="BRIGHT_WHITE"
                ;;
            *)
                # Convert CamelCase to UPPER_CASE
                bash_name=$(echo "$name" | sed 's/\([A-Z]\)/_\1/g' | tr '[:lower:]' '[:upper:]' | sed 's/^_//')
                ;;
        esac
        echo "${bash_name}=\"${value}\"" >> "$COLORS_SH"
    done

chmod +x "$COLORS_SH"
echo "Generated $COLORS_SH from $COLORS_GO"
