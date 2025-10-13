package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// LOWERCASE adds a lowercase processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/lowercase-processor.html.
func LOWERCASE(dst, src string) *LowercaseProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &LowercaseProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "lowercase_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = lowercaseTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type LowercaseProc struct {
	shared[*LowercaseProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *LowercaseProc) IGNORE_MISSING(t bool) *LowercaseProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *LowercaseProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for LOWERCASE %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var lowercaseTemplate = template.Must(template.New("lowercase").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- lowercase:` +
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
