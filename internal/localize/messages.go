package localize

import (
	"bytes"
	"html/template"
	"strings"
)

// A singleMessage contains a translation string or if used as a template a
// parsed template.Template.
type singleMessage struct {
	S string
	T *template.Template
}

func (s singleMessage) IsZero() bool {
	return s.S == "" && s.T == nil
}

func (s singleMessage) Execute(data any) string {
	if s.T == nil {
		return s.S
	}

	var buf bytes.Buffer
	s.T.Execute(&buf, data)
	return buf.String()
}

func newSingleMessage(s string, fns template.FuncMap) singleMessage {
	if strings.Contains(s, "{{") {
		if t, err := template.New("").Funcs(fns).Parse(s); err == nil {
			return singleMessage{T: t}
		}
	}

	return singleMessage{S: s}
}

// pluralMessage contains the different options for plural translations.
type pluralMessage struct {
	One   singleMessage
	Other singleMessage

	// for Welsh only
	Two  singleMessage
	Few  singleMessage
	Many singleMessage
}

type Messages struct {
	Singles map[string]singleMessage
	Plurals map[string]pluralMessage
}

func (m Messages) Find(key string) (singleMessage, bool) {
	if msg, ok := m.Singles[key]; ok {
		return singleMessage(msg), true
	}

	return singleMessage{}, false
}

func (m Messages) FindPlural(key string, count int) (singleMessage, bool) {
	msg, ok := m.Plurals[key]
	if !ok {
		return singleMessage{}, false
	}

	if count == 1 {
		return msg.One, true
	}

	if count == 2 && !msg.Two.IsZero() {
		return msg.Two, true
	}

	if count == 3 && !msg.Few.IsZero() {
		return msg.Few, true
	}

	if count == 6 && !msg.Many.IsZero() {
		return msg.Many, true
	}

	return msg.Other, true
}
