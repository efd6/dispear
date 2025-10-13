package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// JOIN adds a join processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/join-processor.html.
func JOIN(dst, src, sep string) *JoinProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &JoinProc{Field: src, TargetField: pDst, Separator: sep}
	p.recDecl()
	p.Tag = "join_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = joinTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type JoinProc struct {
	shared[*JoinProc]

	Field       string
	Separator   string
	TargetField *string
}

func (p *JoinProc) Name() string { return "join" }

func (p *JoinProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for JOIN %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.Separator == "" {
		return fmt.Errorf("no separator for JOIN %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var joinTemplate = template.Must(template.New("join").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
	preamble + `
    field: {{yaml_string .Field}}
    separator: {{yaml_string .Separator}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}` +
	postamble,
))
