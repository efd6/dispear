package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/dot-expand-processor.html.
func DOT_EXPANDER(src string) *DotExpanderProc {
	p := &DotExpanderProc{Field: src}
	p.recDecl()
	p.Tag = "dot_expander_from_" + PathCleaner.Replace(src)
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type DotExpanderProc struct {
	shared[*DotExpanderProc]

	Field    string
	Path     *string
	Override *bool
}

func (p *DotExpanderProc) PATH(s string) *DotExpanderProc {
	if p.Path != nil {
		panic("multiple PATH calls")
	}
	p.Path = &s
	return p
}

func (p *DotExpanderProc) OVERRIDE(t bool) *DotExpanderProc {
	if p.Override != nil {
		panic("multiple OVERRIDE calls")
	}
	p.Override = &t
	return p
}

func (p *DotExpanderProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for DOT_EXPANDER %s:%d: %s", p.file, p.line, p.Tag)
	}
	dropTemplate := template.Must(template.New("drop").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- dot_expander:` +
		preamble + `
    field: {{yaml_string .Field}}
{{- with .Path}}
    path: {{yaml_string .}}
{{- end -}}
{{- with .Override}}
    override: {{.}}
{{- end -}}` +
		postamble,
	))
	return dropTemplate.Execute(dst, p)
}
