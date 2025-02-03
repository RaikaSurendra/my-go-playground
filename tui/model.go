package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	listView  uint = iota
	titleView      = 1
	bodyView       = 2
)

type model struct {
	state uint
}

func NewModel() model {
	return model{
		state: listView,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
