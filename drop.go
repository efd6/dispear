package dispear

import (
	"io"
	"text/template"
)

// DROP adds a drop processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/drop-processor.html.
func DROP(reason string) *StopProc {
	p := &StopProc{Flavour: "drop"}
	p.recDecl()
	p.Tag = "drop"
	if reason != "" {
		p.Tag += "_" + PathCleaner.Replace(reason)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

// TERMINATE adds a terminate processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/terminate-processor.html.
func TERMINATE(reason string) *StopProc {
	p := &StopProc{Flavour: "terminate"}
	p.recDecl()
	p.Tag = "terminate"
	if reason != "" {
		p.Tag += "_" + PathCleaner.Replace(reason)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type StopProc struct {
	shared[*StopProc]

	Flavour string
}

func (p *StopProc) Render(dst io.Writer) error {
	dropTemplate := template.Must(template.New("drop").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Flavour}}:` +
		preamble +
		postamble,
	))
	return dropTemplate.Execute(dst, p)
}
