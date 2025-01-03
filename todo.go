package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/joho/godotenv"      // Load .env file
)

const (
	Tasks = iota
	User
	About
	LoadingScreen
)

const (
	normalMode = "normal"
	insertMode = "insert"
	undoLimit  = 10 // Limit for undo stack
)

type model struct {
	currentView int
	width       int
	height      int
	loadingDone bool
	tasksModel  tasksModel
	undoStack   []item // Stack to store deleted tasks for undo functionality
	db          *sql.DB
}

type tasksModel struct {
	items    []item
	input    textinput.Model
	selected int
	mode     string
}

type item struct {
	id          int
	title       string
	tags        []string
	status      status
	selected    bool
	createdAt   time.Time // Timestamp for task creation
	completedAt time.Time // Timestamp for task completion
}

type status int

const (
	todo status = iota
	done
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(4).
				Foreground(lipgloss.Color("#FFA500")) // Orange color for hover

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			Padding(1, 2) // Add padding to make tabs appear larger

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(1, 2) // Add padding to make tabs appear larger

	modeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF69B4"))

	loadingTextStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Align(lipgloss.Center).
				Margin(2, 0).
				Padding(1, 0)
)

func newModel() model {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		os.Exit(1)
	}

	// Get database path from .env
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./tui-do.db" // Default value
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database opened successfully.")

	// Ping the database to ensure the connection is valid
	err = db.Ping()
	if err != nil {
		fmt.Printf("Error pinging database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database connection is valid.")

	// Create the tasks table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			tags TEXT,
			status INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME
		);
	`)
	if err != nil {
		fmt.Printf("Error creating table: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Table 'tasks' created or already exists.")

	return model{
		currentView: LoadingScreen,
		tasksModel:  newTasksModel(),
		undoStack:   []item{},
		db:          db,
	}
}

func newTasksModel() tasksModel {
	ti := textinput.New()
	ti.Placeholder = "Press enter to add a new todo..."
	return tasksModel{
		items: []item{},
		input: ti,
		mode:  normalMode,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		func() tea.Msg {
			if m.currentView == LoadingScreen {
				time.Sleep(2 * time.Second)
				return "loading-done"
			}
			return nil
		},
		tick(), // Start the ticker
		m.loadTasks(), // Load tasks from the database
	)
}

func (m model) loadTasks() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query("SELECT id, title, tags, status, created_at, completed_at FROM tasks")
		if err != nil {
			fmt.Printf("Error loading tasks: %v\n", err)
			return nil
		}
		defer rows.Close()

		var tasks []item
		for rows.Next() {
			var task item
			var tags string
			var completedAt sql.NullTime
			err := rows.Scan(&task.id, &task.title, &tags, &task.status, &task.createdAt, &completedAt)
			if err != nil {
				fmt.Printf("Error scanning task: %v\n", err)
				continue
			}
			if completedAt.Valid {
				task.completedAt = completedAt.Time
			}
			if tags != "" {
				task.tags = strings.Split(tags, ",")
			} else {
				task.tags = []string{}
			}
			tasks = append(tasks, task)
		}
		return tasks
	}
}

func (m model) saveTask(task item) error {
	tags := strings.Join(task.tags, ",")
	var completed interface{}
	if task.status == done {
		completed = task.completedAt
	} else {
		completed = nil
	}
	_, err := m.db.Exec(`
		INSERT INTO tasks (title, tags, status, created_at, completed_at)
		VALUES (?, ?, ?, ?, ?)
	`, task.title, tags, task.status, task.createdAt, completed)
	return err
}

func (m model) updateTask(task item) error {
	tags := strings.Join(task.tags, ",")
	var completed interface{}
	if task.status == done {
		completed = task.completedAt
	} else {
		completed = nil
	}
	_, err := m.db.Exec(`
		UPDATE tasks
		SET title = ?, tags = ?, status = ?, completed_at = ?
		WHERE id = ?
	`, task.title, tags, task.status, completed, task.id)
	return err
}

func (m model) deleteTask(id int) error {
	_, err := m.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.tasksModel.mode == normalMode {
			switch msg.String() {
			case "ctrl+c", "q":
				clearScreen()
				return m, tea.Quit
			case "l", "right": // Move to the next tab
				if m.currentView < About {
					m.currentView++
				}
			case "h", "left": // Move to the previous tab
				if m.currentView > Tasks {
					m.currentView--
				}
			case "d":
				if len(m.tasksModel.items) > 0 {
					// Delete the selected task and push it to the undo stack
					deletedTask := m.tasksModel.items[m.tasksModel.selected]
					if len(m.undoStack) >= undoLimit {
						// Remove the oldest item if the stack exceeds the limit
						m.undoStack = m.undoStack[1:]
					}
					m.undoStack = append(m.undoStack, deletedTask)
					err := m.deleteTask(deletedTask.id)
					if err != nil {
						fmt.Printf("Error deleting task: %v\n", err)
					}
					m.tasksModel.items = append(m.tasksModel.items[:m.tasksModel.selected], m.tasksModel.items[m.tasksModel.selected+1:]...)
					if len(m.tasksModel.items) == 0 {
						m.tasksModel.selected = 0 // Reset selected index if no tasks are left
					} else if m.tasksModel.selected >= len(m.tasksModel.items) {
						m.tasksModel.selected = len(m.tasksModel.items) - 1
					}
				}
			case "u":
				if len(m.undoStack) > 0 {
					// Undo the last deletion by restoring the task from the undo stack
					restoredTask := m.undoStack[len(m.undoStack)-1]
					err := m.saveTask(restoredTask)
					if err != nil {
						fmt.Printf("Error restoring task: %v\n", err)
					}
					m.tasksModel.items = append(m.tasksModel.items, restoredTask)
					m.undoStack = m.undoStack[:len(m.undoStack)-1]
					m.tasksModel.selected = len(m.tasksModel.items) - 1 // Select the restored task
				}
			}
		}

		if m.currentView == Tasks {
			if m.tasksModel.mode == normalMode {
				switch msg.String() {
				case "enter":
					m.tasksModel.mode = insertMode
					m.tasksModel.input.Focus()
					return m, textinput.Blink
				case "up", "k":
					if m.tasksModel.selected > 0 {
						m.tasksModel.selected--
					}
				case "down", "j":
					if m.tasksModel.selected < len(m.tasksModel.items)-1 {
						m.tasksModel.selected++
					}
				case " ":
					if len(m.tasksModel.items) > 0 && m.tasksModel.selected >= 0 && m.tasksModel.selected < len(m.tasksModel.items) {
						item := &m.tasksModel.items[m.tasksModel.selected]
						item.status = toggleStatus(item.status)
						if item.status == done {
							item.completedAt = time.Now() // Record completion time
						}
						err := m.updateTask(*item)
						if err != nil {
							fmt.Printf("Error updating task: %v\n", err)
						}
					}
				}
			} else {
				switch msg.String() {
				case "esc":
					m.tasksModel.mode = normalMode
					m.tasksModel.input.Blur()
					return m, nil
				case "enter":
					if m.tasksModel.input.Value() != "" {
						newItem := item{
							title:     removeTags(m.tasksModel.input.Value()),
							status:    todo,
							tags:      parseTags(m.tasksModel.input.Value()),
							createdAt: time.Now(), // Record creation time
						}
						err := m.saveTask(newItem)
						if err != nil {
							fmt.Printf("Error saving task: %v\n", err)
						}
						m.tasksModel.items = append(m.tasksModel.items, newItem)
						m.tasksModel.input.Reset()
						m.tasksModel.mode = normalMode
						m.tasksModel.input.Blur()
					}
				default:
					m.tasksModel.input, cmd = m.tasksModel.input.Update(msg)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case string:
		if msg == "loading-done" {
			m.loadingDone = true
			m.currentView = Tasks
		}

	case []item:
		m.tasksModel.items = msg

	case time.Time:
		// Triggered by the ticker, refresh the UI
		return m, tick()
	}

	return m, cmd
}

func (m model) View() string {
	if m.currentView == LoadingScreen && !m.loadingDone {
		// Define the loading text with "||" in orange and bold
		loadingText := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Render("XTUI") +
			lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")). // Orange color for "||"
			Render("||")

		// Center the loading text
		centeredLoadingText := lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			loadingText,
		)

		return centeredLoadingText
	}

	// Define tabs with larger appearance using padding
	tabs := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.tab("Tasks", Tasks),
		m.tab("User", User),
		m.tab("About", About),
	)

	var content string
	switch m.currentView {
	case Tasks:
		content = m.renderTasks()
	case User:
		content = "User info and account sign-in/creation status display for cloud sync\n(W.I.P)"
	case About:
		content = m.renderAbout()
	}

	footer := "\nPress 'h' and 'l' to switch tabs | space: toggle | enter: new task | d: delete | u: undo | q: quit"
	if m.tasksModel.mode == insertMode {
		footer = "\nesc: normal mode | enter: save task | #tag: add tag"
	}

	// Fixed height for tabs and centered content
	tabsHeight := 3 // Fixed height for tabs
	contentHeight := m.height - tabsHeight - 3 // Remaining height for content and footer

	// Center the content within the available space
	centeredContent := lipgloss.Place(
		m.width,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)

	// Center the tabs and footer horizontally
	centeredTabs := lipgloss.Place(
		m.width,
		tabsHeight,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.NewStyle().PaddingTop(2).Render(tabs), // Add padding above tabs
	)

	centeredFooter := lipgloss.Place(
		m.width,
		3, // Fixed height for footer
		lipgloss.Center,
		lipgloss.Center,
		helpStyle.Render(footer),
	)

	// Combine centered tabs, centered content, and centered footer
	body := fmt.Sprintf("%s\n%s\n%s",
		centeredTabs,
		centeredContent,
		centeredFooter,
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Padding(1, 2).Render(body),
	)
}

func (m model) renderTasks() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Accelerate,Anon") + "\n\n")

	for i, item := range m.tasksModel.items {
		// Fixed-width cursor (2 characters)
		cursor := "  " // Default to two spaces
		if i == m.tasksModel.selected {
			cursor = "▸ " // Right-pointing triangle followed by a space
		}

		// Fixed-width status marker (3 characters)
		statusMarker := "[ ]"
		if item.status == done {
			statusMarker = "[✓]"
		}

		// Align the task title
		itemText := fmt.Sprintf("%s %s %s", cursor, statusMarker, item.title)
		if i == m.tasksModel.selected {
			itemText = selectedItemStyle.Render(itemText)
		} else {
			itemText = itemStyle.Render(itemText)
		}
		s.WriteString(itemText)

		// Add tags if present
		if len(item.tags) > 0 {
			tags := fmt.Sprintf(" [%s]", strings.Join(item.tags, ", "))
			s.WriteString(tagStyle.Render(tags))
		}

		// Show "Completed" for done tasks, no timestamp
		if item.status == done {
			s.WriteString(" - Completed")
		} else {
			s.WriteString(fmt.Sprintf(" - Created %s", formatRelativeTime(item.createdAt)))
		}
		s.WriteString("\n")
	}

	if m.tasksModel.mode == insertMode {
		s.WriteString("\n" + m.tasksModel.input.View())
	}

	return s.String()
}

func (m model) renderAbout() string {
	// Get ASCII art path from .env
	asciiArtPath := os.Getenv("ASCII_ART_PATH")
	if asciiArtPath == "" {
		asciiArtPath = "faqs_ascii.txt" // Default value
	}

	// Read the ASCII image from the file
	asciiArt, err := os.ReadFile(asciiArtPath)
	if err != nil {
		return "Error loading ASCII art."
	}

	// Combine the ASCII image with the About content
	aboutText := `Xtui is a terminal based todo list app to get shit done.
Embrace the beauty of the terminal and get to work anon.
Built for speed, simplicity and the terminal in mind.
Only on Linux for now.
controls inspired by vim
built by @crimxnhaze on X`

	return fmt.Sprintf("%s\n\n%s", string(asciiArt), aboutText)
}

func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg {
		return t
	})
}

func (m model) tab(name string, section int) string {
	if m.currentView == section {
		return activeTabStyle.Render(name)
	}
	return inactiveTabStyle.Render(name)
}

func clearScreen() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func parseTags(input string) []string {
	var tags []string
	words := strings.Fields(input)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			tags = append(tags, word[1:])
		}
	}
	return tags
}

func removeTags(input string) string {
	words := strings.Fields(input)
	var result []string
	for _, word := range words {
		if !strings.HasPrefix(word, "#") {
			result = append(result, word)
		}
	}
	return strings.Join(result, " ")
}

func statusMarker(s status) string {
	if s == done {
		return "[✓]"
	}
	return "[ ]"
}

func toggleStatus(s status) status {
	if s == done {
		return todo
	}
	return done
}

func main() {
	p := tea.NewProgram(newModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting app: %v\n", err)
		os.Exit(1)
	}
}
