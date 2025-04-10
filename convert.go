package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// CONVERT adds a convert processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/convert-processor.html.
func CONVERT(dst, src, typ string) *ConvertProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &ConvertProc{Field: src, Type: typ, TargetField: pDst}
	p.recDecl()
	p.Tag = "convert_" + PathCleaner.Replace(src)
	if typ != "" {
		p.Tag += "_to_" + typ
	}
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type ConvertProc struct {
	shared[*ConvertProc]

	Field         string
	Type          string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *ConvertProc) IGNORE_MISSING(t bool) *ConvertProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *ConvertProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for CONVERT %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.Type == "" {
		return fmt.Errorf("no type for CONVERT %s:%d: %s", p.file, p.line, p.Tag)
	}
	convertTemplate := template.Must(template.New("convert").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- convert:` +
		preamble + `
    field: {{yaml_string .Field}}
    type: {{yaml_string .Type}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return convertTemplate.Execute(dst, p)
}
