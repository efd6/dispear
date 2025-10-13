package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// TRIM adds a trim processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/trim-processor.html.
func TRIM(dst, src string) *TrimProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &TrimProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "trim_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = trimTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type TrimProc struct {
	shared[*TrimProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *TrimProc) Name() string { return "trim" }

func (p *TrimProc) IGNORE_MISSING(t bool) *TrimProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *TrimProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for TRIM %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var trimTemplate = template.Must(template.New("trim").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
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
