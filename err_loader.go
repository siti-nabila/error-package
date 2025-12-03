package errorpackage

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

type (
	ErrCode *string
	ErrPack struct {
		Code     ErrCode `yaml:"code"`
		Messages []LangPack
	}

	Errors         map[string][]error
	DictionaryPack struct {
		Errors   map[string]map[LanguageCode]string `yaml:"errors"`
		Packages map[string]ErrPack
	}
	Error struct {
		code          ErrCode
		err           error
		localMessages map[LanguageCode]string
	}
)

var (
	Language LanguageCode
)

func NewErrYamlPackage() DictionaryPack {
	return DictionaryPack{
		Errors:   make(map[string]map[LanguageCode]string),
		Packages: make(map[string]ErrPack),
	}
}

func (e Error) Code() ErrCode {
	return e.code
}

func (e Error) Error() string {
	if e.err != nil {
		return e.localizedMessage()
	}

	return fmt.Sprintf("validator: something went wrong (code: %d)", e.code)
}

func (e Error) localizedMessage() string {
	if Language == "" {
		Language = DafaultLocale
	}
	if msg, ok := e.localMessages[Language]; ok {
		return msg
	}
	return e.localMessages[DafaultLocale]
}

func (d *DictionaryPack) LoadBytes(data []byte) error {
	return d.collectErrors(data)
}

func (d *DictionaryPack) collectErrors(data []byte) error {
	if err := yaml.Unmarshal(data, d); err != nil {
		return err
	}
	fmt.Println("validator: collecting errors from yaml.")

	for k, v := range d.Errors {
		fmt.Printf("validator: registering: %s.\n", k)
		var code *string
		if cd, exists := v["code"]; exists {
			code = &cd
		}
		d.Packages[k] = ErrPack{
			Code: ErrCode(code),
			Messages: []LangPack{
				{Code: English, Message: v[English]},
				{Code: Indonesia, Message: v[Indonesia]},
			},
		}
	}
	fmt.Println("validator: done collecting errors from yaml.")
	return nil
}

func (d *DictionaryPack) New(key string) Error {
	errValidator := fmt.Sprintf("validator: %s", key)
	err := Error{
		code:          nil,
		err:           errors.New(errValidator),
		localMessages: make(map[LanguageCode]string),
	}

	if pack, exists := d.Packages[key]; exists {
		err.code = pack.Code
		for _, msg := range pack.Messages {
			err.localMessages[msg.Code] = msg.Message
		}
	}

	return err
}

func (d *DictionaryPack) Newf(key string, args ...any) Error {
	e := d.New(key)
	for code, msg := range e.localMessages {
		e.localMessages[code] = fmt.Sprintf(msg, args...)
	}
	return e
}

func SetLanguage(lang string) {
	Language = LanguageCode(lang)
}

func (errs Errors) Error() string {
	if len(errs) == 0 {
		return ""
	}

	var b strings.Builder
	keys := sortedKeys(errs)

	for i, key := range keys {
		if i > 0 {
			b.WriteString("; ")
		}

		// Format value (slice of errors)
		b.WriteString(formatErrorList(key, errs[key]))
	}

	b.WriteString(".")
	return b.String()
}
func (errs Errors) LocalizedError() map[string]any {
	if len(errs) == 0 {
		return map[string]any{}
	}

	result := make(map[string]any)

	for key, list := range errs {
		result[key] = localizeErrorList(list, Language)
	}

	return result
}
func (e Errors) Add(field string, err error) {
	e[field] = append(e[field], err)
}
func (errs Errors) Merge(other Errors) {
	for field, list := range other {
		errs[field] = append(errs[field], list...)
	}
}

func (errs Errors) Empty() bool {
	return len(errs) == 0
}

func sortedKeys(m Errors) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func formatErrorList(field string, list []error) string {
	var b strings.Builder

	if len(list) == 1 {
		// Single error → tampilkan langsung
		fmt.Fprintf(&b, "%s: %s", field, formatSingleError(list[0]))
		return b.String()
	}

	// Multiple → tampilkan array
	fmt.Fprintf(&b, "%s: [", field)
	for i, err := range list {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(formatSingleError(err))
	}
	b.WriteString("]")

	return b.String()
}

func formatSingleError(err error) string {
	switch e := err.(type) {
	case Errors:
		return e.Error()
	case Error: // custom error kamu
		return e.Error()
	default:
		return err.Error()
	}
}

func localizeErrorList(list []error, lang LanguageCode) any {
	// Single error → langsung return string
	if len(list) == 1 {
		return localizeSingleError(list[0], lang)
	}

	// Multiple → array of strings or nested maps
	arr := make([]any, len(list))
	for i, err := range list {
		arr[i] = localizeSingleError(err, lang)
	}
	return arr
}

func localizeSingleError(err error, lang LanguageCode) any {
	switch e := err.(type) {
	case Errors:
		return e.LocalizedError()
	case Error: // custom error kamu
		return e.localizedMessage()
	default:
		return err.Error()
	}
}
