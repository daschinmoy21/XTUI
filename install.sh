#!/bin/bash

# Install Script for Xtui (System-Wide Installation)

# Variables
PROGRAM_NAME="xtui"
INSTALL_DIR="/usr/local/bin"
ASSETS_DIR="/usr/local/share/xtui" # Directory for assets like ASCII art
GO_DEPENDENCIES=(
  "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/bubbles/textinput"
  "github.com/charmbracelet/lipgloss"
  "github.com/mattn/go-sqlite3"
  "github.com/joho/godotenv" # Added for .env support
)

# Function to install a dependency if it's not already installed
install_dependency() {
  local dependency=$1
  local install_command=$2

  if ! command -v $dependency &>/dev/null; then
    echo "$dependency is not installed. Installing..."
    eval $install_command
    if [[ $? -ne 0 ]]; then
      echo "Error: Failed to install $dependency."
      exit 1
    fi
  else
    echo "$dependency is already installed."
  fi
}

# Detect the OS
OS=$(uname -s)
echo "Detected OS: $OS"

# Check if the OS is Linux
if [[ "$OS" != "Linux" ]]; then
  echo "Error: This script currently supports Linux only."
  exit 1
fi

# Detect the Linux distribution
if [[ -f /etc/os-release ]]; then
  source /etc/os-release
  DISTRO=$ID
else
  echo "Error: Unable to detect Linux distribution."
  exit 1
fi

echo "Detected Linux distribution: $DISTRO"

# Check if the system uses pacman (Arch-based)
if command -v pacman &>/dev/null; then
  echo "Detected Arch-based distribution."
  PACKAGE_MANAGER="pacman"
elif [[ "$DISTRO" == "ubuntu" || "$DISTRO" == "debian" ]]; then
  PACKAGE_MANAGER="apt-get"
elif [[ "$DISTRO" == "fedora" ]]; then
  PACKAGE_MANAGER="dnf"
else
  echo "Error: Unsupported Linux distribution."
  exit 1
fi

# Check if Go is installed
if ! command -v go &>/dev/null; then
  echo "Go is not installed. Installing Go..."
  case $PACKAGE_MANAGER in
  apt-get)
    sudo $PACKAGE_MANAGER update
    sudo $PACKAGE_MANAGER install -y golang
    ;;
  dnf)
    sudo $PACKAGE_MANAGER install -y golang
    ;;
  pacman)
    sudo $PACKAGE_MANAGER -S --noconfirm go
    ;;
  esac
  if [[ $? -ne 0 ]]; then
    echo "Error: Failed to install Go."
    exit 1
  fi
else
  echo "Go is already installed."
fi

# Check if SQLite3 is installed
case $PACKAGE_MANAGER in
apt-get)
  install_dependency "sqlite3" "sudo $PACKAGE_MANAGER install -y sqlite3"
  ;;
dnf)
  install_dependency "sqlite3" "sudo $PACKAGE_MANAGER install -y sqlite"
  ;;
pacman)
  install_dependency "sqlite3" "sudo $PACKAGE_MANAGER -S --noconfirm sqlite"
  ;;
esac

# Install Go dependencies
echo "Installing Go dependencies..."
for dep in "${GO_DEPENDENCIES[@]}"; do
  echo "Installing $dep..."
  go get $dep
  if [[ $? -ne 0 ]]; then
    echo "Error: Failed to install $dep."
    exit 1
  fi
done

# Build the program
echo "Building $PROGRAM_NAME..."
go build -o $PROGRAM_NAME

if [[ $? -ne 0 ]]; then
  echo "Error: Failed to build $PROGRAM_NAME."
  exit 1
fi

# Create the assets directory
echo "Creating assets directory at $ASSETS_DIR..."
sudo mkdir -p $ASSETS_DIR

# Copy the assets folder
echo "Copying assets folder to $ASSETS_DIR..."
if [[ -d "assets" ]]; then
  sudo cp -r assets/* $ASSETS_DIR/
  if [[ $? -ne 0 ]]; then
    echo "Error: Failed to copy assets folder."
    exit 1
  fi
else
  echo "Error: Assets folder 'assets' not found in the current directory."
  exit 1
fi

# Create the .env file
echo "Creating .env file at $ASSETS_DIR..."
sudo bash -c "cat <<EOL > $ASSETS_DIR/.env
DATABASE_PATH=$ASSETS_DIR/tui-do.db
ASCII_ART_PATH=$ASSETS_DIR/faqs_ascii.txt
EOL"

if [[ $? -ne 0 ]]; then
  echo "Error: Failed to create .env file."
  exit 1
fi

# Install the program
echo "Installing $PROGRAM_NAME to $INSTALL_DIR..."
sudo mv $PROGRAM_NAME $INSTALL_DIR/

if [[ $? -ne 0 ]]; then
  echo "Error: Failed to install $PROGRAM_NAME."
  exit 1
fi

# Set executable permissions
sudo chmod +x $INSTALL_DIR/$PROGRAM_NAME

echo "Installation complete! Run '$PROGRAM_NAME' to start the program."
