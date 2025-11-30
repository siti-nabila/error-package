package errorpackage

import (
	"golang.org/x/text/language"
)

type (
	LanguageCode string
	LangPack     struct {
		Code    LanguageCode
		Message string
	}
)

var (
	Indonesia     LanguageCode = LanguageCode(language.Indonesian.String())
	English       LanguageCode = LanguageCode(language.English.String())
	DafaultLocale LanguageCode = English
)
