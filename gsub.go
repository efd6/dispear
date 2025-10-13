package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// GSUB adds a gsub processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/gsub-processor.html.
func GSUB(dst, src, match, replace string) *GsubProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &GsubProc{Field: src, TargetField: pDst, Pattern: match, Replacement: replace}
	p.recDecl()
	p.Tag = "gsub_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = gsubTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type GsubProc struct {
	shared[*GsubProc]

	Field         string
	TargetField   *string
	Pattern       string
	Replacement   string
	IgnoreMissing *bool
}

func (p *GsubProc) IGNORE_MISSING(t bool) *GsubProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *GsubProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for GSUB %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var gsubTemplate = template.Must(template.New("gsub").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- gsub:` +
	preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .Pattern}}
    pattern: {{yaml_string .}}
{{- end -}}
{{- with .Replacement}}
    replacement: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
	postamble,
))
