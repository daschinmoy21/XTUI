#!/bin/bash

# Uninstall Script for Xtui (System-Wide Installation)

# Variables
PROGRAM_NAME="xtui"
INSTALL_DIR="/usr/local/bin"

# Check if the program is installed
if [[ ! -f $INSTALL_DIR/$PROGRAM_NAME ]]; then
  echo "Error: $PROGRAM_NAME is not installed in $INSTALL_DIR."
  exit 1
fi

# Uninstall the program
echo "Uninstalling $PROGRAM_NAME from $INSTALL_DIR..."
sudo rm $INSTALL_DIR/$PROGRAM_NAME

if [[ $? -ne 0 ]]; then
  echo "Error: Failed to uninstall $PROGRAM_NAME."
  exit 1
fi

echo "Uninstallation complete! $PROGRAM_NAME has been removed."
