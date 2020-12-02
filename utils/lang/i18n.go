package lang

import (
	"encoding/json"
	"errors"
	"github.com/omecodes/common/futils"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Entry struct {
	Key   string
	Value string
}

func NewManager(dir string) *I18n {
	return &I18n{dir: dir}
}

type I18n struct {
	dir             string
	defaultLanguage language.Tag
	matcher         language.Matcher
	updated         bool
}

func (i *I18n) Load() error {
	langDir, err := os.Open(i.dir)
	if err != nil {
		log.Println("could not read i18n dir")
		return err
	}
	names, err := langDir.Readdirnames(-1)
	if err != nil {
		log.Println("could not read i18n dir")
		return err
	}

	for _, name := range names {
		text := map[string]string{}
		fullPath := filepath.Join(i.dir, name)
		var extension = filepath.Ext(name)
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			log.Printf("[i18n]\ncould not read %s content: %s\n", fullPath, err)
			continue
		}

		if extension == ".yml" {
			err = yaml.Unmarshal(content, &text)
			if err != nil {
				log.Printf("[i18n]\ncould not parse %s content: %s\n", fullPath, err)
				continue
			}
		} else {
			err = json.Unmarshal(content, &text)
			if err != nil {
				log.Printf("[i18n]\ncould not parse %s content: %s\n", fullPath, err)
				continue
			}
		}

		name = name[0 : len(name)-len(extension)]
		tag, err := language.Parse(name)
		if err != nil {
			log.Printf("[i18n]\t%s is not a knwon language name: %s\n", name, err)
		} else {
			for key, value := range text {
				entry := Entry{
					Key:   key,
					Value: value,
				}
				err = i.AddEntry(tag, entry)
				if err != nil {
					log.Printf("[i18n]\tcould not register %v entry for language %s: %s\n", entry, tag, err)
				}
			}
		}
	}
	return nil
}

func (i *I18n) Translations(tag language.Tag) (map[string]string, error) {
	locale := tag.String()
	filename := filepath.Join(i.dir, locale+".yml")
	if futils.FileExists(filename) {
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		translations := map[string]string{}
		err = yaml.Unmarshal(content, &translations)
		if err != nil {
			return nil, err
		}
		return translations, nil
	}

	filename = filepath.Join(i.dir, locale+".json")
	if futils.FileExists(filename) {
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		translations := map[string]string{}
		err = json.Unmarshal(content, &translations)
		if err != nil {
			return nil, err
		}
		return translations, nil
	}

	if locale != i.defaultLanguage.String() {
		return i.Translations(i.defaultLanguage)
	}

	return nil, errors.New("not found")
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
