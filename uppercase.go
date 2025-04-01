package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/uppercase-processor.html.
func UPPERCASE(dst, src string) *UppercaseProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &UppercaseProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "uppercase_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type UppercaseProc struct {
	shared[*UppercaseProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *UppercaseProc) IGNORE_MISSING(t bool) *UppercaseProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *UppercaseProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for UPPERCASE %s:%d: %s", p.file, p.line, p.Tag)
	}
	uppercaseTemplate := template.Must(template.New("uppercase").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- uppercase:` +
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
	return uppercaseTemplate.Execute(dst, p)
}
