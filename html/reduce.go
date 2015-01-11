// copyright 2015 by mathias gumz. all rights reserved. see the LICENSE
// file for more information.

// package 'html' offers means to reduce the size of a given html-document
// it's similar naive and simplistic as 'css' but it yields good-enough
// results to be usable.
package html

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"

	exp_html "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// reads html-document from 'reader' and writes the reduced form of the document
// to 'writer'. example:
//
//     ReduceFile(os.Stdout, "index.html", AddExtraNewlines)
//
// 'Reduce' tries to not change any semantics of the given document. eg, it
// lets <script>, <style> and <pre> tokens "mostly" unaltered (no wspace changes,
// just some html-escaping inside <pre>)
func Reduce(writer io.Writer, reader io.Reader, configs ...Config) (err error) {
	var z *tokenizer
	if z, err = newTokenizer(writer, reader, configs...); err != nil {
		return err
	}
	for reduce, tt := reduceDoctype, exp_html.DoctypeToken; z.err == nil; {
		if tt = z.Tokenizer.Next(); tt == exp_html.ErrorToken {
			if z.Tokenizer.Err() != io.EOF {
				return z.Tokenizer.Err()
			}
			break
		}
		reduce = reduce(tt, z)
	}
	return z.err
}

// convinience wrapper around Reduce(): opens a file 'path' and writes the
// reduced form of the document to 'writer'. see Reduce() for more information.
func ReduceFile(writer io.Writer, path string, configs ...Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return Reduce(writer, file, configs...)
}

// func-type to tweak the way html.Reduce() works
type Config func(*tokenizer) error

// config-option: add a newline after certain html-tags to keep
// some readability in the reduced document. newlines are added after
// <!doctype html>, <html>,</html>, <head>,</head>, </title>,
// </script>, </style>, <link>, <body>,</body>
func AddExtraNewlines(z *tokenizer) error {
	z.config.extra_newline = func() { _, z.err = fmt.Fprintln(z) }
	return nil
}

// see http://godoc.org/golang.org/x/net/html#Tokenizer.AllowCDATA
func AllowCDATA(z *tokenizer) error {
	z.Tokenizer.AllowCDATA(true)
	return nil
}

func DontStripComments(z *tokenizer) error {
	z.config.strip_comments = false
	return nil
}

// ========================================================================
// internal
//

type (
	tokenizer struct {
		*exp_html.Tokenizer
		io.Writer
		prev   reduceFN
		config tkconfig
		err    error
	}
	tkconfig struct {
		extra_newline  func()
		strip_comments bool
	}
	reduceFN func(tt exp_html.TokenType, z *tokenizer) reduceFN
)

var nop = func() {}

func newTokenizer(w io.Writer, r io.Reader, configs ...Config) (z *tokenizer, err error) {
	z = &tokenizer{
		Tokenizer: exp_html.NewTokenizer(r),
		Writer:    w,
		config:    tkconfig{extra_newline: nop, strip_comments: true},
	}
	return z.apply(configs...)
}

func (z *tokenizer) apply(configs ...Config) (_ *tokenizer, err error) {
	for _, fn := range configs {
		if err = fn(z); err != nil {
			return nil, err
		}
	}
	return z, nil
}

func reduceDoctype(tt exp_html.TokenType, z *tokenizer) reduceFN {
	if tt == exp_html.DoctypeToken {
		if _, z.err = z.Write(z.Raw()); z.err != nil {
			return nil
		}
		z.config.extra_newline()
		return reduceHead
	}
	return reduceHead(tt, z)
}

func reduceHead(tt exp_html.TokenType, z *tokenizer) reduceFN {
	text, next := z.Raw(), reduceHead
	switch tt {
	case exp_html.CommentToken:
		if z.config.strip_comments && !is_conditional_comment(text) {
			return next
		}
		defer z.config.extra_newline()
	case exp_html.TextToken:
		// TODO: might be a problem with html5:
		//   <title>foo</title>
		//   that's the way already
		//   yeay eah yeah.
		text = bytes.TrimSpace(text)
	case exp_html.StartTagToken:
		token := z.Token()
		text = []byte(token.String())
		switch token.DataAtom {
		case atom.Script, atom.Style:
			z.prev, next = reduceHead, passthru
		case atom.Html, atom.Head, atom.Link, atom.Meta:
			defer z.config.extra_newline()
		case atom.Title:
		default:
			next = reduceBody
		}
	case exp_html.SelfClosingTagToken:
		defer z.config.extra_newline()
	case exp_html.EndTagToken:
		token := z.Token()
		text = []byte(token.String())
		switch token.DataAtom {
		case atom.Head:
			next = reduceBody
		}
		defer z.config.extra_newline()
	}
	_, z.err = z.Write(text)
	return next
}

func reduceBody(tt exp_html.TokenType, z *tokenizer) reduceFN {
	text, next := z.Raw(), reduceBody
	switch tt {
	case exp_html.CommentToken:
		if z.config.strip_comments && !is_conditional_comment(text) {
			return next
		}
		defer z.config.extra_newline()
	case exp_html.TextToken:
		text = reduce_ws(text)
	case exp_html.StartTagToken:
		token := z.Token()
		switch token.DataAtom {
		case atom.Pre, atom.Script, atom.Style:
			z.prev, next = reduceBody, passthru
			defer z.config.extra_newline()
		case atom.Body:
			defer z.config.extra_newline()
		}
		// start-tags do not contain extra-whitespaces which
		// affect the visual rendering of the dom. thus, we
		// throw any 'raw' representation of them away
		text = []byte(token.String())
	case exp_html.SelfClosingTagToken:
		text = []byte(z.Token().String())
	case exp_html.EndTagToken:
		token := z.Token()
		text = []byte(token.String())
		if token.DataAtom == atom.Body {
			z.config.extra_newline()
			defer z.config.extra_newline()
		}
	}
	_, z.err = z.Write(text)
	return next
}

func passthru(tt exp_html.TokenType, z *tokenizer) reduceFN {
	text, next := z.Raw(), passthru
	if tt == exp_html.EndTagToken {
		token := z.Token()
		text = []byte(token.String())
		switch token.DataAtom {
		case atom.Script, atom.Style:
			defer z.config.extra_newline()
			next, z.prev = z.prev, nil
		case atom.Pre:
			// TODO: might be a problem, <pre>
			// token might contain other blocks
			next, z.prev = z.prev, nil
		}
	}
	_, z.err = z.Write(text)
	return next
}

// ========================================================================
// additional helper functions
//

func reduce_ws(in []byte) []byte {
	var (
		reader  = bytes.NewReader(in)
		writer  = bytes.NewBuffer(in[:0])
		r, prev rune
		err     error
	)

	for {
		if r, _, err = reader.ReadRune(); err != nil {
			break
		}
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			r = ' '
			if !unicode.IsSpace(prev) {
				writer.WriteRune(r)
			}
		} else {
			writer.WriteRune(r)
		}
		prev = r
	}
	if unicode.IsSpace(prev) && writer.Len() == 1 {
		return in[:0]
	}
	return writer.Bytes()
}

// allthough now obsolete (see http://msdn.microsoft.com/en-us/library/ie/hh801214(v=vs.85).aspx )
// we have to handle conditional tags (see http://msdn.microsoft.com/library/ms537512.aspx )
//
// 'in' is expected to be a 'full' html-comment: including the prefix
// `<!--`.
func is_conditional_comment(in []byte) bool {
	if len(in) < 5 {
		return false
	}
	// skip the first 4 bytes `<!--`
	lc := bytes.ToLower(in[4:])
	if bytes.Contains(lc, []byte("[if")) {
		return true
	} else if bytes.Contains(lc, []byte("[endif")) {
		return true
	}
	return false
}
