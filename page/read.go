package page

import (
	"strings"

	"github.com/shaco-go/tomato-terminal/fq"
	"github.com/shaco-go/tomato-terminal/types"

	tea "charm.land/bubbletea/v2"
)

func NewRead() ReadModel {
	return ReadModel{
		hide:   false,
		login:  fq.GetInstance(),
		reader: fq.NewReader(),
	}
}

type ReadModel struct {
	hide    bool
	content []string
	login   *fq.Login
	reader  *fq.Reader
}

func (r ReadModel) Init() tea.Cmd {
	return nil
}

func (r ReadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if r.hide {
				return r, nil
			}
			return r, func() tea.Msg {
				return types.ChangePageMsg(types.MenuPage)
			}
		default:
			return r.Hide(), nil
		}
	case tea.MouseClickMsg:
		switch msg.Button {
		case tea.MouseLeft, tea.MouseRight:
			return r.Hide(), nil
		case tea.MouseMiddle:
			return r.Visible(), nil
		}
	case tea.MouseWheelMsg:
		if !r.hide {
			switch msg.Button {
			case tea.MouseWheelDown:
				r.content = r.reader.Next()
			case tea.MouseWheelUp:
				r.content = r.reader.Prev()
			}
		}
	}

	return r, nil
}

func (r ReadModel) View() tea.View {
	var v tea.View
	if r.hide {
		v = tea.NewView("")
	} else {
		v = tea.NewView(strings.Join(r.content, "\n"))
	}
	v.MouseMode = tea.MouseModeAllMotion
	return v
}

func (r ReadModel) Hide() ReadModel {
	r.hide = true
	return r
}
func (r ReadModel) Visible() ReadModel {
	r.hide = false
	return r
}
