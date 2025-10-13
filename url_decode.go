package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// URL_DECODE adds a url_decode processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/urldecode-processor.html.
func URL_DECODE(dst, src string) *URLDecodeProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &URLDecodeProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "urldecode_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = urlDecodeTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type URLDecodeProc struct {
	shared[*URLDecodeProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *URLDecodeProc) IGNORE_MISSING(t bool) *URLDecodeProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *URLDecodeProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for URL_DECODE %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var urlDecodeTemplate = template.Must(template.New("url_decode").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- urldecode:` +
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
