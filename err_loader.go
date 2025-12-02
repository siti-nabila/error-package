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
		if Language == "" {
			Language = DafaultLocale
		}
		if message, ok := e.localMessages[Language]; ok {
			return message
		} else {
			return e.localMessages[DafaultLocale]
		}
	}

	return fmt.Sprintf("validator: something went wrong (code: %d)", e.code)
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
		b.WriteString(formatErrorString(key, errs[key]))
	}

	b.WriteString(".")
	return b.String()
}

func sortedKeys(m map[string]error) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func formatErrorString(key string, err error) string {
	switch e := err.(type) {

	case Errors:
		return fmt.Sprintf("%s: (%s)", key, e.Error())

	case Error:
		// Error() sudah mengambil Language global
		return fmt.Sprintf("%s: %s", key, e.Error())

	default:
		return fmt.Sprintf("%s: %s", key, e.Error())
	}
}
