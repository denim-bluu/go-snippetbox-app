package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	FieldErrors    map[string]string
	NonFieldErrors []string
}

func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

func (v *Validator) AddNonFieldError(message string) {
	if v.NonFieldErrors == nil {
		v.NonFieldErrors = make([]string, 0)
	}
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

func (v *Validator) AddFieldError(field, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, ok := v.FieldErrors[field]; !ok {
		v.FieldErrors[field] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxStringLength(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func PermitteValues[T comparable](value T, options ...T) bool {
	return slices.Contains(options, value)
}

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func StrPattenMatch(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
