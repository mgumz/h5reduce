package html

import (
	"bytes"
	"testing"
)

type pair struct {
	a string
	b string
}

func TestReduceWhiteSpace(t *testing.T) {

	pairs := []pair{
		{"what is ", "what is "},
		{`what is
	it`, "what is it"},
		{" foo bar ", " foo bar "},
	}

	for i, p := range pairs {
		sample, expected := []byte(p.a), []byte(p.b)
		trimmed := reduceWhiteSpace(sample)
		if bytes.Compare(trimmed, expected) != 0 {
			t.Fatalf("%d: %q => %q vs %q", i, sample, expected, trimmed)
		}
	}
}

func TestReduce(t *testing.T) {

	buf := bytes.Buffer{}
	pairs := []pair{
		{ // simplest html5-doc ever
			`<!DOCTYPE html>
<title>x<title>`,
			`<!DOCTYPE html>
<title>x<title>`},
		{ // head
			`<!doctype html>
		<head>   <title> foo bar </title> </head>`,
			`<!doctype html>
<head>
<title>foo bar</title>
</head>
`},
		{ // blocks vs inline
			`<div>foo bar <span> bazz

			</span> ban </div>`,
			`<div>foo bar <span> bazz </span> ban </div>`,
		},
	}

	for i, p := range pairs {
		buf.Reset()
		if err := Reduce(&buf, bytes.NewReader([]byte(p.a))); err != nil {
			t.Fatalf("%i:", err)
			break
		}
		if bytes.Compare(buf.Bytes(), []byte(p.b)) != 0 {
			t.Fatalf("pair %d: %q => %q vs %q", i, p.a, p.b, buf.Bytes())
		}
	}
}
