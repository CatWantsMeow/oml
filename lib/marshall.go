package lib

import (
	"bytes"
	"fmt"
	"strings"
)

type HTMLMarshaller struct {
	buf strings.Builder
}

func (m *HTMLMarshaller) marshallTextTag(tag *Tag) {
	content := bytes.Replace(tag.Content, []byte{'\n'}, []byte("<br/>"), -1)
	m.buf.Write(content)
}

func (m *HTMLMarshaller) marshallMainTag(tag *Tag) {
	m.buf.WriteString(`<div class="main-tag">`)
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</div>`)
}

func (m *HTMLMarshaller) marshallParagraphTag(tag *Tag) {
	m.buf.WriteString(fmt.Sprintf(
		`<p class="text-%s">`,
		Get(tag.Options, "alignment", "left"),
	))
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</p>`)
}

func (m *HTMLMarshaller) marshallImagehTag(tag *Tag) {
	m.buf.WriteString(fmt.Sprintf(
		`<img src="%s" class="rounded mx-auto d-block" >`,
		tag.Options["uri"],
	))
}

func (m *HTMLMarshaller) marshallLinkTag(tag *Tag) {
	m.buf.WriteString(fmt.Sprintf(`<a href="%s">`, tag.Options["uri"]))
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</a>`)
}

func (m *HTMLMarshaller) marshallColumnsTag(tag *Tag) {
	m.buf.WriteString(`<div class="row">`)
	for _, t := range tag.Nested {
		m.buf.WriteString(`<div class="col-sm">`)
		m.marshall(t)
		m.buf.WriteString(`</div>`)
	}
	m.buf.WriteString(`</div>`)
}

func (m *HTMLMarshaller) marshallHeadingTag(tag *Tag) {
	m.buf.WriteString(fmt.Sprintf(
		`<p class="h%s text-%s">`,
		Get(tag.Options, "level", "1"),
		Get(tag.Options, "alignment", "left"),
	))
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</p>`)
}

func (m *HTMLMarshaller) marshallListTag(tag *Tag) {
	switch Get(tag.Options, "style", "unordered") {
	case "ordered":
		m.buf.WriteString(`<ol>`)
	default:
		m.buf.WriteString(`<ul>`)
	}

	for _, t := range tag.Nested {
		m.buf.WriteString(`<li>`)
		m.marshall(t)
		m.buf.WriteString(`</li>`)
	}

	switch Get(tag.Options, "style", "unordered") {
	case "ordered":
		m.buf.WriteString(`</ol>`)
	default:
		m.buf.WriteString(`</ul>`)
	}
}

func (m *HTMLMarshaller) marshallBoldTag(tag *Tag) {
	m.buf.WriteString(`<strong>`)
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</strong>`)
}

func (m *HTMLMarshaller) marshallItalicTag(tag *Tag) {
	m.buf.WriteString(`<em>`)
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</em>`)
}

func (m *HTMLMarshaller) marshallUnderlinedTag(tag *Tag) {
	m.buf.WriteString(`<u>`)
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</u>`)
}

func (m *HTMLMarshaller) marshallFontTag(tag *Tag) {
	style := ""
	if _, ok := tag.Options["size"]; ok {
		style += fmt.Sprintf("font-size: %spx; ", tag.Options["size"])
	}
	if _, ok := tag.Options["family"]; ok {
		style += fmt.Sprintf("font-family: %s; ", tag.Options["family"])
	}
	if _, ok := tag.Options["color"]; ok {
		style += fmt.Sprintf("color: %s; ", tag.Options["color"])
	}

	m.buf.WriteString(fmt.Sprintf(`<span style="%s">`, style))
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</span>`)
}

func (m *HTMLMarshaller) marshallUnknownTag(tag *Tag) {
	m.buf.WriteString(`<span>`)
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</span>`)
}

func (m *HTMLMarshaller) marshallBlockTag(tag *Tag) {
	m.buf.WriteString(`<div>`)
	for _, t := range tag.Nested {
		m.marshall(t)
	}
	m.buf.WriteString(`</div>`)
}

func (m *HTMLMarshaller) marshall(tag *Tag) {
	switch tag.Name {
	case "text":
		m.marshallTextTag(tag)
	case "main":
		m.marshallMainTag(tag)
	case "paragraph":
		m.marshallParagraphTag(tag)
	case "heading":
		m.marshallHeadingTag(tag)
	case "list":
		m.marshallListTag(tag)
	case "bold":
		m.marshallBoldTag(tag)
	case "italic":
		m.marshallItalicTag(tag)
	case "underlined":
		m.marshallUnderlinedTag(tag)
	case "block":
		m.marshallBlockTag(tag)
	case "columns":
		m.marshallColumnsTag(tag)
	case "image":
		m.marshallImagehTag(tag)
	case "link":
		m.marshallLinkTag(tag)
	case "font":
		m.marshallFontTag(tag)
	default:
		m.marshallUnknownTag(tag)
	}
}

func Marshall(tag *Tag) string {
	m := HTMLMarshaller{}
	m.marshall(tag)
	return m.buf.String()
}
