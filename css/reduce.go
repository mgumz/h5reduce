package css

import (
	"io"
	"io/ioutil"

	"github.com/gorilla/css/scanner"
)

type Config func(*reducer) error

// a very simple and naive approach to css-minification:
//
//  * ignore linebreaks
//  * condense multiples spaces/tabs into one " "
//  * ignore whitespace before/after "{", ",", ":", ".", ";"
//  * ignore comments
//
//  * add linebreaks after "}" (optional)
func Reduce(out io.Writer, in io.Reader, configs ...Config) (err error) {

	var r *reducer
	r, err = newReducer(in)
	if err = r.apply(configs...); err != nil {
		return err
	}

	for {
		r.cur = r.Scanner.Next()
		if r.cur.Type == scanner.TokenEOF || r.cur.Type == scanner.TokenError {
			break
		}

		switch r.cur.Type {
		case scanner.TokenComment, scanner.TokenS:
			if r.on_hold == nil && (r.cur.Value == " " || r.cur.Value == "\t") {
				r.cur.Value = " "
				r.on_hold = write_on_change(out, r.on_hold, r.cur)
			}
			continue
		case scanner.TokenChar:
			switch r.cur.Value {
			case "}":
				r.on_hold = nil
				io.WriteString(out, "}")
				if r.config.linebreaks {
					io.WriteString(out, "\n")
				}
				continue
			case ";", "{", ",", ".", ":":
				if r.on_hold != nil && r.on_hold.Type == scanner.TokenS {
					r.on_hold = nil
				}
				r.on_hold = write_on_change(out, r.on_hold, r.cur)
				continue
			}
		}
		r.on_hold = write_on_change(out, r.on_hold, nil)
		io.WriteString(out, r.cur.Value)
	}
	return
}

func AddLineBreaks(r *reducer) error {
	r.config.linebreaks = true
	return nil
}

// ========================================================================
// internal
//

type (
	reducer struct {
		*scanner.Scanner
		cur     *scanner.Token
		on_hold *scanner.Token
		config  rconfig
	}
	rconfig struct {
		linebreaks bool
	}
)

func newReducer(reader io.Reader) (r *reducer, err error) {
	var css []byte
	if css, err = ioutil.ReadAll(reader); err != nil {
		return
	}
	return &reducer{
		Scanner: scanner.New(string(css)),
		cur:     &scanner.Token{Type: scanner.TokenEOF},
	}, nil
}

func (r *reducer) apply(configs ...Config) (err error) {
	for _, fn := range configs {
		if err = fn(r); err != nil {
			break
		}
	}
	return
}

// write a.Value to 'out' if
// * 'b' is nil
// * 'a' and 'b' are of the same type but have different values
func write_on_change(out io.Writer, a, b *scanner.Token) *scanner.Token {
	if a != nil && b == nil {
		io.WriteString(out, a.Value)
	}
	if a != nil && b != nil &&
		(a.Type != b.Type ||
			(a.Type == scanner.TokenChar && a.Value != b.Value)) {
		io.WriteString(out, a.Value)
	}
	return b
}

func write_strings(w io.Writer, s ...string) {
	for i := range s {
		io.WriteString(w, s[i])
	}
}
