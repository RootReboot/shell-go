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

echo  "The Os of the pipeline machine is $OS"

install_readline_if_missing() {
  echo "🔍 Checking for GNU Readline..."

  echo "❌ GNU Readline not found. Attempting to install..."

  case "$OS" in
    Linux)
      if command -v apt-get >/dev/null 2>&1; then
        sudo apt-get update
        sudo apt-get install -y libreadline-dev pkg-config
      elif command -v pacman >/dev/null 2>&1; then
        sudo pacman -Sy --noconfirm readline pkgconf
      elif command -v apk >/dev/null 2>&1; then
        apk add --no-cache readline-dev
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

install_build_tools_if_missing() {
  echo "🔍 Checking for C compiler (gcc)..."

  if command -v gcc >/dev/null 2>&1; then
    echo "✅ gcc found."
    return
  fi

  echo "❌ gcc not found. Installing build tools..."

  OS=$(uname -s)
  case "$OS" in
    Linux*)
      if command -v apt-get >/dev/null 2>&1; then
        apt-get update
        apt-get install -y build-essential
      elif command -v apk >/dev/null 2>&1; then
        apk add --no-cache build-base
      else
        echo "❌ Unsupported Linux distro. Install gcc manually."
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
install_build_tools_if_missing
install_readline_if_missing



echo "Building Go project with CGO enabled..."
export CGO_ENABLED=1

# 🛠 Compile your shell
go build -o /tmp/codecrafters-build-shell-go ./app

