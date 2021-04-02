package type1

import "go.matteson.dev/gfx/font/adobe"

type PrivateDictionary struct {
	adobe.PrivateDictionary

	UniqueID    int
	IVLen       int
	RoundStemUp bool
	Password    int
}
