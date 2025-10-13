package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// BYTES adds a bytes processor to the global context.
//
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
	p.template = bytesTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type BytesProc struct {
	shared[*BytesProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *BytesProc) Name() string { return "bytes" }

func (p *BytesProc) IGNORE_MISSING(t bool) *BytesProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *BytesProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for BYTES %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var bytesTemplate = template.Must(template.New("bytes").Funcs(templateHelpers).Parse(`
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
