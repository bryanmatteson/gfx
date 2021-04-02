package type1

import "fmt"

func Parse(b []byte) (err error) {
	lexer := newlexer(b)
	for lexer.next() != nil {
		str := string(lexer.tok.v)
		fmt.Println(str)
	}

	return
}
