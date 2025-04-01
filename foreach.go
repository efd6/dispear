package dispear

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/foreach-processor.html.
func FOREACH(src string, proc Renderer) *ForeachProc {
	ctx.processors = slices.DeleteFunc(ctx.processors, func(e Renderer) bool {
		return proc == e
	})
	p := &ForeachProc{Field: src, Processor: proc}
	p.recDecl()
	p.Tag = "foreach"
	if src != "" {
		p.Tag += "_of_" + PathCleaner.Replace(src)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type ForeachProc struct {
	shared[*ForeachProc]

	Field         string
	Processor     Renderer
	IgnoreMissing *bool
}

func (p *ForeachProc) IGNORE_MISSING(t bool) *ForeachProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *ForeachProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for FOREACH %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.Processor == nil {
		return fmt.Errorf("no processor for FOREACH %s:%d: %s", p.file, p.line, p.Tag)
	}
	foreachTemplate := template.Must(template.New("foreach").Funcs(templateHelpers).Funcs(template.FuncMap{
		// render is overridden due to foreach only taking a single processor.
		"render": func(r Renderer) (string, error) {
			var buf bytes.Buffer
			err := r.Render(&buf)
			if err != nil {
				return "", err
			}
			// Trim the list marker from the processor and then reindent.
			return indent(dedent(strings.TrimSpace(buf.String()), 2), 6), nil
		},
	}).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- foreach:` +
		preamble + `
    field: {{yaml_string .Field}}
    processor:
{{render .Processor}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return foreachTemplate.Execute(dst, p)
}
