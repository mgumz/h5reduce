package css

import (
	"bytes"
	"testing"
)

func TestCssMin(t *testing.T) {

	var (
		err  error
		buf  bytes.Buffer
		test = []struct {
			descr         string
			before, after string
			do_linebreak  bool
		}{
			{"word", "body", "body", false},
			{"comment", "/* comment */", "", false},
			{"simple rule", "body { font-size: 23px; }", "body{font-size:23px}", false},
			{"multiline rule", `body {
                font-size: 12px;
                background: none;
             }`, "body{font-size:12px;background:none}", false},
		}

		// TODO: test failures
	)

	for i := range test {
		buf.Reset()

		err = Reduce(&buf, test[i].before, test[i].do_linebreak)

		t.Log(test[i].descr, err)
		if err != nil {
			t.Fatal(err)
		}
		if test[i].after != buf.String() {
			t.Fatal(err)
		}
	}

}
