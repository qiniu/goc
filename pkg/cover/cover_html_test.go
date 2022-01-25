package cover

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCoverHtmlTmpl(t *testing.T) {
	bf := &bytes.Buffer{}
	coverHtmlTemplate := &CoverHtmlTemplate{
		HtmlTemplate: fmt.Sprintf("`%s`", tmplHTML),
	}
	if err := coverHtmlTmpl.Execute(bf, coverHtmlTemplate); err != nil {
		return
	}
	CoverHtmlFuncGocUtils("./testcoverhtml.go", bf.String())
}
