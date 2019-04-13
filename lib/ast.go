package lib

import (
	"fmt"
	"strings"
)

type Tag struct {
	col int
	row int
	pos int

	Name    string
	Options map[string]string
	Nested  []*Tag
	Content []byte
}

func NewTag(lex *Lexer, name string, options map[string]string) *Tag {
	return &Tag{
		col: lex.col,
		row: lex.row,
		pos: lex.pos,

		Name:    name,
		Options: options,
	}
}

func NewTextTag(lex *Lexer, content []byte) *Tag {
	return &Tag{
		col: lex.col,
		row: lex.row,
		pos: lex.pos,

		Name:    "text",
		Options: nil,
		Content: content,
	}
}

func (tag *Tag) string(prefix string) string {
	b := new(strings.Builder)
	b.WriteString(prefix)
	b.WriteString("@")
	b.WriteString(tag.Name)

	b.WriteString("(")
	options := make([]string, 0)
	for k, v := range tag.Options {
		options = append(options, fmt.Sprintf(`%q: "%q"`, k, v))
	}
	b.WriteString(strings.Join(options, ", "))

	b.WriteString(") {\n")
	for _, t := range tag.Nested {
		b.WriteString(t.string(prefix + "    "))
	}
	b.WriteString(prefix)

	if len(tag.Content) > 0 {
		b.WriteString(fmt.Sprintf("  %q\n", tag.Content))
		b.WriteString(prefix)
	}
	b.WriteString("}\n")

	return b.String()
}

func (tag *Tag) String() string {
	return tag.string("")
}

func (tag *Tag) Validate() error {
	return Validate(tag)
}

func (tag *Tag) HTML() string {
	err := tag.Validate()
	if err != nil {
		return ""
	}
	return Marshall(tag)
}
