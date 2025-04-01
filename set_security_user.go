package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-node-set-security-user-processor.html.
func SET_SECURITY_USER(dst string) *SetSecurityUserProc {
	p := &SetSecurityUserProc{Field: dst}
	p.recDecl()
	p.Tag = "set_security_user_to_" + PathCleaner.Replace(dst)
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type SetSecurityUserProc struct {
	shared[*SetSecurityUserProc]

	Field      string
	Properties []string
}

func (p *SetSecurityUserProc) PROPERTIES(s ...string) *SetSecurityUserProc {
	if p.Properties != nil {
		panic("multiple PROPERTIES calls")
	}
	p.Properties = s
	return p
}

func (p *SetSecurityUserProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no dst for SET_SECURITY_USER %s:%d: %s", p.file, p.line, p.Tag)
	}
	ipLocationTemplate := template.Must(template.New("set_security_user").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- set_security_user:` +
		preamble + `
    field: {{yaml_string .Field}}
{{- with .Properties}}
    properties:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}` +
		postamble,
	))
	return ipLocationTemplate.Execute(dst, p)
}
