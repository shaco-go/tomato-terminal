package types

const ProjectName = "tomato"

type Page int8

const (
	MenuPage Page = iota
	ConfigPage
	ReaderPage
	LoginPage
)

func (p Page) String() string {
	switch p {
	case MenuPage:
		return "menu"
	case ConfigPage:
		return "config"
	case ReaderPage:
		return "reader"
	case LoginPage:
		return "login"
	default:
		return "unknown"
	}
}

type ChangePageMsg Page

type QuitMsg struct{}
