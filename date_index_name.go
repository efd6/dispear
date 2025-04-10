package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// DATE_INDEX_NAME adds a date_index_name processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/date-index-name-processor.html.
func DATE_INDEX_NAME(src, rounding string) *DateIndexNameProc {
	p := &DateIndexNameProc{Field: src, Rounding: rounding}
	p.recDecl()
	p.Tag = "date_index_name_" + PathCleaner.Replace(src) + "_round_to_" + rounding
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type DateIndexNameProc struct {
	shared[*DateIndexNameProc]

	Field           string
	Rounding        string
	Formats         []string
	IndexNameFormat *string
	IndexNamePrefix *string
	Locale          *string
	Timezone        *string
	Separator       *string
}

func (p *DateIndexNameProc) DATE_FORMATS(s ...string) *DateIndexNameProc {
	if p.Formats != nil {
		panic("multiple DATE_FORMATS calls")
	}
	p.Formats = s
	return p
}

func (p *DateIndexNameProc) INDEX_NAME_FORMAT(s string) *DateIndexNameProc {
	if p.IndexNameFormat != nil {
		panic("multiple INDEX_NAME_FORMAT calls")
	}
	p.IndexNameFormat = &s
	return p
}

func (p *DateIndexNameProc) INDEX_NAME_PREFIX(s string) *DateIndexNameProc {
	if p.IndexNamePrefix != nil {
		panic("multiple INDEX_NAME_PREFIX calls")
	}
	p.IndexNamePrefix = &s
	return p
}

func (p *DateIndexNameProc) LOCALE(s string) *DateIndexNameProc {
	if p.Locale != nil {
		panic("multiple LOCALE calls")
	}
	p.Locale = &s
	return p
}

func (p *DateIndexNameProc) TIMEZONE(s string) *DateIndexNameProc {
	if p.Timezone != nil {
		panic("multiple TIMEZONE calls")
	}
	p.Timezone = &s
	return p
}

func (p *DateIndexNameProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for DATE_INDEX_NAME %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.Rounding == "" {
		return fmt.Errorf("no rounding for DATE_INDEX_NAME %s:%d: %s", p.file, p.line, p.Tag)
	}
	dateIndexNameTemplate := template.Must(template.New("date_index_name").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- date_index_name:` +
		preamble + `
    field: {{yaml_string .Field}}
    date_rounding: {{yaml_string .Rounding}}
{{- with .Formats}}
    date_formats:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .IndexNameFormat}}
    index_name_format: {{yaml_string .}}
{{- end -}}
{{- with .IndexNamePrefix}}
    index_name_prefix: {{yaml_string .}}
{{- end -}}
{{- with .Locale}}
    locale: {{yaml_string .}}
{{- end -}}
{{- with .Timezone}}
    timezone: {{yaml_string .}}
{{- end -}}` +
		postamble,
	))
	return dateIndexNameTemplate.Execute(dst, p)
}
