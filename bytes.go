package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/bytes-processor.html.
func BYTES(dst, src string) *BytesProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &BytesProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "bytes_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type BytesProc struct {
	shared[*BytesProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *BytesProc) IGNORE_MISSING(t bool) *BytesProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *BytesProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for BYTES %s:%d: %s", p.file, p.line, p.Tag)
	}
	bytesTemplate := template.Must(template.New("bytes").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- bytes:` +
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
	return bytesTemplate.Execute(dst, p)
}
