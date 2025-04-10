package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// CSV adds a csv processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/csv-processor.html.
func CSV(dst, src string) *CSVProc {
	p := &CSVProc{Field: src, TargetField: dst}
	p.recDecl()
	p.Tag = "csv_" + PathCleaner.Replace(src) + "_into_" + PathCleaner.Replace(dst)
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type CSVProc struct {
	shared[*CSVProc]

	Field         string
	TargetField   string
	EmptyValue    *string
	Quote         *string
	Separator     *string
	Trim          *bool
	IgnoreMissing *bool
}

func (p *CSVProc) EMPTY_VALUE(s string) *CSVProc {
	if p.EmptyValue != nil {
		panic("multiple EMPTY_VALUE calls")
	}
	p.EmptyValue = &s
	return p
}

func (p *CSVProc) QUOTE(s string) *CSVProc {
	if p.Quote != nil {
		panic("multiple QUOTE calls")
	}
	p.Quote = &s
	return p
}

func (p *CSVProc) SEPARATOR(s string) *CSVProc {
	if p.Separator != nil {
		panic("multiple SEPARATOR calls")
	}
	p.Separator = &s
	return p
}

func (p *CSVProc) TRIM(t bool) *CSVProc {
	if p.Trim != nil {
		panic("multiple TRIM calls")
	}
	p.Trim = &t
	return p
}

func (p *CSVProc) IGNORE_MISSING(t bool) *CSVProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *CSVProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for CSV %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.TargetField == "" {
		return fmt.Errorf("no dst for CSV %s:%d: %s", p.file, p.line, p.Tag)
	}
	csvTemplate := template.Must(template.New("csv").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- csv:` +
		preamble + `
    field: {{yaml_string .Field}}
    target_field: {{yaml_string .TargetField}}
{{- with .Quote}}
    quote: {{yaml_string .}}
{{- end -}}
{{- with .Separator}}
    separator: {{yaml_string .}}
{{- end -}}
{{- with .Trim}}
    trim: {{.}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return csvTemplate.Execute(dst, p)
}
