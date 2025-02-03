	package tui

	import (
	        "strings"

	        tea "github.com/charmbracelet/bubbletea"
	        "github.com/charmbracelet/lipgloss"
	)

	var (
	        appWidth = 60
	        
	        subtleBlue   = lipgloss.Color("69")
	        subtleGray   = lipgloss.Color("241")
	        primaryBlue  = lipgloss.Color("39")
	        
	        focusedStyle = lipgloss.NewStyle().
	                Background(lipgloss.Color("236")).
	                Foreground(lipgloss.Color("252")).
	                Padding(1, 2).
	                MarginLeft(2).
	                Width(appWidth - 4)
	        
	        blurredStyle = lipgloss.NewStyle().
	                Foreground(lipgloss.Color("242")).
	                Padding(1, 2).
	                MarginLeft(2).
	                Width(appWidth - 4)
	        
	        focusIndicator = lipgloss.NewStyle().
	                Foreground(primaryBlue).
	                SetString("▎")
	        
	        titleStyle = lipgloss.NewStyle().
	                Foreground(subtleBlue).
	                Bold(true).
	                Padding(1, 2).
	                MarginTop(1).
	                MarginBottom(2).
	                Width(appWidth).
	                Align(lipgloss.Center)
	        
	        labelStyle = lipgloss.NewStyle().
	                Foreground(subtleGray).
	                PaddingLeft(2).
	                MarginBottom(1)
	        
	        helpStyle = lipgloss.NewStyle().
	                Foreground(subtleGray).
	                Width(appWidth).
	                Align(lipgloss.Center).
	                MarginTop(2)
	        
	        footerStyle = lipgloss.NewStyle().
	                Foreground(subtleGray).
	                Width(appWidth).
	                Align(lipgloss.Center).
	                PaddingTop(2).
	                BorderTop(true).
	                BorderStyle(lipgloss.Border{
	                        Top: "─",
	                })
	        
	        docStyle = lipgloss.NewStyle().
	                Padding(2).
	                Align(lipgloss.Center)
	)

	type LoginModel struct {
	        InstanceURL     string
	        Username        string
	        Password        string
	        focused        int
	        cursorPosition map[int]int
	        err           error
	}

	func InitialModel() LoginModel {
	        return LoginModel{
	                focused:        0,
	                cursorPosition: make(map[int]int),
	        }
	}

	func CreateProgram(m LoginModel) *tea.Program {
	  return tea.NewProgram(m)
	}

	func (m LoginModel) Init() tea.Cmd {
	  return nil
	}

	func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	        switch msg := msg.(type) {
	        case tea.KeyMsg:
	                switch msg.Type {
	                case tea.KeyCtrlC, tea.KeyEsc:
	                        return m, tea.Quit

	                case tea.KeyTab:
	                        m.focused = (m.focused + 1) % 3
	                        return m, nil

	                case tea.KeyShiftTab:
	                        m.focused = (m.focused - 1)
	                        if m.focused < 0 {
	                                m.focused = 2
	                        }
	                        return m, nil

	                case tea.KeyEnter:
	                        if m.focused == 2 && m.validate() {
	                                return m, tea.Quit
	                        }
	                        m.focused = (m.focused + 1) % 3
	                        return m, nil

	                case tea.KeyLeft, tea.KeyRight:
	                        switch m.focused {
	                        case 0:
	                                m.cursorPosition[0], _ = m.updateCursor(m.InstanceURL, msg)
	                        case 1:
	                                m.cursorPosition[1], _ = m.updateCursor(m.Username, msg)
	                        case 2:
	                                m.cursorPosition[2], _ = m.updateCursor(m.Password, msg)
	                        }
	                        return m, nil
	                }

	                switch m.focused {
	                case 0:
	                        m.InstanceURL, _ = m.updateInput(m.InstanceURL, msg)
	                        if pos, exists := m.cursorPosition[0]; !exists || pos > len(m.InstanceURL) {
	                                m.cursorPosition[0] = len(m.InstanceURL)
	                        }
	                case 1:
	                        m.Username, _ = m.updateInput(m.Username, msg)
	                        if pos, exists := m.cursorPosition[1]; !exists || pos > len(m.Username) {
	                                m.cursorPosition[1] = len(m.Username)
	                        }
	                case 2:
	                        m.Password, _ = m.updateInput(m.Password, msg)
	                        if pos, exists := m.cursorPosition[2]; !exists || pos > len(m.Password) {
	                                m.cursorPosition[2] = len(m.Password)
	                        }
	                }
	        }

	        return m, nil
	}

	func (m *LoginModel) updateInput(current string, msg tea.KeyMsg) (string, tea.Cmd) {
	        pos := m.cursorPosition[m.focused]
	        switch msg.Type {
	        case tea.KeyBackspace:
	                if pos > 0 {
	                        current = current[:pos-1] + current[pos:]
	                        m.cursorPosition[m.focused]--
	                }
	        case tea.KeyDelete:
	                if pos < len(current) {
	                        current = current[:pos] + current[pos+1:]
	                }
	        case tea.KeyRunes:
	                if pos == len(current) {
	                        current += string(msg.Runes)
	                } else {
	                        current = current[:pos] + string(msg.Runes) + current[pos:]
	                }
	                m.cursorPosition[m.focused]++
	        }
	        return current, nil
	}

	func (m *LoginModel) updateCursor(current string, msg tea.KeyMsg) (int, tea.Cmd) {
	        pos := m.cursorPosition[m.focused]
	        switch msg.Type {
	        case tea.KeyLeft:
	                if pos > 0 {
	                        pos--
	                }
	        case tea.KeyRight:
	                if pos < len(current) {
	                        pos++
	                }
	        }
	        return pos, nil
	}

	func (m LoginModel) validate() bool {
	  return m.InstanceURL != "" && m.Username != "" && m.Password != ""
	}

	func (m LoginModel) View() string {
	        var s strings.Builder

									title := titleStyle.Render("ServiceNow CLI Login")
									s.WriteString(title + "\n")

									s.WriteString(strings.Repeat("─", appWidth) + "\n\n")

									// Instance URL field
									instanceStyle := blurredStyle
									if m.focused == 0 {
									        s.WriteString(focusIndicator.Render())
									        instanceStyle = focusedStyle
									} else {
									        s.WriteString(" ")
									}
	        instanceInput := m.InstanceURL
	        if m.focused == 0 {
	                pos := m.cursorPosition[0]
	                if pos < len(instanceInput) {
	                        instanceInput = instanceInput[:pos] + "│" + instanceInput[pos:]
	                } else {
	                        instanceInput = instanceInput + "│"
	                }
	        }
									s.WriteString(labelStyle.Render("Instance Name (e.g., dev12345):"))
									s.WriteString("\n" + instanceStyle.Render(instanceInput) + "\n\n")

									// Username field
									usernameStyle := blurredStyle
									if m.focused == 1 {
									        s.WriteString(focusIndicator.Render())
									        usernameStyle = focusedStyle
									} else {
									        s.WriteString(" ")
									}
	        usernameInput := m.Username
	        if m.focused == 1 {
	                pos := m.cursorPosition[1]
	                if pos < len(usernameInput) {
	                        usernameInput = usernameInput[:pos] + "│" + usernameInput[pos:]
	                } else {
	                        usernameInput = usernameInput + "│"
	                }
	        }
	        s.WriteString(labelStyle.Render("Username:"))
	        s.WriteString("\n" + usernameStyle.Render(usernameInput) + "\n\n")

	        // Password field
									passwordStyle := blurredStyle
									if m.focused == 2 {
									        s.WriteString(focusIndicator.Render())
									        passwordStyle = focusedStyle
									} else {
									        s.WriteString(" ")
									}
	        passwordInput := strings.Repeat("•", len(m.Password))
	        if m.focused == 2 {
	                pos := m.cursorPosition[2]
	                if pos < len(passwordInput) {
	                        passwordInput = passwordInput[:pos] + "│" + passwordInput[pos:]
	                } else {
	                        passwordInput = passwordInput + "│"
	                }
	        }
	        s.WriteString(labelStyle.Render("Password:"))
	        s.WriteString("\n" + passwordStyle.Render(passwordInput) + "\n\n")

									s.WriteString("\n" + strings.Repeat("─", appWidth) + "\n")
									help := helpStyle.Render("Tab: Navigate • Enter: Submit • Ctrl+c: Exit")
									s.WriteString(help + "\n")

									footer := footerStyle.Render("v1.0.0 • ServiceNow CLI")
									s.WriteString(footer)

	        return docStyle.Render(s.String())
	}
