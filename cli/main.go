package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	  	"net/http"
	  	"os"
	  	"strings"
	  	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AuthError struct {
	message string
}

func (e AuthError) Error() string {
	return fmt.Sprintf("Authentication error: %s", e.message)
}

type NetworkError struct {
	message string
}

func (e NetworkError) Error() string {
	return fmt.Sprintf("Network error: %s", e.message)
}

type DatabaseError struct {
	message string
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("Database error: %s", e.message)
}

// Credential model for database
type Credential struct {
	gorm.Model
	Instance string `gorm:"uniqueIndex"`
	Username string
	Password string
}

	type prettyLogger struct {
	    fileLogger *log.Logger
	    styles    struct {
	        info      lipgloss.Style
	        error     lipgloss.Style
	        timestamp lipgloss.Style
	        instance  lipgloss.Style
	        username  lipgloss.Style
	    }
	}

	var db *gorm.DB
	var authLogger *prettyLogger

	func newPrettyLogger(fileLogger *log.Logger) *prettyLogger {
	    pl := &prettyLogger{
	        fileLogger: fileLogger,
	    }
	    
	    // Initialize styles
	    pl.styles.info = lipgloss.NewStyle().
	        Foreground(lipgloss.Color("86")).  // Cyan
	        PaddingLeft(1)
	    
	    pl.styles.error = lipgloss.NewStyle().
	        Foreground(lipgloss.Color("196")).  // Red
	        Bold(true).
	        PaddingLeft(1)
	    
	    pl.styles.timestamp = lipgloss.NewStyle().
	        Foreground(lipgloss.Color("240")).  // Dim white
	        Faint(true)
	    
	    pl.styles.instance = lipgloss.NewStyle().
	        Bold(true).
	        Foreground(lipgloss.Color("87"))    // Light cyan
	    
	    pl.styles.username = lipgloss.NewStyle().
	        Bold(true).
	        Foreground(lipgloss.Color("87"))    // Light cyan
	    
	    return pl
	}

	func (pl *prettyLogger) logf(level string, format string, args ...interface{}) {
	    // Format the message
	    msg := fmt.Sprintf(format, args...)
	    
	    // Log to file
	    pl.fileLogger.Printf("%s: %s", level, msg)
	    
	    // Format timestamp
	    timestamp := pl.styles.timestamp.Render(time.Now().Format("15:04:05"))
	    
	    // Style the message based on level
	    var styledMsg string
	    if level == "ERROR" {
	        styledMsg = pl.styles.error.Render(msg)
	    } else {
	        styledMsg = pl.styles.info.Render(msg)
	    }
	    
	    // Print to console
	    fmt.Printf("%s %s\n", timestamp, styledMsg)
	}

	func (pl *prettyLogger) Printf(format string, args ...interface{}) {
	    pl.logf("INFO", format, args...)
	}

	func (pl *prettyLogger) Errorf(format string, args ...interface{}) {
	    pl.logf("ERROR", format, args...)
	}

	func initDB() error {
	var err error
	db, err = gorm.Open(sqlite.Open("credentials.db"), &gorm.Config{})
	if err != nil {
		return DatabaseError{message: "failed to connect to database"}
	}

	err = db.AutoMigrate(&Credential{})
	if err != nil {
		return DatabaseError{message: "failed to migrate database"}
	}
	return nil
}

func saveCredentials(instance, username, password string) error {
	credential := Credential{
		Instance: instance,
		Username: username,
		Password: password,
	}

	  result := db.Where("instance = ?", instance).FirstOrCreate(&credential)
	  if result.Error != nil {
							authLogger.Errorf("Failed to save credentials for instance %s: %v", instance, result.Error)
	      return DatabaseError{message: "failed to save credentials"}
	  }
	  authLogger.Printf("INFO: Successfully saved credentials for instance %s", instance)
	  return nil
}

func getCredentials(instance string) (*Credential, error) {
	var credential Credential
	  result := db.Where("instance = ?", instance).First(&credential)
	  if result.Error != nil {
	      if errors.Is(result.Error, gorm.ErrRecordNotFound) {
	          authLogger.Printf("INFO: No credentials found for instance %s", instance)
	          return nil, nil
	      }
							authLogger.Errorf("Failed to retrieve credentials for instance %s: %v", instance, result.Error)
	      return nil, DatabaseError{message: "failed to retrieve credentials"}
	  }
	  authLogger.Printf("INFO: Successfully retrieved credentials for instance %s", instance)
	  return &credential, nil
}

type Config struct {
	Instance string
	Username string
	Password string
	Table    string
	Limit    int
}

type ServiceNowRecord map[string]interface{}

type ServiceNowResponse struct {
	Result []ServiceNowRecord `json:"result"`
}

func getBasicAuthHeader(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func fetchRecords(config Config) ([]ServiceNowRecord, error) {
	url := fmt.Sprintf("https://%s/api/now/table/%s?sysparm_limit=%d", config.Instance, config.Table, config.Limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, NetworkError{message: fmt.Sprintf("failed to create request: %v", err)}
	}

	req.Header.Add("Authorization", getBasicAuthHeader(config.Username, config.Password))
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, NetworkError{message: fmt.Sprintf("failed to connect: %v", err)}
	}
	defer resp.Body.Close()

	  if resp.StatusCode == http.StatusUnauthorized {
							authLogger.Errorf("Authentication failed for instance %s with user %s", config.Instance, config.Username)
	      return nil, AuthError{message: "invalid credentials"}
	  } else if resp.StatusCode != http.StatusOK {
		return nil, NetworkError{message: fmt.Sprintf("server returned status code %d", resp.StatusCode)}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var response ServiceNowResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
		  
		  authLogger.Printf("INFO: Successfully authenticated for instance %s with user %s", config.Instance, config.Username)
		  return response.Result, nil
}

func createTable(records []ServiceNowRecord) table.Model {
	var columns []table.Column
	var rows []table.Row

	if len(records) > 0 {
		// Create columns from first record
		for key := range records[0] {
			columns = append(columns, table.Column{Title: key, Width: 20})
		}

		// Create rows
		for _, record := range records {
			var row []string
			for _, col := range columns {
				value := record[col.Title]
				row = append(row, fmt.Sprintf("%v", value))
			}
			rows = append(rows, row)
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	t.SetStyles(s)

	return t
}

var (
	recordLimit int

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(1, 0)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

type inputState int

const (
	instanceInput inputState = iota
	usernameInput
	passwordInput
	tableInput
	fetching
	display
)

type model struct {
	inputs  []textinput.Model
	config  Config
	table   table.Model
	err     error
	records []ServiceNowRecord
	state   inputState
	spinner spinner.Model
	loading bool
}

func initialModel() model {
	var inputs []textinput.Model
	labels := []string{"Instance", "Username", "Password", "Table"}

	for i := 0; i < 4; i++ {
		input := textinput.New()
		input.Width = 40
		input.Prompt = "> "
		input.PromptStyle = focusedStyle
		input.PlaceholderStyle = blurredStyle
		input.Placeholder = labels[i]

		if i == 2 { // Password field
			input.EchoMode = textinput.EchoPassword
			input.EchoCharacter = '•'
		}

		inputs = append(inputs, input)
	}

	inputs[0].Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = focusedStyle

	return model{
		inputs:  inputs,
		state:   instanceInput,
		spinner: s,
		loading: false,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == display {
				return m, tea.Quit
			}
		case "enter":
			if m.state < fetching {
				m.state++
				if m.state < fetching {
					m.inputs[m.state].Focus()
					return m, textinput.Blink
				}

				// Process form
				m.loading = true
				m.config = Config{
					Instance: m.inputs[0].Value(),
					Username: m.inputs[1].Value(),
					Password: m.inputs[2].Value(),
					Table:    m.inputs[3].Value(),
					Limit:    recordLimit,
				}
				return m, tea.Batch(
					m.spinner.Tick,
					func() tea.Msg {
						// Save credentials first
						if err := saveCredentials(m.config.Instance, m.config.Username, m.config.Password); err != nil {
						    m.err = err
						    m.state = display
						    m.loading = false
						    return nil
						}

						records, err := fetchRecords(m.config)
						if err != nil {
						    m.err = err
						} else {
						    m.records = records
						    if len(records) > 0 {
						        m.table = createTable(records)
						    }
						}
						m.state = display
						m.loading = false
						return nil
					},
				)
			}
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.state < fetching {
		var cmd tea.Cmd
		m.inputs[m.state], cmd = m.inputs[m.state].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.state < fetching {
		var b strings.Builder

		b.WriteString(titleStyle.Render("ServiceNow Record Fetcher"))
		b.WriteString("\n\n")

		for i := 0; i < len(m.inputs); i++ {
			if inputState(i) == m.state {
				b.WriteString(focusedStyle.Render(m.inputs[i].Placeholder + ": "))
			} else {
				b.WriteString(blurredStyle.Render(m.inputs[i].Placeholder + ": "))
			}
			b.WriteString(m.inputs[i].View())
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(blurredStyle.Render("Press Enter to continue • ESC to quit"))

		return b.String()
	}

	if m.loading {
		return fmt.Sprintf("\n%s Loading records...\n", m.spinner.View())
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("\nError: %v\nPress q to quit", m.err))
	}

	if len(m.records) == 0 {
		return blurredStyle.Render("\nNo records found\nPress q to quit")
	}

	return m.table.View() + "\n" + blurredStyle.Render("Press q to quit")
}

func main() {
    // Setup authentication logger
				logFile, err := os.OpenFile("auth.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
				    log.Fatalf("Failed to open auth.log: %v", err)
				}
				defer logFile.Close()
				fileLogger := log.New(logFile, "", log.LstdFlags)
				authLogger = newPrettyLogger(fileLogger)

    flag.IntVar(&recordLimit, "limit", 5, "number of records to fetch")
    flag.Parse()

    if err := initDB(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
        os.Exit(1)
    }
}
