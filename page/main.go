package page

import (
	"github.com/shaco-go/tomato-terminal/types"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"go.uber.org/zap"
)

func New() Model {
	m := Model{}
	m.pageManage = make(map[types.Page]tea.Model)
	m.pageManage[types.MenuPage] = NewMenu()
	m.pageManage[types.ReaderPage] = NewRead()
	m.pageManage[types.ConfigPage] = NewConfig()
	m.pageManage[types.LoginPage] = NewLogin()
	m.pageIndex = types.MenuPage
	return m
}

type Model struct {
	configForm *huh.Form
	pageIndex  types.Page
	quit       bool
	pageManage map[types.Page]tea.Model
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, page := range m.pageManage {
		if page.Init() != nil {
			cmds = append(cmds, page.Init())
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if msg.String() == "esc" && m.pageIndex != types.ReaderPage {
				m.pageIndex = types.ReaderPage
				return m, nil
			}
			m.quit = true
			return m, tea.Quit
		}
	case types.ChangePageMsg:
		zap.L().Debug("切换页面", zap.Any("msg", msg))
		m.pageIndex = types.Page(msg)
		return m, m.GetPage().Init()
	}
	var cmd tea.Cmd
	m.pageManage[m.pageIndex], cmd = m.GetPage().Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	if m.quit {
		return tea.NewView("")
	}
	return m.GetPage().View()
}

func (m Model) GetPage() tea.Model {
	val, ok := m.pageManage[m.pageIndex]
	if !ok {
		return m.pageManage[types.ReaderPage]
	}
	return val
}
