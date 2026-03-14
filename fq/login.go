package fq

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/shaco-go/tomato-terminal/config"
	"github.com/shaco-go/tomato-terminal/pkg"
	"go.uber.org/zap"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var (
	instance *Login
	once     sync.Once
)

const fqURI = "https://fanqienovel.com"

var reInitialState = regexp.MustCompile(`<script>window\.__INITIAL_STATE__=(\{.+?\})\(function`)

type chapterInfo struct {
	itemID     string
	preItemID  string
	nextItemID string
	title      string
}

type Login struct {
	mu          sync.RWMutex
	b           *rod.Browser
	content     []string
	nextContent []string
	prevContent []string
	currentInfo chapterInfo
	nextInfo    chapterInfo
	prevInfo    chapterInfo
	loadedURL   string
}

func GetInstance() *Login {
	once.Do(func() {
		instance = &Login{}
	})
	return instance
}

func (l *Login) GetContent() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.content
}

func closePage(page *rod.Page) {
	if page == nil {
		return
	}
	_ = page.Close()
}

func (l *Login) waitLogin(attempts int, interval time.Duration) (bool, error) {
	page := l.b.MustPage(fqURI)
	page.MustWaitStable()
	defer page.MustClose()
	for i := 0; i < attempts; i++ {
		hasLoginButton, _, err := page.HasR(".slogin-user-avatar__buttons__item", "登录")
		if err != nil {
			return false, err
		}
		if !hasLoginButton {
			return true, nil
		}
		time.Sleep(interval)
		page.MustReload()
		page.MustWaitStable()
	}
	return false, nil
}

func (l *Login) StartBrowser() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.b != nil {
		return nil
	}

	path, has := launcher.LookPath()
	launch := launcher.New().Headless(true)
	if has {
		launch.Bin(path)
	}

	controlURL, err := launch.Launch()
	if err != nil {
		return err
	}

	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return err
	}

	l.b = browser
	return nil
}

func (l *Login) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.b == nil {
		return nil
	}

	if err := l.b.Close(); err != nil {
		return err
	}

	l.b = nil
	return nil
}

func (l *Login) IsLogin() (bool, error) {
	page, err := l.openPage(fqURI)
	if err != nil {
		return false, err
	}
	defer closePage(page)

	hasLoginButton, _, err := page.HasR(".slogin-user-avatar__buttons__item", "登录")
	if err != nil {
		return false, err
	}

	return !hasLoginButton, nil
}

func (l *Login) Login() (bool, error) {
	page, err := l.openPage(fqURI)
	if err != nil {
		return false, err
	}
	defer closePage(page)

	loginButton, err := page.ElementR(".slogin-user-avatar__buttons__item", "登录")
	if err != nil {
		return false, err
	}

	if err := loginButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return false, err
	}
	if err := page.WaitStable(time.Second); err != nil {
		return false, err
	}

	qrTab, err := page.ElementR(".slogin-pc-form-header__title__tab.slogin-pc-form-header__title__tab--dim", "扫码登录")
	if err != nil {
		return false, err
	}
	if err := qrTab.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return false, err
	}
	if err := page.WaitStable(time.Second); err != nil {
		return false, err
	}

	qrImage, err := page.Element("img.slogin-qrcode-scan-page__content__code__img")
	if err != nil {
		return false, err
	}

	resource, err := qrImage.Resource()
	if err != nil {
		return false, err
	}

	filename := pkg.WorkspaceDir("login-qrcode.png")
	if err := os.WriteFile(filename, resource, 0o644); err != nil {
		return false, err
	}

	_ = exec.Command("cmd", "/c", "start", filename).Start()

	ok, err := l.waitLogin(15, time.Second)
	zap.L().Debug("是否登录成功", zap.Error(err), zap.Bool("ok", ok))
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	rawCookies, err := page.Cookies([]string{fqURI})
	if err != nil {
		return false, err
	}

	cookies := make([]*http.Cookie, 0, len(rawCookies))
	for _, item := range rawCookies {
		c := &http.Cookie{
			Name:     item.Name,
			Value:    item.Value,
			Domain:   item.Domain,
			Path:     item.Path,
			Secure:   item.Secure,
			HttpOnly: item.HTTPOnly,
		}
		if item.Expires > 0 {
			c.Expires = time.Unix(int64(item.Expires), 0)
		}
		cookies = append(cookies, c)
	}

	config.Conf.Cookie = cookies

	return true, nil
}

func (l *Login) openPage(url string) (*rod.Page, error) {
	if err := l.StartBrowser(); err != nil {
		return nil, err
	}

	l.mu.RLock()
	browser := l.b
	l.mu.RUnlock()
	if browser == nil {
		return nil, errors.New("browser is not initialized")
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return nil, err
	}

	if err := page.WaitStable(time.Second); err != nil {
		closePage(page)
		return nil, err
	}

	return page, nil
}
