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

	Errors         map[string]error
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
		return e.localizedMessage(Language)
	}

	return fmt.Sprintf("validator: something went wrong (code: %d)", e.code)
}

func (e Error) localizedMessage(lang LanguageCode) string {
	if lang == "" {
		lang = DafaultLocale
	}
	if msg, ok := e.localMessages[lang]; ok {
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
		b.WriteString(formatErrorEntry(key, errs[key]))
	}

	b.WriteString(".")
	return b.String()
}
func (errs Errors) LocalizedError(lang LanguageCode) map[string]any {
	if len(errs) == 0 {
		return map[string]any{}
	}

	result := make(map[string]any)
	for key, err := range errs {
		result[key] = localizeErrValue(err, lang)
	}
	return result
}

func localizeErrValue(err error, lang LanguageCode) any {
	switch e := err.(type) {

	case Errors:
		return e.LocalizedError(lang)

	case Error:
		return e.localizedMessage(lang)

	default:
		return err.Error()
	}
}
func sortedKeys(m map[string]error) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func formatErrorEntry(key string, err error) string {
	switch e := err.(type) {

	case Errors:
		// nested Errors
		return fmt.Sprintf("%s: (%s)", key, e.Error())

	case Error:
		// custom Error type with localization
		msg := e.Error() // already localized based on global Language
		return fmt.Sprintf("%s: %s", key, msg)

	default:
		// fallback
		return fmt.Sprintf("%s: %s", key, err.Error())
	}
}
