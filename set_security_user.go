package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// SET_SECURITY_USER adds a set_security_user processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-node-set-security-user-processor.html.
func SET_SECURITY_USER(dst string) *SetSecurityUserProc {
	p := &SetSecurityUserProc{Field: dst}
	p.recDecl()
	p.Tag = "set_security_user_to_" + PathCleaner.Replace(dst)
	p.template = setSecurityUserTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type SetSecurityUserProc struct {
	shared[*SetSecurityUserProc]

	Field      string
	Properties []string
}

func (p *SetSecurityUserProc) Name() string { return "set_security_user" }

func (p *SetSecurityUserProc) PROPERTIES(s ...string) *SetSecurityUserProc {
	if p.Properties != nil {
		panic("multiple PROPERTIES calls")
	}
	p.Properties = s
	return p
}

func (p *SetSecurityUserProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no dst for SET_SECURITY_USER %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var setSecurityUserTemplate = template.Must(template.New("set_security_user").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
	preamble + `
    field: {{yaml_string .Field}}
{{- with .Properties}}
    properties:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}` +
	postamble,
))
