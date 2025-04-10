package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// SORT adds a sort processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/sort-processor.html.
func SORT(dst, src, order string) *SortProc {
	var pDst, pOrder *string
	if dst != "" {
		pDst = &dst
	}
	if order != "" {
		pOrder = &order
	}
	p := &SortProc{Field: src, TargetField: pDst, Order: pOrder}
	p.recDecl()
	p.Tag = "sort_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	if order != "" {
		p.Tag += "_" + PathCleaner.Replace(order)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type SortProc struct {
	shared[*SortProc]

	Field       string
	TargetField *string
	Order       *string
}

func (p *SortProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for SORT %s:%d: %s", p.file, p.line, p.Tag)
	}
	sortTemplate := template.Must(template.New("sort").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- sort:` +
		preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .Order}}
    order: {{yaml_string .}}
{{- end -}}` +
		postamble,
	))
	return sortTemplate.Execute(dst, p)
}
