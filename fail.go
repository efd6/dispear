package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// FAIL adds a fail processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/fail-processor.html.
func FAIL(message string) *FailProc {
	p := &FailProc{Message: message}
	p.recDecl()
	p.Tag = "fail"
	p.template = failTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type FailProc struct {
	shared[*FailProc]

	Message string
}

func (p *FailProc) Name() string { return "fail" }

func (p *FailProc) Render(dst io.Writer, notag bool) error {
	if p.Message == "" {
		return fmt.Errorf("no message for FAIL: %s", p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var failTemplate = template.Must(template.New("fail").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
	preamble + `
    message: {{yaml_string .Message}}` +
	postamble,
))
