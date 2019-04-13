package lib

import (
	"fmt"
	"strings"
	"errors"
)

type ValidationError struct {
	msg string
	col int
	row int
	pos int
}

func NewValidationError(tag *Tag, msg string, params ...interface{}) *ValidationError {
	msg = fmt.Sprintf(msg, params...)
	msg = strings.Replace(msg, "\n", "\\n", -1)
	return &ValidationError{msg: msg, col: tag.col, row: tag.row, pos: tag.pos}
}

func (err ValidationError) Error() string {
	return fmt.Sprintf("Validation Error: %s (line %d, column %d)", err.msg, err.row+1, err.col+1)
}

type Validator struct {
	errors []*ValidationError
}

func NewValidator() *Validator {
	return &Validator{errors: make([]*ValidationError, 0)}
}

func (v *Validator) addError(tag *Tag, msg string, params ...interface{}) {
	err := NewValidationError(tag, msg, params...)
	v.errors = append(v.errors, err)
}

func (v *Validator) validateParagraphTag(tag *Tag) {
	align, ok := tag.Options["alignment"]
	if ok {
		switch align {
		case "left", "right", "center":
			return
		default:
			v.addError(tag, "alignment must be: left, right, center")
		}
	}
}

func (v *Validator) validateImageTag(tag *Tag) {
	_, ok := tag.Options["uri"]
	if !ok {
		v.addError(tag, "uri parameter is required")
	}
}

func (v *Validator) validateLinkTag(tag *Tag) {
	_, ok := tag.Options["uri"]
	if !ok {
		v.addError(tag, "uri parameter is required")
	}
}

func (v *Validator) validateHeadingTag(tag *Tag) {
	align, ok := tag.Options["alignment"]
	if ok {
		switch align {
		case "left", "right", "center":
			return
		default:
			v.addError(tag, "alignment must be: left, right, center")
		}
	}

	level, ok := tag.Options["level"]
	if ok {
		switch level {
		case "1", "2", "3", "4", "5":
			return
		default:
			v.addError(tag, "level must be: 1, 2, 3, 4, 5")
		}
	}
}

func (v *Validator) validateListTag(tag *Tag) {
	style, ok := tag.Options["style"]
	if ok {
		switch style {
		case "unordered", "ordered":
			return
		default:
			v.addError(tag, "style must be: unordered, ordered")
		}
	}
}

func (v *Validator) validate(tag *Tag) {
	switch tag.Name {
	case "paragraph":
		v.validateParagraphTag(tag)
	case "heading":
		v.validateHeadingTag(tag)
	case "list":
		v.validateListTag(tag)
	case "image":
		v.validateImageTag(tag)
	case "link":
		v.validateLinkTag(tag)
	}

	for _, nested := range tag.Nested {
		v.validate(nested)
	}
}

func Validate(tag *Tag) error {
	v := NewValidator()
	if tag.Name != "main" {
		v.addError(tag, "root tag must be @main")
	}

	v.validate(tag)
	if len(v.errors) > 0 {
		msgs := make([]string, 0)
		for _, err := range v.errors {
			msgs = append(msgs, err.Error())
		}
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}
