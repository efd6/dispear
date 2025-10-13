package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// SET adds a set processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/set-processor.html.
func SET(dst string) *SetProc {
	p := &SetProc{Field: dst}
	p.recDecl()
	p.Tag = "set_" + PathCleaner.Replace(dst)
	p.template = setTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type SetProc struct {
	shared[*SetProc]

	Field       string
	Value       any
	CopyFrom    *string
	Override    *bool
	MediaType   *string
	IgnoreEmpty *bool
}

func (p *SetProc) IGNORE_EMPTY(t bool) *SetProc {
	if p.IgnoreEmpty != nil {
		panic("multiple IGNORE_EMPTY calls")
	}
	p.IgnoreEmpty = &t
	return p
}

func (p *SetProc) VALUE(v any) *SetProc {
	if p.Value != nil {
		panic("multiple VALUE calls")
	}
	p.Value = &v
	return p
}

func (p *SetProc) COPY_FROM(s string) *SetProc {
	if p.CopyFrom != nil {
		panic("multiple COPY_FROM calls")
	}
	p.CopyFrom = &s
	return p
}

func (p *SetProc) OVERRIDE(t bool) *SetProc {
	if p.Override != nil {
		panic("multiple OVERRIDE calls")
	}
	p.Override = &t
	return p
}

func (p *SetProc) MEDIA_TYPE(s string) *SetProc {
	if p.MediaType != nil {
		panic("multiple MEDIA_TYPE calls")
	}
	p.MediaType = &s
	return p
}

func (p *SetProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no dst for SET %s:%d: %s", p.file, p.line, p.Tag)
	}
	if (p.Value == nil) == (p.CopyFrom == nil) {
		return fmt.Errorf("must have one of value or copy from for SET %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var setTemplate = template.Must(template.New("set").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- set:` +
	preamble + `
    field: {{yaml_string .Field}}
{{- with .Value}}
{{yaml 4 2 "value" .}}
{{- end -}}
{{- with .CopyFrom}}
    copy_from: {{yaml_string .}}
{{- end -}}
{{- with .Override}}
    override: {{.}}
{{- end -}}
{{- with .MediaType}}
    media_type: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreEmpty}}
    ignore_empty_value: {{.}}
{{- end -}}` +
	postamble,
))
