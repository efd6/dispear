package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/lowercase-processor.html.
func LOWERCASE(dst, src string) *LowercaseProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &LowercaseProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "lowercase_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type LowercaseProc struct {
	shared[*LowercaseProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *LowercaseProc) IGNORE_MISSING(t bool) *LowercaseProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *LowercaseProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for LOWERCASE %s:%d: %s", p.file, p.line, p.Tag)
	}
	lowercaseTemplate := template.Must(template.New("lowercase").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- lowercase:` +
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
	return lowercaseTemplate.Execute(dst, p)
}
