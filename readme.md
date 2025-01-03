# Xtui - Terminal-Based Todo List App

![Xtui Screenshot](assets/xtui_screenshot.png) <!-- Add a screenshot if available -->

**Xtui** is a fast, simple, and elegant terminal-based todo list app designed to help you get things done efficiently. Built with Go and leveraging the power of SQLite3, Xtui is lightweight, customizable, and perfect for anyone who loves working in the terminal.

---

## Features

- **Terminal-Based Interface**: Beautiful and intuitive TUI (Terminal User Interface) powered by [BubbleTea](https://github.com/charmbracelet/bubbletea).
- **Task Management**:
  - Add, delete, and mark tasks as done.
  - Undo deleted tasks (up to 10 actions).
  - Tag tasks for better organization (e.g., `#work`, `#personal`).
- **Persistent Storage**: Tasks are stored in an SQLite3 database for persistence across sessions.
- **Customizable**: Configure paths for assets and database using a `.env` file.
- **Cross-Platform**: Designed for Linux (other platforms may require minor adjustments).
- **Lightweight**: Minimal dependencies and fast performance.

---

## Installation

### Prerequisites

- **Go** (version 1.16 or higher)
- **SQLite3** (for database management)
- **Git** (for cloning the repository)

### Steps

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/yourusername/xtui.git
   cd xtui
# Xtui - Terminal User Interface for Task Management

## Installation

### Run the Install Script:

```bash
chmod +x install.sh
sudo ./install.sh
This script will:
```
Install dependencies (Go, SQLite3, and required Go packages).

Build the program.

Copy assets to /usr/local/share/xtui.

Create a .env file for configuration.

Install the xtui executable to /usr/local/bin.

Usage
Run the Program:
```bash
Copy
xtui
```
Keybindings
Key(s)	Action
h, left	Switch to the previous tab.
l, right	Switch to the next tab.
enter	Add a new task (in insert mode).
esc	Exit insert mode.
space	Toggle task status (todo/done).
d	Delete the selected task.
u	Undo the last deletion.
q, ctrl+c	Quit the program.
Tabs
Tasks: Manage your todo list.

User: (Work in Progress) User info and cloud sync status.

About: Learn more about Xtui.

Configuration
The .env file is created during installation and contains the following configuration:

```env
Copy
DATABASE_PATH=/usr/local/share/xtui/tui-do.db
ASCII_ART_PATH=/usr/local/share/xtui/faqs_ascii.txt
```
You can modify these paths if needed.

Project Structure
Copy
xtui/
├── assets/               # Contains assets like ASCII art
│   └── faqs_ascii.txt
├── main.go               # Main application code
├── install.sh            # Installation script
├── .env                  # Configuration file (created during installation)
└── README.md             # This file
Contributing
Contributions are welcome! Here’s how you can help:

Report Bugs: Open an issue on GitHub.

Suggest Features: Dynamic ASCII arts, cloud save features, bug fixes.

Submit Pull Requests: Fork the repository, make your changes, and submit a PR.

License
This project is licensed under the MIT License. See the LICENSE file for details.

Acknowledgments
BubbleTea for the TUI framework.

SQLite3 for lightweight database management.

Charmbracelet for inspiring terminal-based tools.

Screenshots
Tasks Tab

About Tab

Contact
For questions or feedback, reach out to:

Your Name - @yourusername

Email: your.email@example.com

GitHub: yourusername

Enjoy using Xtui! 🚀
