package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/kv-processor.html.
func KV(dst, src, fieldsplit, valuesplit string) *KVProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &KVProc{Field: src, TargetField: pDst, FieldSplit: fieldsplit, ValueSplit: valuesplit}
	p.recDecl()
	p.Tag = "kv_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = kvTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type KVProc struct {
	shared[*KVProc]

	Field         string
	TargetField   *string
	FieldSplit    string
	ValueSplit    string
	IgnoreMissing *bool
	ExcludeKeys   *bool
	IncludeKeys   *bool
	Prefix        *string
	StripBrackets *bool
	TrimKey       *bool
	TrimValue     *bool
}

func (p *KVProc) IGNORE_MISSING(t bool) *KVProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *KVProc) EXCLUDE_KEYS(t bool) *KVProc {
	if p.ExcludeKeys != nil {
		panic("multiple EXCLUDE_KEYS calls")
	}
	p.ExcludeKeys = &t
	return p
}

func (p *KVProc) INCLUDE_KEYS(t bool) *KVProc {
	if p.IncludeKeys != nil {
		panic("multiple INCLUDE_KEYS calls")
	}
	p.IncludeKeys = &t
	return p
}

func (p *KVProc) PREFIX(s string) *KVProc {
	if p.Prefix != nil {
		panic("multiple PREFIX calls")
	}
	p.Prefix = &s
	return p
}

func (p *KVProc) STRIP_BRACKETS(t bool) *KVProc {
	if p.StripBrackets != nil {
		panic("multiple STRIP_BRACKETS calls")
	}
	p.StripBrackets = &t
	return p
}

func (p *KVProc) TRIM_KEY(t bool) *KVProc {
	if p.TrimKey != nil {
		panic("multiple TRIM_KEY calls")
	}
	p.TrimKey = &t
	return p
}

func (p *KVProc) TRIM_VALUE(t bool) *KVProc {
	if p.TrimValue != nil {
		panic("multiple TRIM_VALUE calls")
	}
	p.TrimValue = &t
	return p
}

func (p *KVProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for KV %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.FieldSplit == "" {
		return fmt.Errorf("no field split for KV %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.ValueSplit == "" {
		return fmt.Errorf("no value split for KV %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var kvTemplate = template.Must(template.New("kv").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- kv:` +
	preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .FieldSplit}}
    field_split: {{yaml_string .}}
{{- end -}}
{{- with .ValueSplit}}
    value_split: {{yaml_string .}}
{{- end -}}
{{- with .ExcludeKeys}}
    exclude_keys: {{.}}
{{- end -}}
{{- with .IncludeKeys}}
    include_keys: {{.}}
{{- end -}}
{{- with .Prefix}}
    prefix: {{yaml_string .}}
{{- end -}}
{{- with .StripBrackets}}
    strip_brackets: {{.}}
{{- end -}}
{{- with .TrimKey}}
    trim_key: {{.}}
{{- end -}}
{{- with .TrimValue}}
    trim_value: {{.}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
	postamble,
))
