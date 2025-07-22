#!/bin/sh
#
# This script is used to compile your program on CodeCrafters
#
# This runs before .codecrafters/run.sh
#
# Learn more: https://codecrafters.io/program-interface

set -e # Exit on failure

# 🧠 Detect OS
OS="$(uname -s)"

install_readline_if_missing() {
  echo "🔍 Checking for GNU Readline..."

  if pkg-config --exists readline; then
    echo "✅ GNU Readline already installed"
    return
  fi

  echo "❌ GNU Readline not found. Attempting to install..."

  case "$OS" in
    Linux)
      if command -v apt-get >/dev/null 2>&1; then
        sudo apt-get update
        sudo apt-get install -y libreadline-dev pkg-config
      elif command -v pacman >/dev/null 2>&1; then
        sudo pacman -Sy --noconfirm readline pkgconf
      else
        echo "❌ Unsupported Linux distro. Install libreadline-dev manually."
        exit 1
      fi
      ;;
    *)
      echo "❌ Unsupported OS: $OS"
      exit 1
      ;;
  esac
}

# 🔧 Install Readline if needed
install_readline_if_missing



echo "Building Go project with CGO enabled..."
export CGO_ENABLED=1

# 🛠 Compile your shell
go build -o /tmp/codecrafters-build-shell-go app/*.go

