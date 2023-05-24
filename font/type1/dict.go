package type1

import "github.com/bryanmatteson/gfx/font/adobe"

type PrivateDictionary struct {
	adobe.PrivateDictionary

	UniqueID    int
	IVLen       int
	RoundStemUp bool
	Password    int
}
