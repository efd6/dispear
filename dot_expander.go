package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// DOT_EXPANDER adds a dot_expander processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/dot-expand-processor.html.
func DOT_EXPANDER(src string) *DotExpanderProc {
	p := &DotExpanderProc{Field: src}
	p.recDecl()
	p.Tag = "dot_expander_from_" + PathCleaner.Replace(src)
	p.template = dotExpanderTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type DotExpanderProc struct {
	shared[*DotExpanderProc]

	Field    string
	Path     *string
	Override *bool
}

func (p *DotExpanderProc) Name() string { return "dot_expander" }

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

func (p *DotExpanderProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for DOT_EXPANDER %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var dotExpanderTemplate = template.Must(template.New("dot_expander").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
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
