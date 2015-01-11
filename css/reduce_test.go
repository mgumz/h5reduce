// copyright 2015 by mathias gumz. all rights reserved. see the LICENSE
// file for more information.

package css

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCssMin(t *testing.T) {

	var (
		err  error
		buf  bytes.Buffer
		test = []struct {
			descr         string
			before, after string
			config        []Config
		}{
			{"word", "body", "body", []Config{}},
			{"word-with-double", "tree", "tree", []Config{}},
			{"comment", "/*! comment */", "", []Config{}},
			{"comment-exclm", "/*! comment */", "/*! comment */", []Config{KeepExclamationComments}},
			{"comment-exclm2",
				"a{width:0} /*! comment */ a{width:0}",
				"a{width:0}/*! comment */a{width:0}",
				[]Config{KeepExclamationComments}},
			{"simple rule", "body { font-size: 23px; }", "body{font-size:23px}", []Config{}},
			{"multiline rule", `body {
                font-size: 12px;
                background: none;
             }`, "body{font-size:12px;background:none}", []Config{}},
		}

		// TODO: test failures
	)

	for i := range test {
		t.Logf("%s:\n%q => %q", test[i].descr, test[i].before, test[i].after)
		buf.Reset()
		err = Reduce(&buf, bytes.NewReader([]byte(test[i].before)), test[i].config...)
		if err == nil {
			if test[i].after != buf.String() {
				err = fmt.Errorf("expected:\n%q,got:\n%q", test[i].after, buf.String())
			}
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}
