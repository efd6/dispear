package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// UPPERCASE adds an uppercase processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/uppercase-processor.html.
func UPPERCASE(dst, src string) *UppercaseProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &UppercaseProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "uppercase_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = uppercaseTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type UppercaseProc struct {
	shared[*UppercaseProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *UppercaseProc) Name() string { return "uppercase" }

func (p *UppercaseProc) IGNORE_MISSING(t bool) *UppercaseProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *UppercaseProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for UPPERCASE %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var uppercaseTemplate = template.Must(template.New("uppercase").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
	preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
	postamble,
))
