package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// RENAME adds a rename processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/rename-processor.html.
func RENAME(from, to string) *RenameProc {
	p := &RenameProc{Field: from, TargetField: to}
	p.recDecl()
	p.Tag = "rename_" + PathCleaner.Replace(from) + "_to_" + PathCleaner.Replace(to)
	p.template = renameTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type RenameProc struct {
	shared[*RenameProc]

	Field         string
	TargetField   string
	Override      *bool
	IgnoreMissing *bool
}

func (p *RenameProc) IGNORE_MISSING(t bool) *RenameProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *RenameProc) OVERRIDE(t bool) *RenameProc {
	if p.Override != nil {
		panic("multiple OVERRIDE calls")
	}
	p.Override = &t
	return p
}

func (p *RenameProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no from name for RENAME %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.TargetField == "" {
		return fmt.Errorf("no to name for RENAME %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var renameTemplate = template.Must(template.New("rename").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- rename:` +
	preamble + `
    field: {{yaml_string .Field}}
    target_field: {{yaml_string .TargetField}}
{{- with .Override}}
    override: {{.}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
	postamble,
))
