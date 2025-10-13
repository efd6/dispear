package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// SPLIT adds a split processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/split-processor.html.
func SPLIT(dst, src, sep string) *SplitProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &SplitProc{Field: src, TargetField: pDst, Separator: sep}
	p.recDecl()
	p.Tag = "split_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = splitTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type SplitProc struct {
	shared[*SplitProc]

	Field            string
	Separator        string
	TargetField      *string
	PreserveTrailing *bool
	IgnoreMissing    *bool
}

func (p *SplitProc) Name() string { return "split" }

func (p *SplitProc) PRESERVE_TRAILING(t bool) *SplitProc {
	if p.PreserveTrailing != nil {
		panic("multiple PRESERVE_TRAILING calls")
	}
	p.PreserveTrailing = &t
	return p
}

func (p *SplitProc) IGNORE_MISSING(t bool) *SplitProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *SplitProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for SPLIT %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.Separator == "" {
		return fmt.Errorf("no separator for SPLIT %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var splitTemplate = template.Must(template.New("split").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
	preamble + `
    field: {{yaml_string .Field}}
    separator: {{yaml_string .Separator}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .PreserveTrailing}}
    preserve_trailing: {{.}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
	postamble,
))
