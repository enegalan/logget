#!/bin/sh

case "$0" in
	/*)
		SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
		;;
	*)
		SCRIPT_DIR="$(cd "$(dirname "$(pwd)/$0")" && pwd)"
		;;
esac
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COLORS_DEF="$PROJECT_ROOT/src/colors/colors.def"

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
	case "$key" in
		"")
			continue
			;;
		\#*)
			continue
			;;
		*)
			key=$(echo "$key" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
			value=$(echo "$value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
			if [ -z "$key" ] || [ -z "$value" ]; then
				continue
			fi
			esc_char=$(printf '\033')
			value=$(echo "$value" | sed "s/\\\\033/$esc_char/g")
			bash_name=$(convert_name "$key")
			export "$bash_name=$value"
			;;
	esac
done < "$COLORS_DEF"
