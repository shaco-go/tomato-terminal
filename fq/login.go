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

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"go.uber.org/zap"
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

func (f *Login) GetContent() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.content
}

func closePage(page *rod.Page) {
	if page == nil {
		return
	}
	if err := page.Close(); err != nil {
		zap.L().Warn("close page failed", zap.Error(err))
	}
}

func waitLogin(page *rod.Page, attempts int, interval time.Duration) (bool, error) {
	for i := 0; i < attempts; i++ {
		hasLoginButton, _, err := page.HasR(".slogin-user-avatar__buttons__item", "登录")
		if err != nil {
			return false, err
		}
		if !hasLoginButton {
			return true, nil
		}
		time.Sleep(interval)
	}
	return false, nil
}

func (f *Login) StartBrowser() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.b != nil {
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

	f.b = browser
	return nil
}

func (f *Login) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.b == nil {
		return nil
	}

	if err := f.b.Close(); err != nil {
		return err
	}

	f.b = nil
	return nil
}

func (f *Login) IsLogin() (bool, error) {
	page, err := f.openPage(fqURI)
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

func (f *Login) Login() (bool, error) {
	page, err := f.openPage(fqURI)
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

	if err := exec.Command("cmd", "/c", "start", filename).Start(); err != nil {
		zap.L().Warn("open qrcode file failed", zap.Error(err), zap.String("file", filename))
	}

	ok, err := waitLogin(page, 10, time.Second)
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

func (f *Login) openPage(url string) (*rod.Page, error) {
	if err := f.StartBrowser(); err != nil {
		return nil, err
	}

	f.mu.RLock()
	browser := f.b
	f.mu.RUnlock()
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
