package dispear

import (
	"fmt"
	"io"
	"slices"
	"text/template"
)

// REMOVE adds a remove processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/remove-processor.html.
func REMOVE(fields ...string) *RemoveProc {
	p := &RemoveProc{Fields: fields}
	p.recDecl()
	p.Tag = "remove"
	p.template = removeTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type RemoveProc struct {
	shared[*RemoveProc]

	Fields        []string
	Keep          []string
	IgnoreMissing *bool
}

func (p *RemoveProc) KEEP(fields ...string) *RemoveProc {
	if p.Keep != nil {
		panic("multiple KEEP calls")
	}
	p.Keep = fields
	return p
}

func (p *RemoveProc) IGNORE_MISSING(t bool) *RemoveProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *RemoveProc) Render(dst io.Writer, notag bool) error {
	if len(p.Fields) == 0 && len(p.Keep) == 0 {
		return fmt.Errorf("no field or keep for REMOVE %s:%d: %s", p.file, p.line, p.Tag)
	}
	if slices.Contains(p.Fields, "") {
		return fmt.Errorf("empty field element for REMOVE %s:%d: %s", p.file, p.line, p.Tag)
	}
	if slices.Contains(p.Keep, "") {
		return fmt.Errorf("empty keep element for REMOVE %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var removeTemplate = template.Must(template.New("remove").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- remove:` +
	preamble + `
{{- with .Fields}}
    field:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .Keep}}
    keep:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
	postamble,
))
