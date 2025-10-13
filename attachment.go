package dispear

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

// ATTACHMENT adds an attachment processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/attachment.html.
func ATTACHMENT(dst, src string) *AttachmentProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &AttachmentProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "attachment_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = attachmentTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type AttachmentProc struct {
	shared[*AttachmentProc]

	Field             string
	TargetField       *string
	IgnoreMissing     *bool
	IndexedChars      *int
	IndexedCharsField *string
	Properties        []string
	RemoveBinary      *bool
	ResourceName      *string
}

func (p *AttachmentProc) Name() string { return "attachment" }

func (p *AttachmentProc) IGNORE_MISSING(t bool) *AttachmentProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *AttachmentProc) INDEXED_CHARS(i int) *AttachmentProc {
	if p.IndexedChars != nil {
		panic("multiple INDEXED_CHARS calls")
	}
	p.IndexedChars = &i
	return p
}

func (p *AttachmentProc) INDEXED_CHARS_FIELD(s string) *AttachmentProc {
	if p.IndexedCharsField != nil {
		panic("multiple INDEXED_CHARS_FIELD calls")
	}
	p.IndexedCharsField = &s
	return p
}

func (p *AttachmentProc) PROPERTIES(s ...string) *AttachmentProc {
	if p.Properties != nil {
		panic("multiple PROPERTIES calls")
	}
	var invalid []string
	for _, v := range s {
		switch v {
		case "content",
			"title",
			"name",
			"author",
			"keywords",
			"date",
			"content_type",
			"content_length",
			"language":
		default:
			invalid = append(invalid, v)
		}
	}
	if invalid != nil {
		if len(invalid) == 1 {
			panic("invalid attachment property: " + invalid[0])
		}
		panic("invalid attachment properties: " + strings.Join(invalid, ", "))
	}
	p.Properties = s
	return p
}

func (p *AttachmentProc) REMOVE_BINARY(t bool) *AttachmentProc {
	if p.RemoveBinary != nil {
		panic("multiple REMOVE_BINARY calls")
	}
	p.RemoveBinary = &t
	return p
}

func (p *AttachmentProc) RESOURCE_NAME(s string) *AttachmentProc {
	if p.ResourceName != nil {
		panic("multiple RESOURCE_NAME calls")
	}
	p.ResourceName = &s
	return p
}

func (p *AttachmentProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for ATTACHMENT %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var attachmentTemplate = template.Must(template.New("attachment").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
	preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}
{{- with .IndexedChars}}
    indexed_chars: {{.}}
{{- end -}}
{{- with .IndexedCharsField}}
    indexed_chars_field: {{yaml_string .}}
{{- end -}}
{{- with .Properties}}
    properties:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .RemoveBinary}}
    remove_binary: {{.}}
{{- end -}}
{{- with .ResourceName}}
    resource_name: {{yaml_string .}}
{{- end -}}` +
	postamble,
))
