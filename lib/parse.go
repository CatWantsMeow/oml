package lib

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	tagDeclarationRegexp = regexp.MustCompile(`^[\w]+$`)
	tagOptionRegexp      = regexp.MustCompile(`^[\s]*([\w-]+)[\s]*:[\s]*"([^\n,"]*)"[\s]*$`)
)

type syntaxError struct {
	msg string
	col int
	row int
	pos int
}

type SyntaxError interface {
	error
}

func NewSyntaxError(lex *Lexer, msg string, params ...interface{}) SyntaxError {
	msg = fmt.Sprintf(msg, params...)
	msg = strings.Replace(msg, "\n", "\\n", -1)
	return &syntaxError{msg: msg, col: lex.col, row: lex.row, pos: lex.pos}
}

func (err syntaxError) Error() string {
	return fmt.Sprintf("Syntax Error: %s (line %d, column %d)", err.msg, err.row+1, err.col+1)
}

type Lexer struct {
	buf     *bytes.Buffer
	pos     int
	col     int
	row     int
	content []byte
}

func NewLexer(content []byte) *Lexer {
	return &Lexer{
		buf:     new(bytes.Buffer),
		pos:     -1,
		col:     -1,
		row:     0,
		content: content,
	}
}

func (lex *Lexer) cur() byte {
	if lex.pos >= len(lex.content) {
		return 0
	}
	return lex.content[lex.pos]
}

func (lex *Lexer) next() byte {
	lex.pos++
	lex.col++
	for lex.pos < len(lex.content) && lex.content[lex.pos] == '\n' {
		lex.pos++
		lex.row++
		lex.col = 0
	}
	if lex.pos >= len(lex.content) {
		return 0
	}
	return lex.content[lex.pos]
}

func (lex *Lexer) skipAllowed(stops, allowed []byte) SyntaxError {
mainLoop:
	for b := lex.cur(); b != 0; b = lex.next() {
		for _, c := range stops {
			if b == c {
				return nil
			}
		}

		if allowed != nil {
			for _, c := range allowed {
				if b == c {
					continue mainLoop
				}
			}
			return NewSyntaxError(lex, "unexpected token")
		}
	}
	return NewSyntaxError(lex, "unexpected end of content")
}

func (lex *Lexer) skipDisallowed(stops, disallowed []byte) SyntaxError {
	for b := lex.cur(); b != 0; b = lex.next() {
		for _, c := range stops {
			if b == c {
				return nil
			}
		}

		if disallowed != nil {
			for _, c := range disallowed {
				if b == c {
					return NewSyntaxError(lex, "unexpected token")
				}
			}
		}
	}
	return NewSyntaxError(lex, "unexpected end of content")
}

func (lex *Lexer) parseTagOptions() (map[string]string, SyntaxError) {
	if lex.cur() != '(' {
		return nil, nil
	}

	lex.next()
	start := lex.pos
	if err := lex.skipDisallowed([]byte{')'}, []byte{'{', '}', '(', '@'}); err != nil {
		return nil, err
	}

	options := make(map[string]string)
	parts := bytes.Split(lex.content[start:lex.pos], []byte{','})
	for _, part := range parts {
		if !tagOptionRegexp.Match(part) {
			return nil, NewSyntaxError(lex, "invalid tag option format `%s`", part)
		}

		matches := tagOptionRegexp.FindSubmatch(part)
		if len(matches) != 3 {
			return nil, NewSyntaxError(lex, "unexpected tag options inconsistency")
		}
		options[string(matches[1])] = string(matches[2])
	}
	lex.next()
	return options, nil
}

func (lex *Lexer) parseTagDefinition() (*Tag, SyntaxError) {
	start := lex.pos + 1
	if err := lex.skipAllowed([]byte{'{', '(', ' ', '\n'}, nil); err != nil {
		return nil, err
	}

	name := lex.content[start:lex.pos]
	if !tagDeclarationRegexp.Match(name) {
		return nil, NewSyntaxError(lex, "invalid tag name `%s`", name)
	}

	if err := lex.skipAllowed([]byte{'{', '('}, []byte{' '}); err != nil {
		return nil, err
	}

	options, err := lex.parseTagOptions()
	if err != nil {
		return nil, err
	}

	if err := lex.skipAllowed([]byte{'{'}, []byte{' '}); err != nil {
		return nil, err
	}

	return NewTag(lex, string(name), options), nil
}

func (lex *Lexer) parse(depth int) ([]*Tag, SyntaxError) {
	tags := make([]*Tag, 0)
	for b := lex.next(); b != 0; b = lex.next() {
		switch {
		case b == '\\':
			b = lex.next()
			if b == 'n' {
				lex.buf.WriteByte('\n')
			} else if b != 0 {
				lex.buf.WriteByte(b)
			}

		case b == '@':
			buf := lex.buf.Bytes()
			if len(buf) > 0 && !ContainsSpacesOnly(buf) {
				if len(buf) > 0 {
					tags = append(tags, NewTextTag(lex, CompressSpaces(buf)))
					lex.buf = new(bytes.Buffer)
				}
			}

			tag, err := lex.parseTagDefinition()
			if err != nil {
				return nil, err
			}

			nested, err := lex.parse(depth + 1)
			if err != nil {
				return nil, err
			}

			tag.Nested = nested
			tags = append(tags, tag)

		case b == '}':
			buf := lex.buf.Bytes()
			if len(buf) > 0 && !ContainsSpacesOnly(buf) {
				tags = append(tags, NewTextTag(lex, CompressSpaces(buf)))
				lex.buf = new(bytes.Buffer)
			}
			return tags, nil

		case b == '\n':
			lex.buf.WriteByte(' ')

		default:
			lex.buf.WriteByte(b)
		}
	}

	if depth > 0 {
		return nil, NewSyntaxError(lex, "unexpected end of content")
	}
	return tags, nil
}

func Parse(content []byte) (*Tag, error) {
	lex := NewLexer(content)
	tags, err := lex.parse(0)
	if err != nil {
		return nil, err
	}

	if len(tags) == 0 {
		return nil, NewSyntaxError(lex, "no root tag found.")
	} else if len(tags) > 1 {
		return nil, NewSyntaxError(lex, "multiple root tags found.")
	}

	err = tags[0].Validate()
	if err != nil {
		return nil, err
	}

	return tags[0], nil
}
