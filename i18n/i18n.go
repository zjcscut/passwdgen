package i18n

import (
	"embed"
	"fmt"
	"fyne.io/fyne/v2"
	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"log"
	"strings"
	"sync"
)

//go:embed bundles/locale.*.toml
var LocaleFS embed.FS

const LocaleTemplate = "bundles/locale.%s.toml"

var refresherMap = make(map[MessageId][]refresher)

var loadedBundleMap = make(map[string]*goi18n.Bundle)

type CanvasLocalizer struct {
	Lang string
	mu   sync.Mutex
}

func (cl *CanvasLocalizer) SwitchLang(canvasObjectsToRefresh []fyne.CanvasObject) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	lang, _, ok := strings.Cut(cl.Lang, "_")
	if !ok {
		log.Fatal("parse lang string failed")
	}
	langTag, le := language.Parse(lang)
	if le != nil {
		log.Fatal("parse language failed")
	}
	bundle, ok := loadedBundleMap[lang]
	if !ok {
		bundle = goi18n.NewBundle(langTag)
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
		_, err := bundle.LoadMessageFileFS(LocaleFS, fmt.Sprintf(LocaleTemplate, lang))
		if err != nil {
			log.Fatal("load message file failed")
		}
		loadedBundleMap[lang] = bundle
	}
	localizer := goi18n.NewLocalizer(bundle, lang)
	if len(refresherMap) > 0 {
		refreshAll(localizer)
		if len(canvasObjectsToRefresh) > 0 {
			for _, co := range canvasObjectsToRefresh {
				co.Refresh()
			}
		}
	}
}

func RegisterRefresher(messageId MessageId, m func(value string)) {
	refresherCache, ok := refresherMap[messageId]
	if !ok {
		refresherCache = make([]refresher, 0)
	}
	refresherCache = append(refresherCache, &defaultRefresher{
		messageId: messageId,
		m:         m,
	})
	refresherMap[messageId] = refresherCache
}

func refreshAll(localizer *goi18n.Localizer) {
	for _, refreshers := range refresherMap {
		for _, refresher := range refreshers {
			refresher.refresh(localizer)
		}
	}
}

type refresher interface {
	refresh(localizer *goi18n.Localizer)
}

type defaultRefresher struct {
	messageId MessageId
	m         func(value string)
}

func (dr *defaultRefresher) refresh(localizer *goi18n.Localizer) {
	value := localizer.MustLocalize(&goi18n.LocalizeConfig{
		MessageID: string(dr.messageId),
	})
	dr.m(value)
}
