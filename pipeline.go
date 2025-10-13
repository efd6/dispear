package dispear

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
)

// PIPELINE adds a pipeline processor to the global context. Only the name of
// the target pipeline is required. The template will be constructed by dispear.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/pipeline-processor.html.
func PIPELINE(name string) *PipelineProc {
	p := &PipelineProc{Name: name}
	p.recDecl()
	p.Tag = "pipeline_" + PathCleaner.Replace(name)
	p.template = pipelineTemplate
	p.parent = p
	ctx.Add(p)
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

func (p *PipelineProc) Render(dst io.Writer, notag bool) error {
	if p.Name == "" {
		return fmt.Errorf("no name for PIPELINE %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var pipelineTemplate = template.Must(template.New("pipeline").Funcs(template.FuncMap{
	"comment": func(s string) string {
		return "# " + strings.Join(strings.Split(s, "\n"), "\n# ")
	},
	"gutter":      gutter,
	"yaml":        yamlValue,
	"yaml_string": yamlString,
	"render": func(r Renderer, notag bool) (string, error) {
		var buf bytes.Buffer
		err := r.Render(&buf, notag)
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
