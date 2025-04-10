package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// HTML_STRIP adds an html_strip processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/htmlstrip-processor.html.
func HTML_STRIP(dst, src string) *HTMLStripProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &HTMLStripProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "html_strip_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type HTMLStripProc struct {
	shared[*HTMLStripProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *HTMLStripProc) IGNORE_MISSING(t bool) *HTMLStripProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *HTMLStripProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for HTML_STRIP %s:%d: %s", p.file, p.line, p.Tag)
	}
	htmlStripTemplate := template.Must(template.New("html_strip").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- html_strip:` +
		preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return htmlStripTemplate.Execute(dst, p)
}
