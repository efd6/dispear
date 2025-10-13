package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// DISSECT adds a dissect processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/dissect-processor.html.
func DISSECT(src, pattern string) *DissectProc {
	p := &DissectProc{Field: src, Pattern: pattern}
	p.recDecl()
	p.Tag = "dissect_" + PathCleaner.Replace(src)
	p.template = dissectTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type DissectProc struct {
	shared[*DissectProc]

	Field           string
	Pattern         string
	AppendSeparator *string
	IgnoreMissing   *bool
}

func (p *DissectProc) Name() string { return "dissect" }

func (p *DissectProc) APPEND_SEPARATOR(s string) *DissectProc {
	if p.AppendSeparator != nil {
		panic("multiple APPEND_SEPARATOR calls")
	}
	p.AppendSeparator = &s
	return p
}

func (p *DissectProc) IGNORE_MISSING(t bool) *DissectProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *DissectProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for ATTACHMENT %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.Pattern == "" {
		return fmt.Errorf("no pattern for DISSECT %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var dissectTemplate = template.Must(template.New("dissect").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- dissect:` +
	preamble + `
    field: {{yaml_string .Field}}
    pattern: {{yaml_string .Pattern}}
{{- with .AppendSeparator}}
    append_separator: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
	postamble,
))
