package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// CIRCLE adds a circle processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-circle-processor.html.
func CIRCLE(dst, src, typ string, err float64) *CircleProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	switch typ {
	case "geo_shape", "shape":
	default:
		panic("invalid shape type: " + typ)
	}
	p := &CircleProc{Field: src, TargetField: pDst, ShapeType: typ, ErrorDistance: err}
	p.recDecl()
	p.Tag = "circle_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = circleTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type CircleProc struct {
	shared[*CircleProc]

	Field         string
	TargetField   *string
	IgnoreMissing *bool
	ShapeType     string
	ErrorDistance float64
}

func (p *CircleProc) IGNORE_MISSING(t bool) *CircleProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *CircleProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for CIRCLE %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var circleTemplate = template.Must(template.New("circle").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- circle:` +
	preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}
{{/**/}}
    shape_type: {{.ShapeType}}
    error_distance: {{.ErrorDistance}}` +
	postamble,
))
