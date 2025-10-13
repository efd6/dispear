package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// DATE adds a date processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/date-processor.html.
func DATE(dst, src string, formats ...string) *DateProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &DateProc{Field: src, TargetField: pDst, Formats: formats}
	p.recDecl()
	p.Tag = "date_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = dateTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type DateProc struct {
	shared[*DateProc]

	Field        string
	Formats      []string
	TargetField  *string
	Locale       *string
	OutputFormat *string
	Timezone     *string
}

func (p *DateProc) Name() string { return "date" }

func (p *DateProc) LOCALE(s string) *DateProc {
	if p.Locale != nil {
		panic("multiple LOCALE calls")
	}
	p.Locale = &s
	return p
}

func (p *DateProc) OUTPUT_FORMAT(s string) *DateProc {
	if p.OutputFormat != nil {
		panic("multiple OUTPUT_FORMAT calls")
	}
	p.OutputFormat = &s
	return p
}

func (p *DateProc) TIMEZONE(s string) *DateProc {
	if p.Timezone != nil {
		panic("multiple TIMEZONE calls")
	}
	p.Timezone = &s
	return p
}

func (p *DateProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for DATE %s:%d: %s", p.file, p.line, p.Tag)
	}
	if len(p.Formats) == 0 {
		return fmt.Errorf("no formats for DATE %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var dateTemplate = template.Must(template.New("date").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
	preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .Formats}}
    formats:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .Locale}}
    locale: {{yaml_string .}}
{{- end -}}
{{- with .Timezone}}
    timezone: {{yaml_string .}}
{{- end -}}
{{- with .OutputFormat}}
    output_format: {{yaml_string .}}
{{- end -}}` +
	postamble,
))
