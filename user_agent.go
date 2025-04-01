package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/user-agent-processor.html.
func USER_AGENT(dst, src string) *UserAgentProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &UserAgentProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "user_agent_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type UserAgentProc struct {
	shared[*UserAgentProc]

	Field             string
	TargetField       *string
	IgnoreMissing     *bool
	ExtractDeviceType *bool
	RegexFile         *string
	Properties        []string
}

func (p *UserAgentProc) IGNORE_MISSING(t bool) *UserAgentProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *UserAgentProc) REGEX_FILE(s string) *UserAgentProc {
	if p.RegexFile != nil {
		panic("multiple REGEX_FILE calls")
	}
	p.RegexFile = &s
	return p
}

func (p *UserAgentProc) EXTRACT_DEVICE_TYPE(t bool) *UserAgentProc {
	if p.ExtractDeviceType != nil {
		panic("multiple EXTRACT_DEVICE_TYPE calls")
	}
	p.ExtractDeviceType = &t
	return p
}

func (p *UserAgentProc) PROPERTIES(s ...string) *UserAgentProc {
	if p.Properties != nil {
		panic("multiple PROPERTIES calls")
	}
	p.Properties = s
	return p
}

func (p *UserAgentProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for USER_AGENT %s:%d: %s", p.file, p.line, p.Tag)
	}
	userAgentTemplate := template.Must(template.New("user_agent").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- user_agent:` +
		preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .ExtractDeviceType}}
    extract_device_type: {{.}}
{{- end -}}
{{- with .RegexFile}}
    regex_file: {{yaml_string .}}
{{- end -}}
{{- with .Properties}}
    properties:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return userAgentTemplate.Execute(dst, p)
}
