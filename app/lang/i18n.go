package lang

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log"
)

type Entry struct {
	Key   string
	Value string
}

type I18n struct {
	defaultLanguage language.Tag
	matcher         language.Matcher
	updated         bool
}

func NewTranslations(defaultLanguage language.Tag) *I18n {
	return &I18n{defaultLanguage: defaultLanguage, matcher: nil, updated: false}
}

func (i *I18n) AddEntry(tag language.Tag, entry Entry) error {
	i.updated = true
	return message.SetString(tag, entry.Key, entry.Value)
}

func (i *I18n) LanguageFromAcceptLanguageHeader(header string) language.Tag {
	if i.matcher == nil || i.updated {
		i.matcher = language.NewMatcher(message.DefaultCatalog.Languages())
	}
	tags, _, err := language.ParseAcceptLanguage(header)
	if err != nil {
		log.Println("Accept-Languages header parsing failed:", err)
		tags = []language.Tag{i.defaultLanguage}
	}
	t, _, _ := i.matcher.Match(tags...)
	return t
}

func (i *I18n) Translator(acceptLanguagesHeader string) *message.Printer {
	t := i.LanguageFromAcceptLanguageHeader(acceptLanguagesHeader)
	return message.NewPrinter(t)
}

func (i *I18n) Translated(tag language.Tag, key string, args ...interface{}) string {
	printer := message.NewPrinter(tag)
	return printer.Sprintf(key, args...)
}

func (i *I18n) TranslatedFromAcceptLanguageHeader(acceptLanguagesHeader string, key string, args ...interface{}) string {
	translator := i.Translator(acceptLanguagesHeader)
	return translator.Sprintf(key, args...)
}
