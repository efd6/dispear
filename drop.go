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
	p.template = dropTemplate
	p.parent = p
	ctx.Add(p)
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
	p.template = dropTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type StopProc struct {
	shared[*StopProc]

	Flavour string
}

func (p *StopProc) Name() string { return p.Flavour }

func (p *StopProc) Render(dst io.Writer, notag bool) error {
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var dropTemplate = template.Must(template.New("drop").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Flavour}}:` +
	preamble +
	postamble,
))
