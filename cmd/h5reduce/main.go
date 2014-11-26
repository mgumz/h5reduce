package main

// h5reduce

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"2hoch5.com/h5reduce/css"
	"2hoch5.com/h5reduce/html"
)

const (
	HtmlMode = iota
	CssMode
)

func main() {

	var (
		mode                         = HtmlMode
		in                 io.Reader = os.Stdin
		out                io.Writer = os.Stdout
		iname                        = flag.String("i", "", "input file (default: stdin)")
		oname                        = flag.String("o", "", "output file (default: stdout)")
		print_stats                  = flag.Bool("stats", false, "print stats")
		strip_comments               = flag.Bool("strip-comments", true, "strip away (most) comments")
		keep_excl_comments           = flag.Bool("keep-excl-comments", true, "don't stip away /*! or <!--! comments")
		html_flags                   = struct{ nl, allow_cdata *bool }{
			nl:          flag.Bool("html-nl", true, "place some additional newlines"),
			allow_cdata: flag.Bool("html-cdata", true, "treat cdata as text-token (not as comment)"),
		}
		css_flags = struct{ nl *bool }{
			nl: flag.Bool("css-nl", true, "newlines after each rule"),
		}
		err error
	)

	flag.Parse()

	if *iname != "" {
		f, err := os.Open(*iname)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		in = f

		if filepath.Ext(*iname) == ".css" {
			mode = CssMode
		}
	}

	if len(flag.Args()) > 0 && flag.Args()[0] == "css" {
		mode = CssMode
	}

	if *oname != "" {
		f, err := os.Create(*oname)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		out = f
	}

	if *print_stats {
		var count_writes, count_reads Counter
		defer func() {
			ratio := float32(count_writes) / float32(count_reads)
			fmt.Fprintf(os.Stderr, "\n%.2f\t%d\t%d\n", ratio, count_reads, count_writes)
		}()
		out = io.MultiWriter(out, &count_writes)
		in = io.TeeReader(in, &count_reads)
	}

	switch mode {
	case HtmlMode:
		opts := make([]html.Config, 0)
		if *html_flags.nl {
			opts = append(opts, html.AddExtraNewlines)
		}
		if !*strip_comments {
			opts = append(opts, html.DontStripComments)
		}
		if *html_flags.allow_cdata {
			opts = append(opts, html.AllowCDATA)
		}
		err = html.Reduce(out, in, opts...)
	case CssMode:
		opts := make([]css.Config, 0)
		if *css_flags.nl {
			opts = append(opts, css.AddLineBreaks)
		}
		if *keep_excl_comments {
			opts = append(opts, css.KeepExclamationComments)
		}
		err = css.Reduce(out, in, opts...)
	}

	if err != nil {
		log.Fatal("error: ", err)
	}
}

type Counter int64

func (c *Counter) Write(data []byte) (n int, err error) {
	(*c) += Counter(len(data))
	return len(data), nil
}
