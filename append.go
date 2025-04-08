package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/append-processor.html.
func APPEND(dst string, val any) *AppendProc {
	p := &AppendProc{Field: dst, Value: &val}
	p.recDecl()
	p.Tag = "append_" + PathCleaner.Replace(dst)
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type AppendProc struct {
	shared[*AppendProc]

	Field           string
	Value           any
	AllowDuplicates *bool
	MediaType       *string
}

func (p *AppendProc) ALLOW_DUPLICATES(t bool) *AppendProc {
	if p.AllowDuplicates != nil {
		panic("multiple ALLOW_DUPLICATES calls")
	}
	p.AllowDuplicates = &t
	return p
}

func (p *AppendProc) MEDIA_TYPE(s string) *AppendProc {
	if p.MediaType != nil {
		panic("multiple MEDIA_TYPE calls")
	}
	p.MediaType = &s
	return p
}

func (p *AppendProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no dst for APPEND %s:%d: %s", p.file, p.line, p.Tag)
	}
	appendTemplate := template.Must(template.New("append").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- append:` +
		preamble + `
    field: {{yaml_string .Field}}
    {{yaml_value .Value}}
{{- with .MediaType}}
    media_type: {{yaml_string .}}
{{- end -}}
{{- with .AllowDuplicates}}
    allow_duplicates: {{.}}
{{- end -}}` +
		postamble,
	))
	return appendTemplate.Execute(dst, p)
}
