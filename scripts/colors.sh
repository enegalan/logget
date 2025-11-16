#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COLORS_DEF="$PROJECT_ROOT/src/colors.def"

if [ ! -f "$COLORS_DEF" ]; then
	echo "Error: $COLORS_DEF not found" >&2
	exit 1
fi

convert_name() {
	local name=$1
	case "$name" in
		Reset)
			echo "NC"
			;;
		BrightRed)
			echo "BRIGHT_RED"
			;;
		BrightGreen)
			echo "BRIGHT_GREEN"
			;;
		BrightYellow)
			echo "BRIGHT_YELLOW"
			;;
		BrightBlue)
			echo "BRIGHT_BLUE"
			;;
		BrightMagenta)
			echo "BRIGHT_MAGENTA"
			;;
		BrightCyan)
			echo "BRIGHT_CYAN"
			;;
		BrightWhite)
			echo "BRIGHT_WHITE"
			;;
		*)
			echo "$name" | sed 's/\([A-Z]\)/_\1/g' | tr '[:lower:]' '[:upper:]' | sed 's/^_//'
			;;
	esac
}

while IFS='=' read -r key value || [ -n "$key" ]; do
	[[ -z "$key" || "$key" =~ ^[[:space:]]*# ]] && continue
	key="${key#"${key%%[![:space:]]*}"}"
	key="${key%"${key##*[![:space:]]}"}"
	value="${value#"${value%%[![:space:]]*}"}"
	value="${value%"${value##*[![:space:]]}"}"
	[[ -z "$key" || -z "$value" ]] && continue
	esc_char=$'\033'
	value="${value//\\033/$esc_char}"
	bash_name=$(convert_name "$key")
	export "$bash_name=$value"
done < "$COLORS_DEF"
