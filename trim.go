package dispear

import (
	"fmt"
	"io"
	"text/template"
)

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
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type TrimProc struct {
	shared[*TrimProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *TrimProc) IGNORE_MISSING(t bool) *TrimProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *TrimProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for TRIM %s:%d: %s", p.file, p.line, p.Tag)
	}
	trimTemplate := template.Must(template.New("trim").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- trim:` +
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
	return trimTemplate.Execute(dst, p)
}
