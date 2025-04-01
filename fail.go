package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/fail-processor.html.
func FAIL(message string) *FailProc {
	p := &FailProc{Message: message}
	p.recDecl()
	p.Tag = "fail"
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type FailProc struct {
	shared[*FailProc]

	Message string
}

func (p *FailProc) Render(dst io.Writer) error {
	if p.Message == "" {
		return fmt.Errorf("no message for FAIL: %s", p.Tag)
	}
	failTemplate := template.Must(template.New("fail").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- fail:` +
		preamble + `
    message: {{yaml_string .Message}}` +
		postamble,
	))
	return failTemplate.Execute(dst, p)
}
