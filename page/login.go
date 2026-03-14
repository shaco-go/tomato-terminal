package page

import (
	"strings"
	"time"

	"github.com/shaco-go/tomato-terminal/config"
	"github.com/shaco-go/tomato-terminal/fq"
	"github.com/shaco-go/tomato-terminal/types"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"go.uber.org/zap"
)

const (
	StatusLoggedOut = iota
	StatusLoggedIn
	StatusLoggingIn
	StatusLoginFailed
)

type tickMsg time.Time

type loginResultMsg struct {
	success bool
	err     error
}

func NewLogin() LoginPage {
	return LoginPage{
		selectedOption: 1, // 默认选中关闭
		status:         StatusLoggedOut,
		fq:             &fq.Login{},
	}
}

type LoginPage struct {
	status         int // 0=未登录, 1=已登录, 2=登录中, 3=登录失败
	selectedOption int // 0=登录/登出, 1=关闭
	animationFrame int
	fq             *fq.Login
}

func (l LoginPage) Init() tea.Cmd {
	var cmds = []tea.Cmd{l.tick()}
	if len(config.Conf.Cookie) > 0 {
		l.status = StatusLoggedIn
		cmds = append(cmds, func() tea.Msg {
			return loginResultMsg{
				success: true,
			}
		})
	}
	return tea.Batch(cmds...)
}

func (l LoginPage) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (l LoginPage) doLogin() tea.Cmd {
	return func() tea.Msg {
		err := l.fq.StartBrowser()
		if err != nil {
			zap.L().Error("failed to start browser", zap.Error(err))
			return loginResultMsg{success: false, err: err}
		}
		success, err := l.fq.Login()
		l.Close()
		return loginResultMsg{success: success, err: err}
	}
}

func (l LoginPage) isLoginEnabled() bool {
	return l.status == StatusLoggedOut || l.status == StatusLoggedIn || l.status == StatusLoginFailed
}

func (l LoginPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if l.status == StatusLoggingIn {
			l.animationFrame = (l.animationFrame + 1) % 4
		}
		return l, l.tick()
	case loginResultMsg:
		if msg.err != nil {
			zap.L().Error("login failed", zap.Error(msg.err))
			l.status = StatusLoginFailed
		} else if msg.success {
			config.Conf.Flush()
			l.status = StatusLoggedIn
		} else {
			l.status = StatusLoginFailed
		}
		return l, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "right", "tab":
			if l.isLoginEnabled() {
				l.selectedOption = 1 - l.selectedOption
			}
		case "enter":
			zap.L().Debug("login enter", zap.Any("selectedOption", l.selectedOption))
			if l.selectedOption == 0 && l.isLoginEnabled() {
				if l.status == StatusLoggedIn {
					config.Conf.Cookie = nil
					config.Conf.Flush()
					l.status = StatusLoggedOut
				} else {
					l.status = StatusLoggingIn
					return l, l.doLogin()
				}
			} else if l.selectedOption == 1 {
				return l, func() tea.Msg {
					return l.ChangePage(types.ReaderPage)
				}
			}
		}
	}
	return l, nil
}

func (l LoginPage) getStatusText() string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸"}
	switch l.status {
	case StatusLoggedOut:
		return "○ Logged Out"
	case StatusLoggedIn:
		return "✓ Logged In"
	case StatusLoggingIn:
		return spinners[l.animationFrame] + " Logging In"
	case StatusLoginFailed:
		return "✗ Login Failed"
	default:
		return "○ Logged Out"
	}
}

func (l LoginPage) getActionButton() string {
	if l.status == StatusLoggedIn {
		return "Logout"
	}
	return "Login"
}

func (l LoginPage) View() tea.View {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Background(lipgloss.Color("235"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	disabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("238")).
		Strikethrough(true)

	var content strings.Builder

	// Status
	content.WriteString(titleStyle.Render(l.getStatusText()))
	content.WriteString("\n\n")

	// Buttons
	actionBtn := l.getActionButton()
	loginEnabled := l.isLoginEnabled()

	if !loginEnabled {
		content.WriteString(disabledStyle.Render("  " + actionBtn))
	} else if l.selectedOption == 0 {
		content.WriteString(selectedStyle.Render("► " + actionBtn))
	} else {
		content.WriteString(normalStyle.Render("  " + actionBtn))
	}

	content.WriteString("   ")

	if l.selectedOption == 1 {
		content.WriteString(selectedStyle.Render("► Close"))
	} else {
		content.WriteString(normalStyle.Render("  Close"))
	}

	content.WriteString("\n\n")
	content.WriteString(normalStyle.Render("←/→/Tab Switch | Enter Confirm"))

	return tea.NewView(content.String())
}

func (l LoginPage) Close() {
	err := l.fq.Close()
	if err != nil {
		zap.L().Error("failed to close browser", zap.Error(err))
	}
}

func (l LoginPage) ChangePage(p types.Page) tea.Msg {
	return types.ChangePageMsg(p)
}
