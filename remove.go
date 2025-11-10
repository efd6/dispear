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
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
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

func (p *RemoveProc) Render(dst io.Writer) error {
	if len(p.Fields) == 0 && len(p.Keep) == 0 {
		return fmt.Errorf("no field or keep for REMOVE %s:%d: %s", p.file, p.line, p.Tag)
	}
	if slices.Contains(p.Fields, "") {
		return fmt.Errorf("empty field element for REMOVE %s:%d: %s", p.file, p.line, p.Tag)
	}
	if slices.Contains(p.Keep, "") {
		return fmt.Errorf("empty keep element for REMOVE %s:%d: %s", p.file, p.line, p.Tag)
	}
	removeTemplate := template.Must(template.New("remove").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- remove:` +
		preamble + `
{{- if eq (len .Fields) 1}}
    field: {{index .Fields 0}}
{{- else if gt (len .Fields) 1}}
    field:{{range .Fields}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- if eq (len .Keep) 1}}
    keep: {{index .Keep 0}}
{{- else if gt (len .Keep) 1}}
    keep:{{range .Keep}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return removeTemplate.Execute(dst, p)
}
