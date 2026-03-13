package page

import (
	"strconv"
	"strings"
	"github.com/shaco-go/tomato-terminal/config"
	"github.com/shaco-go/tomato-terminal/types"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func NewConfig() ConfigPage {
	c := ConfigPage{}
	c.form = c.newForm()
	return c
}

type ConfigPage struct {
	form *huh.Form
}

const defaultLine = 10

func parseLineOrDefault(raw string, current int) int {
	line, err := strconv.Atoi(strings.TrimSpace(raw))
	if err == nil && line > 0 {
		return line
	}
	if current > 0 {
		return current
	}
	return defaultLine
}

func applyReadConfig(itemID string, rawLine string) {
	trimmedItemID := strings.TrimSpace(itemID)
	if config.Conf.ItemID != trimmedItemID {
		config.Conf.Cursor = 0
	}
	config.Conf.ItemID = trimmedItemID
	config.Conf.Line = parseLineOrDefault(rawLine, config.Conf.Line)
}

func (c ConfigPage) Init() tea.Cmd {
	return c.form.Init()
}

func (c ConfigPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	form, cmd := c.form.Update(msg)
	if val, ok := form.(*huh.Form); ok {
		c.form = val
		cmds = append(cmds, cmd)
	}
	if c.form.State == huh.StateCompleted {
		applyReadConfig(c.form.GetString("itemId"), c.form.GetString("line"))
		c.form = c.newForm()
		cmds = append(cmds, func() tea.Msg {
			return types.ChangePageMsg(types.ReaderPage)
		})
	}
	return c, tea.Batch(cmds...)
}

func (c ConfigPage) View() tea.View {
	return tea.NewView(c.form.View())
}

func (c ConfigPage) newForm() *huh.Form {
	var itemID = config.Conf.ItemID
	var line = strconv.Itoa(config.Conf.Line)
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("itemId").Title("ItemID").Value(&itemID),
			huh.NewInput().Key("line").Title("Line").Value(&line),
		),
	)
}
