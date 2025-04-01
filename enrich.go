package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/enrich-processor.html.
func ENRICH(dst, src string) *EnrichProc {
	p := &EnrichProc{Field: src, TargetField: dst}
	p.recDecl()
	p.Tag = "enrich_" + PathCleaner.Replace(src) + "_into_" + PathCleaner.Replace(dst)
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type EnrichProc struct {
	shared[*EnrichProc]

	Field         string
	TargetField   string
	IgnoreMissing *bool
	MaxMatches    *int
	Override      *bool
	PolicyName    *string
	ShapeRelation *string
}

func (p *EnrichProc) MAX_MATCHES(n int) *EnrichProc {
	if p.MaxMatches != nil {
		panic("multiple MAX_MATCHES calls")
	}
	p.MaxMatches = &n
	return p
}

func (p *EnrichProc) IGNORE_MISSING(t bool) *EnrichProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *EnrichProc) OVERRIDE(t bool) *EnrichProc {
	if p.Override != nil {
		panic("multiple OVERRIDE calls")
	}
	p.Override = &t
	return p
}

func (p *EnrichProc) POLICY_NAME(s string) *EnrichProc {
	if p.PolicyName != nil {
		panic("multiple POLICY_NAME calls")
	}
	p.PolicyName = &s
	return p
}

func (p *EnrichProc) SHAPE_RELATION(s string) *EnrichProc {
	if p.ShapeRelation != nil {
		panic("multiple SHAPE_RELATION calls")
	}
	p.ShapeRelation = &s
	return p
}

func (p *EnrichProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for ENRICH %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.TargetField == "" {
		return fmt.Errorf("no dst for ENRICH %s:%d: %s", p.file, p.line, p.Tag)
	}
	enrichTemplate := template.Must(template.New("enrich").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- enrich:` +
		preamble + `
    field: {{yaml_string .Field}}
    target_field: {{yaml_string .TargetField}}
{{- with .MaxMatches}}
    max_matches: {{.}}
{{- end -}}
{{- with .PolicyName}}
    policy_name: {{yaml_string .}}
{{- end -}}
{{- with .ShapeRelation}}
    shape_relation: {{yaml_string .}}
{{- end -}}
{{- with .Override}}
    override: {{.}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return enrichTemplate.Execute(dst, p)
}
