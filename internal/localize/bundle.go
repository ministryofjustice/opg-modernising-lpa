package localize

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

type parsedMessage struct {
	S string

	// when plural
	One   string
	Other string

	// for Welsh only
	Two  string
	Few  string
	Many string
}

func (m *parsedMessage) UnmarshalJSON(text []byte) error {
	var s string
	if err := json.Unmarshal(text, &s); err == nil {
		m.S = s
		return nil
	}

	var v map[string]string
	if err := json.Unmarshal(text, &v); err == nil {
		m.One = v["one"]
		m.Other = v["other"]
		m.Two = v["two"]
		m.Few = v["few"]
		m.Many = v["many"]
		return nil
	}

	return errors.New("message malformed")
}

type Bundle struct {
	messages map[string]Messages
}

func NewBundle(paths ...string) (*Bundle, error) {
	bundle := &Bundle{messages: map[string]Messages{}}

	for _, path := range paths {
		if err := bundle.LoadMessageFile(path); err != nil {
			return nil, err
		}
	}

	return bundle, nil
}

func (b *Bundle) LoadMessageFile(p string) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}

	var v map[string]parsedMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	lang, _ := strings.CutSuffix(path.Base(p), ".json")

	fns := map[string]any{
		"lowerFirst": LowerFirst,
	}

	if lang == "en" {
		if err := verifyEn(v); err != nil {
			return err
		}
		fns["possessive"] = func(s string) string {
			format := "%s’s"

			if strings.HasSuffix(s, "s") {
				format = "%s’"
			}

			return fmt.Sprintf(format, s)
		}
	} else if lang == "cy" {
		if err := verifyCy(v); err != nil {
			return err
		}
		fns["possessive"] = func(s string) string { return s }
	} else {
		return errors.New("only supports en or cy")
	}

	messages := Messages{
		Singles: map[string]singleMessage{},
		Plurals: map[string]pluralMessage{},
	}

	for key, message := range v {
		if message.S != "" {
			messages.Singles[key] = newSingleMessage(message.S, fns)
		} else {
			messages.Plurals[key] = pluralMessage{
				One:   newSingleMessage(message.One, fns),
				Two:   newSingleMessage(message.Two, fns),
				Few:   newSingleMessage(message.Few, fns),
				Many:  newSingleMessage(message.Many, fns),
				Other: newSingleMessage(message.Other, fns),
			}
		}
	}

	b.messages[lang] = messages
	return nil
}

func verifyEn(v map[string]parsedMessage) error {
	for key, message := range v {
		if message.S != "" {
			continue
		}

		if message.One != "" && message.Other != "" && message.Two == "" && message.Few == "" && message.Many == "" {
			continue
		}

		return fmt.Errorf("problem with key: %s", key)
	}

	return nil
}

func verifyCy(v map[string]parsedMessage) error {
	for key, message := range v {
		if message.S != "" {
			continue
		}

		if message.One != "" && message.Other != "" && message.Two != "" && message.Few != "" && message.Many != "" {
			continue
		}

		return fmt.Errorf("problem with key: %s", key)
	}

	return nil
}

func (b *Bundle) For(lang Lang) *Localizer {
	return &Localizer{b.messages[lang.String()], false, lang}
}
