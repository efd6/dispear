package dispear

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/pipeline-processor.html.
func PIPELINE(name string) *PipelineProc {
	p := &PipelineProc{Name: name}
	p.recDecl()
	p.Tag = "pipeline_" + PathCleaner.Replace(name)
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type PipelineProc struct {
	shared[*PipelineProc]

	Name          string
	IgnoreMissing *bool
}

func (p *PipelineProc) IGNORE_MISSING(t bool) *PipelineProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *PipelineProc) Render(dst io.Writer) error {
	if p.Name == "" {
		return fmt.Errorf("no name for PIPELINE %s:%d: %s", p.file, p.line, p.Tag)
	}
	pipelineTemplate := template.Must(template.New("pipeline").Funcs(template.FuncMap{
		"comment": func(s string) string {
			return "# " + strings.Join(strings.Split(s, "\n"), "\n# ")
		},
		"yaml_string": yamlString,
		"render": func(r Renderer) (string, error) {
			var buf bytes.Buffer
			err := r.Render(&buf)
			if err != nil {
				return "", err
			}
			return indent(strings.TrimSpace(buf.String()), 6), nil
		},
		"left_braces":  func() string { return "{{" },
		"right_braces": func() string { return "}}" },
	}).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- pipeline:` +
		preamble + `
    name: '{{left_braces}} IngestPipeline "{{.Name}}" {{right_braces}}'
{{- with .IgnoreMissing}}
    ignore_missing_pipeline: {{.}}
{{- end -}}` +
		postamble,
	))
	return pipelineTemplate.Execute(dst, p)
}
