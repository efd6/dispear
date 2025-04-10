package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// REGISTERED_DOMAIN adds a registered_domain processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/registered-domain-processor.html.
func REGISTERED_DOMAIN(dst, src string) *RegisteredDomainProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &RegisteredDomainProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "registered_domain_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type RegisteredDomainProc struct {
	shared[*RegisteredDomainProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
}

func (p *RegisteredDomainProc) IGNORE_MISSING(t bool) *RegisteredDomainProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *RegisteredDomainProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for REGISTERED_DOMAIN %s:%d: %s", p.file, p.line, p.Tag)
	}
	registeredDomainTemplate := template.Must(template.New("registered_domain").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- registered_domain:` +
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
	return registeredDomainTemplate.Execute(dst, p)
}
