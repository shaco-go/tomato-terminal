package fq

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/shaco-go/tomato-terminal/config"
	"github.com/shaco-go/tomato-terminal/types"

	"github.com/PuerkitoBio/goquery"
	"github.com/dop251/goja"
	"github.com/tidwall/gjson"

	"resty.dev/v3"
)

type chapterItem struct {
	Content    []string
	ItemID     string
	PrevItemID string
	NextItemID string
	Title      string
}

type moveDirection int

const (
	moveNone moveDirection = iota
	movePrev
	moveNext
)

func NewReader() *Reader {
	return &Reader{
		chapter: &chapterItem{},
	}
}

type Reader struct {
	chapter  *chapterItem
	lastMove moveDirection
}

func normalizeLine(line int) int {
	if line <= 0 {
		return 1
	}
	return line
}

func (r *Reader) Prev() []string {
	err := r.GetChapter()
	if err != nil {
		return []string{err.Error()}
	}
	line := normalizeLine(config.Conf.Line)

	// Cursor 在 Next 后会指向当前页末尾后一位，方向切换到 Prev 时需要额外回退一页。
	if r.lastMove == moveNext {
		config.Conf.Cursor -= line
	}
	if config.Conf.Cursor < 0 {
		config.Conf.Cursor = 0
	}

	// 检查是否需要翻到上一章
	if config.Conf.Cursor-line < 0 {
		if r.chapter.PrevItemID == "" {
			r.lastMove = moveNone
			return []string{"没有上一章节"}
		}
		config.Conf.ItemID = r.chapter.PrevItemID
		err := r.GetChapter()
		if err != nil {
			return []string{err.Error()}
		}
		// 定位到上一章的末尾
		config.Conf.Cursor = len(r.chapter.Content)
	}

	// 计算开始位置
	startPos := config.Conf.Cursor - line
	if startPos < 0 {
		startPos = 0
	}

	content := r.chapter.Content[startPos:config.Conf.Cursor]
	config.Conf.Cursor = startPos
	r.lastMove = movePrev

	return content
}

func (r *Reader) Current() []string {
	err := r.GetChapter()
	if err != nil {
		return []string{err.Error()}
	}
	line := normalizeLine(config.Conf.Line)
	if config.Conf.Cursor >= len(r.chapter.Content) {
		config.Conf.Cursor = len(r.chapter.Content) - line
		if config.Conf.Cursor < 0 {
			config.Conf.Cursor = 0
		}
	}
	endPos := config.Conf.Cursor + line
	if endPos > len(r.chapter.Content) {
		endPos = len(r.chapter.Content)
	}
	return r.chapter.Content[config.Conf.Cursor:endPos]
}

func (r *Reader) Next() []string {
	err := r.GetChapter()
	if err != nil {
		return []string{err.Error()}
	}
	line := normalizeLine(config.Conf.Line)

	// Cursor 在 Prev 后会落在当前页起点，方向切换到 Next 时先前进一页。
	if r.lastMove == movePrev {
		config.Conf.Cursor += line
	}
	if config.Conf.Cursor > len(r.chapter.Content) {
		config.Conf.Cursor = len(r.chapter.Content)
	}
	if config.Conf.Cursor >= len(r.chapter.Content) {
		if r.chapter.NextItemID == "" {
			r.lastMove = moveNone
			return []string{"已到最新章节"}
		}
		// 需要翻页了
		config.Conf.Cursor = 0
		config.Conf.ItemID = r.chapter.NextItemID
		err := r.GetChapter()
		if err != nil {
			return []string{err.Error()}
		}
	}
	if config.Conf.Cursor+line > len(r.chapter.Content) {
		// 说明超出了
		content := r.chapter.Content[config.Conf.Cursor:]
		config.Conf.Cursor = len(r.chapter.Content)
		r.lastMove = moveNext
		return content
	}
	content := r.chapter.Content[config.Conf.Cursor : config.Conf.Cursor+line]
	config.Conf.Cursor += line
	r.lastMove = moveNext
	return content
}

func (r *Reader) GetChapter() error {
	if config.Conf.ItemID == "" {
		return errors.New("ItemID is empty")
	}
	if r.chapter.ItemID == "" || r.chapter.ItemID != config.Conf.ItemID {
		chapter, err := r.getChapterByItemID(config.Conf.ItemID)
		if err != nil {
			return err
		}
		r.chapter = chapter
		r.lastMove = moveNone
	}
	return nil
}

func (r *Reader) getChapterByItemID(itemID string) (*chapterItem, error) {
	url := fmt.Sprintf("https://fanqienovel.com/reader/%s", itemID)
	if len(config.Conf.Cookie) == 0 {
		return nil, errors.New("please log in")
	}

	client := resty.New()
	req := client.R().SetCookies(config.Conf.Cookie)
	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= http.StatusBadRequest {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode())
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	var js string
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "__INITIAL_STATE__") {
			js = s.Text()
		}
	})
	if strings.TrimSpace(js) == "" {
		return nil, errors.New("获取章节失败: 未找到 __INITIAL_STATE__ 章节脚本")
	}
	dom, err := r.parseContent(js)
	if err != nil {
		return nil, err
	}
	chapter := &chapterItem{
		ItemID:     decodeText(dom.Get("itemId").String()),
		PrevItemID: dom.Get("preItemId").String(),
		NextItemID: dom.Get("nextItemId").String(),
		Title:      dom.Get("title").String(),
	}
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(dom.Get("content").String()))
	if err != nil {
		return nil, err
	}
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		temp := decodeText(s.Text())
		if strings.TrimSpace(temp) != "" {
			chapter.Content = append(chapter.Content, temp)
		}
	})
	if len(chapter.Content) == 0 {
		return nil, errors.New("获取章节失败:" + url)
	}
	chapter.Content = append([]string{chapter.Title}, chapter.Content...)
	return chapter, nil
}

func (r *Reader) parseContent(script string) (*gjson.Result, error) {
	if strings.TrimSpace(script) == "" {
		return nil, errors.New("未找到 __INITIAL_STATE__ 章节脚本")
	}
	vm := goja.New()

	// 创建 window 对象
	err := vm.Set("window", vm.NewObject())
	if err != nil {
		return nil, err
	}

	// 执行脚本
	_, err = vm.RunString(script + "\n const parseData = JSON.stringify(window.__INITIAL_STATE__.reader.chapterData)")
	if err != nil {
		return nil, err
	}

	// 获取 __INITIAL_STATE__
	p := gjson.Parse(vm.Get("parseData").String())

	return &p, nil
}

func decodeText(text string) string {
	var result strings.Builder
	for _, r := range text {
		code := int(r)
		if code > 50000 && code < 60000 {
			key := strconv.Itoa(code)
			if val, ok := types.CharMap[key]; ok {
				result.WriteString(val)
			} else {
				result.WriteString(fmt.Sprintf("[%d]", code))
			}
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
