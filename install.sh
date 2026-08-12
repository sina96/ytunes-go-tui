#!/usr/bin/env bash
# Installs ytunes's dependencies (mpv, yt-dlp) via the system package
# manager, then builds and installs the ytunes binary itself.
set -euo pipefail

BINARY_NAME="ytunes"
INSTALL_DIR="${HOME}/.local/bin"

info()  { printf '\033[1;34m==>\033[0m %s\n' "$1"; }
warn()  { printf '\033[1;33m!!\033[0m %s\n' "$1"; }
error() { printf '\033[1;31mERROR:\033[0m %s\n' "$1" >&2; }
have()  { command -v "$1" >/dev/null 2>&1; }

install_macos() {
	if ! have brew; then
		error "Homebrew not found. Install it from https://brew.sh, then re-run this script."
		exit 1
	fi
	for pkg in mpv yt-dlp; do
		if have "$pkg"; then
			info "$pkg already installed"
		else
			info "Installing $pkg via Homebrew..."
			brew install "$pkg"
		fi
	done
}

install_linux() {
	local pm update
	if have apt-get; then
		pm="sudo apt-get install -y"; update="sudo apt-get update"
	elif have dnf; then
		pm="sudo dnf install -y"; update=":"
	elif have pacman; then
		pm="sudo pacman -S --noconfirm"; update=":"
	else
		error "No supported package manager found (apt, dnf, pacman)."
		error "Install mpv and yt-dlp manually, then re-run this script."
		exit 1
	fi

	$update
	for pkg in mpv yt-dlp; do
		if have "$pkg"; then
			info "$pkg already installed"
		else
			info "Installing $pkg..."
			$pm "$pkg"
		fi
	done
}

case "$(uname -s)" in
Darwin) install_macos ;;
Linux) install_linux ;;
*)
	error "Unsupported OS: $(uname -s)."
	error "Install mpv and yt-dlp manually, then run 'go build -o ${BINARY_NAME} .' yourself."
	exit 1
	;;
esac

if ! have go; then
	error "Go not found. Install it from https://go.dev/dl, then re-run this script."
	exit 1
fi

info "Building ${BINARY_NAME}..."
mkdir -p "$INSTALL_DIR"
go build -o "${INSTALL_DIR}/${BINARY_NAME}" .

case ":${PATH}:" in
*":${INSTALL_DIR}:"*) ;;
*)
	warn "${INSTALL_DIR} is not on your PATH. Add this to your shell profile:"
	warn "  export PATH=\"${INSTALL_DIR}:\$PATH\""
	;;
esac

info "Installed! Run it with: ${BINARY_NAME}"
