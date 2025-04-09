package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/script-processor.html.
func SCRIPT() *ScriptProc {
	p := &ScriptProc{}
	p.recDecl()
	p.Tag = "script"
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type ScriptProc struct {
	shared[*ScriptProc]

	ScriptID *string
	Source   *string
	Language *string
	Params   map[string]any
}

func (p *ScriptProc) ID(s string) *ScriptProc {
	if p.ScriptID != nil {
		panic("multiple ID calls")
	}
	p.ScriptID = &s
	return p
}

func (p *ScriptProc) SOURCE(s string) *ScriptProc {
	if p.Source != nil {
		panic("multiple SOURCE calls")
	}
	p.Source = &s
	return p
}

func (p *ScriptProc) PARAMS(m map[string]any) *ScriptProc {
	if p.Params != nil {
		panic("multiple PARAMS calls")
	}
	p.Params = m
	return p
}

func (p *ScriptProc) LANG(s string) *ScriptProc {
	if p.Language != nil {
		panic("multiple LANG calls")
	}
	p.Language = &s
	return p
}

func (p *ScriptProc) Render(dst io.Writer) error {
	if (p.ScriptID == nil) == (p.Source == nil) {
		return fmt.Errorf("must have one of id or source for SCRIPT %s:%d: %s", p.file, p.line, p.Tag)
	}
	scriptTemplate := template.Must(template.New("lowercase").Funcs(templateHelpers).Funcs(template.FuncMap{
		"indent": indentScript,
	}).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- script:` +
		preamble + `
{{- with .Language}}
    lang: {{yaml_string .}}
{{- end -}}
{{- with .Params}}
{{yaml 4 2 "params" .}}
{{- end -}}
{{- with .ScriptID}}
    id: {{yaml_string .}}
{{- end -}}
{{- with .Source}}
    source: |-
{{indent 6 .}}
{{- end -}}` +
		postamble,
	))
	return scriptTemplate.Execute(dst, p)
}
