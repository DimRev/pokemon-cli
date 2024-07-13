package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SidebarModel struct {
	Routes         []string
	SelectedRouted int
	IsFocused      bool
	Styles         *SidebarStyle
}

type SidebarStyle struct {
	FocusedSelectedStyle     lipgloss.Style
	FocusedUnselectedStyle   lipgloss.Style
	UnfocusedSelectedStyle   lipgloss.Style
	UnfocusedUnselectedStyle lipgloss.Style
}

func (s SidebarModel) Init() tea.Cmd {
	return nil
}

func (s SidebarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			s.SelectedRouted++
			if s.SelectedRouted >= len(s.Routes) {
				s.SelectedRouted = 0
			}
		case "up":
			s.SelectedRouted--
			if s.SelectedRouted < 0 {
				s.SelectedRouted = len(s.Routes) - 1
			}
		}
	}
	return s, nil
}

func (s SidebarModel) View() string {
	renderRoutes := []string{}
	for i, route := range s.Routes {
		if i == s.SelectedRouted {
			if s.IsFocused {
				renderRoutes = append(renderRoutes, s.Styles.FocusedSelectedStyle.Render(">"+route))
			} else {
				renderRoutes = append(renderRoutes, s.Styles.UnfocusedSelectedStyle.Render(">"+route))
			}
		} else {
			if s.IsFocused {
				renderRoutes = append(renderRoutes, s.Styles.FocusedUnselectedStyle.Render(" "+route))
			} else {
				renderRoutes = append(renderRoutes, s.Styles.UnfocusedUnselectedStyle.Render(" "+route))
			}
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderRoutes...,
	)
}
