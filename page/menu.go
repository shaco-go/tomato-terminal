package page

import (
	"github.com/shaco-go/tomato-terminal/types"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func NewMenu() MenuModel {
	m := MenuModel{}
	m.form = m.newForm()
	return m
}

type MenuModel struct {
	form *huh.Form
}

func (m MenuModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	form, cmd := m.form.Update(msg)
	if val, ok := form.(*huh.Form); ok {
		m.form = val
		cmds = append(cmds, cmd)
	}
	if m.form.State == huh.StateCompleted {
		val := m.form.Get("index").(types.Page)
		m.form = m.newForm()
		cmds = append(cmds, func() tea.Msg {
			return types.ChangePageMsg(val)
		})
	}
	return m, tea.Batch(cmds...)
}

func (m MenuModel) View() tea.View {
	return tea.NewView(m.form.View())
}

func (m MenuModel) newForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(huh.NewSelect[types.Page]().
			Key("index").
			Options(
				huh.NewOption("login", types.LoginPage),
				huh.NewOption("config", types.ConfigPage),
			),
		),
	)
}
