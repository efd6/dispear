package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/uri-parts-processor.html.
func URI_PARTS(dst, src string) *URIPartsProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &URIPartsProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "uri_parts_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type URIPartsProc struct {
	shared[*URIPartsProc]

	Field              string
	TargetField        *string
	KeepOriginal       *bool
	RemoveIfSuccessful *bool
	IgnoreMissing      *bool
}

func (p *URIPartsProc) KEEP_ORIGINAL(t bool) *URIPartsProc {
	if p.KeepOriginal != nil {
		panic("multiple KEEP_ORIGINAL calls")
	}
	p.KeepOriginal = &t
	return p
}

func (p *URIPartsProc) REMOVE_IF_SUCCESSFUL(t bool) *URIPartsProc {
	if p.RemoveIfSuccessful != nil {
		panic("multiple REMOVE_IF_SUCCESSFUL calls")
	}
	p.RemoveIfSuccessful = &t
	return p
}

func (p *URIPartsProc) IGNORE_MISSING(t bool) *URIPartsProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *URIPartsProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for URI_PARTS %s:%d: %s", p.file, p.line, p.Tag)
	}
	uriPartsTemplate := template.Must(template.New("uri_parts").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- uri_parts:` +
		preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .KeepOriginal}}
    keep_original: {{.}}
{{- end -}}
{{- with .RemoveIfSuccessful}}
    remove_if_successful: {{.}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return uriPartsTemplate.Execute(dst, p)
}
