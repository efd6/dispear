package dispear

import (
	"fmt"
	"io"
	"sort"
	"text/template"
)

// REDACT adds a redact processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/redact-processor.html.
func REDACT(src string, patterns ...string) *RedactProc {
	p := &RedactProc{Field: src, Patterns: patterns}
	p.recDecl()
	p.Tag = "redact_" + PathCleaner.Replace(src)
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type RedactProc struct {
	shared[*RedactProc]

	Field              string
	Patterns           []string
	PatternDefinitions []Definition
	Prefix             *string
	Suffix             *string
	IgnoreMissing      *bool
	SkipIfUnlicensed   *bool
	TraceRedact        *bool
}

func (p *RedactProc) IGNORE_MISSING(t bool) *RedactProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *RedactProc) PREFIX(s string) *RedactProc {
	if p.Prefix != nil {
		panic("multiple PREFIX calls")
	}
	p.Prefix = &s
	return p
}

func (p *RedactProc) SUFFIX(s string) *RedactProc {
	if p.Suffix != nil {
		panic("multiple SUFFIX calls")
	}
	p.Suffix = &s
	return p
}

func (p *RedactProc) SKIP_IF_UNLICENSED(t bool) *RedactProc {
	if p.SkipIfUnlicensed != nil {
		panic("multiple SKIP_IF_UNLICENSED calls")
	}
	p.SkipIfUnlicensed = &t
	return p
}

func (p *RedactProc) PATTERN_DEFINITIONS(m map[string]string) *RedactProc {
	if p.PatternDefinitions != nil {
		panic("multiple PATTERN_DEFINITIONS calls")
	}
	p.PatternDefinitions = make([]Definition, 0, len(m))
	for k, v := range m {
		p.PatternDefinitions = append(p.PatternDefinitions, Definition{Name: k, Pattern: v})
	}
	sort.Slice(p.PatternDefinitions, func(i, j int) bool {
		return p.PatternDefinitions[i].Name < p.PatternDefinitions[j].Name
	})
	return p
}

func (p *RedactProc) TRACE_REDACT(t bool) *RedactProc {
	if p.TraceRedact != nil {
		panic("multiple TRACE_REDACT calls")
	}
	p.TraceRedact = &t
	return p
}

func (p *RedactProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for REDACT %s:%d: %s", p.file, p.line, p.Tag)
	}
	if len(p.Patterns) == 0 {
		return fmt.Errorf("no patterns for REDACT %s:%d: %s", p.file, p.line, p.Tag)
	}
	redactTemplate := template.Must(template.New("redact").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- redact:` +
		preamble + `
    field: {{yaml_string .Field}}
{{- with .Patterns}}
    patterns:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .PatternDefinitions}}
    pattern_definitions:{{range .}}
      {{yaml_string .Name}}: {{yaml_string .Pattern}}{{end}}
{{- end -}}
{{- with .Prefix}}
    prefix: {{yaml_string .}}
{{- end -}}
{{- with .Suffix}}
    suffix: {{yaml_string .}}
{{- end -}}
{{- with .SkipIfUnlicensed}}
    skip_if_unlicensed: {{.}}
{{- end -}}
{{- with .TraceRedact}}
    trace_redact: {{.}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return redactTemplate.Execute(dst, p)
}
