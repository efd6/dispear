package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// JSON adds a json processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/json-processor.html.
func JSON(dst, src string) *JSONProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &JSONProc{Field: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "json_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = jsonTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type JSONProc struct {
	shared[*JSONProc]

	Field                     string
	TargetField               *string
	AddToRoot                 *bool
	AddToRootConflictStrategy *string
	AllowDuplicateKeys        *bool
	StrictJSONParsing         *bool
}

func (p *JSONProc) ADD_TO_ROOT(t bool) *JSONProc {
	if p.AddToRoot != nil {
		panic("multiple ADD_TO_ROOT calls")
	}
	p.AddToRoot = &t
	return p
}

func (p *JSONProc) ADD_TO_ROOT_CONFLICT_STRATEGY(s string) *JSONProc {
	if p.AddToRootConflictStrategy != nil {
		panic("multiple ADD_TO_ROOT_CONFLICT_STRATEGY calls")
	}
	p.AddToRootConflictStrategy = &s
	return p
}

func (p *JSONProc) ALLOW_DUPLICATE_KEYS(t bool) *JSONProc {
	if p.AllowDuplicateKeys != nil {
		panic("multiple ALLOW_DUPLICATE_KEYS calls")
	}
	p.AllowDuplicateKeys = &t
	return p
}

func (p *JSONProc) STRICT_JSON_PARSING(t bool) *JSONProc {
	if p.StrictJSONParsing != nil {
		panic("multiple STRICT_JSON_PARSING calls")
	}
	p.StrictJSONParsing = &t
	return p
}

func (p *JSONProc) Render(dst io.Writer, notag bool) error {
	if p.Field == "" {
		return fmt.Errorf("no src for JSON %s:%d: %s", p.file, p.line, p.Tag)
	}
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var jsonTemplate = template.Must(template.New("json").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- json:` +
	preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .AddToRoot}}
    add_to_root: {{.}}
{{- end -}}
{{- with .AddToRootConflictStrategy}}
    add_to_root_conflict_strategy: {{yaml_string .}}
{{- end -}}
{{- with .AllowDuplicateKeys}}
    allow_duplicate_keys: {{.}}
{{- end -}}
{{- with .StrictJSONParsing}}
    strict_json_parsing: {{.}}
{{- end -}}` +
	postamble,
))
