package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Define a new Validator type which contains a map of validation errors for our
// form fields.
type Validator struct {
	Errors map[string]string
}

// Valid() returns true if the Errors map doesn't contain any entries.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddFieldError() adds an error message to the Errors map (so long as no
// entry already exists for the given key).
func (v *Validator) AddFieldError(key, message string) {
	// Note: We need to initialize the map first, if it isn't already
	// initialized.
	if v.Errors == nil {
		v.Errors = make(map[string]string)
	}
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// CheckField() adds an error message to the Errors map only if a
// validation check is not 'ok'.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank() returns true if a value is not an empty string.
func (v *Validator) NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars() returns true if a value contains no more than n characters.
func (v *Validator) MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func (v *Validator) ValidEmail(email string) bool {
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailPattern.MatchString(email)
}

// PermittedInt() returns true if a value is in a list of permitted integers.
func PermittedInt(value int, permittedValues ...int) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}
