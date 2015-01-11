// copyright 2015 by mathias gumz. all rights reserved. see the LICENSE
// file for more information.

// package 'css' helps to reduce the file-size of a given .css-file.
// it takes a very simple and naive approach to css-minification:
//
//  * ignore linebreaks
//  * condense multiples spaces/tabs into one " "
//  * ignore whitespace before/after "{", ",", ":", ".", ";"
//  * ignore comments
//
//  * add linebreaks after "}" (optional)
//
package css

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/gorilla/css/scanner"
)

const PREFIX_EXCL_COMMENT = "/*!"

// reads from 'in' and applies the rules sketched out above to reduce
// the bytes written to 'out' by not break anything. the output can be
// tweaked by adding several config functions, see css.AddLineBreaks and
// css.KeepExclamationComments
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
			if r.cur.Type == scanner.TokenComment {
				if r.config.keep_exclamation_comment([]byte(r.cur.Value)) {
					io.WriteString(out, r.cur.Value)
					r.cur.Value = ""
					r.on_hold = r.cur
					continue
				}
			}
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
				r.config.linebreak(out)
				continue
			case ";", "{", ",", ".", ":":
				if r.on_hold != nil && r.on_hold.Type == scanner.TokenS {
					r.on_hold = nil
				}
				r.on_hold = write_on_change(out, r.on_hold, r.cur)
				continue
			}
		}
		// default-mode: copy each char. using 'nil' as the
		// 2nd parameter ensures this.
		r.on_hold = write_on_change(out, r.on_hold, nil)
		io.WriteString(out, r.cur.Value)
	}
	return
}

// func-type to tweak the way css.Reduce() works
type Config func(*reducer) error

// add linebreaks after each rule
func AddLineBreaks(r *reducer) error {
	r.config.linebreak = func(w io.Writer) { io.WriteString(w, "\n") }
	return nil
}

// keeps comments like /*! this */
func KeepExclamationComments(r *reducer) error {
	r.config.keep_exclamation_comment = func(comment []byte) bool {
		return bytes.HasPrefix(comment, []byte(PREFIX_EXCL_COMMENT))
	}
	return nil
}

// ========================================================================
// internal
//

type (
	reducer struct {
		*scanner.Scanner
		cur     *scanner.Token
		on_hold *scanner.Token // postpone the last token
		config  rconfig
	}
	rconfig struct {
		linebreak                func(io.Writer)
		keep_exclamation_comment func([]byte) bool
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
		config: rconfig{
			linebreak:                func(io.Writer) {}, // nop
			keep_exclamation_comment: func([]byte) bool { return false },
		},
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
