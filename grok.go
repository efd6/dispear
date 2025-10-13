package dispear

import (
	"fmt"
	"io"
	"sort"
	"text/template"
)

// GROK adds a grok processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/grok-processor.html.
func GROK(src string, patterns ...string) *GrokProc {
	p := &GrokProc{Field: src, Patterns: patterns}
	p.recDecl()
	p.Tag = "grok_" + PathCleaner.Replace(src)
	p.template = grokTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type GrokProc struct {
	shared[*GrokProc]

	Field              string
	Patterns           []string
	PatternDefinitions []Definition
	IgnoreMissing      *bool
	ECSCompatibility   *string
	TraceMatch         *bool
}

func (p *GrokProc) IGNORE_MISSING(t bool) *GrokProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *GrokProc) ECS_COMPATIBILITY(s string) *GrokProc {
	if p.ECSCompatibility != nil {
		panic("multiple ECS_COMPATIBILITY calls")
	}
	p.ECSCompatibility = &s
	return p
}

func (p *GrokProc) PATTERN_DEFINITIONS(m map[string]string) *GrokProc {
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

func (p *GrokProc) TRACE_MATCH(t bool) *GrokProc {
	if p.TraceMatch != nil {
		panic("multiple TRACE_MATCH calls")
	}
	p.TraceMatch = &t
	return p
}

func (p *GrokProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for GROK %s:%d: %s", p.file, p.line, p.Tag)
	}
	if len(p.Patterns) == 0 {
		return fmt.Errorf("no patterns for GROK %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var grokTemplate = template.Must(template.New("grok").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- grok:` +
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
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
	postamble,
))
