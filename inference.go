package dispear

import (
	"fmt"
	"io"
	"sort"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/inference-processor.html.
func INFERENCE(dst, id string) *InferenceProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &InferenceProc{ModelID: id, TargetField: pDst}
	p.recDecl()
	p.Tag = "inference_" + PathCleaner.Replace(id)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type InferenceProc struct {
	shared[*InferenceProc]

	ModelID       string
	TargetField   *string
	InputOutput   []InOutMapping
	FieldMapping  []FieldMapping
	Config        map[string]any
	IgnoreMissing *bool
}

func (p *InferenceProc) INPUT_OUTPUT(m map[string]string) *InferenceProc {
	if p.InputOutput != nil {
		panic("multiple INPUT_OUTPUT calls")
	}
	p.InputOutput = make([]InOutMapping, 0, len(m))
	for k, v := range m {
		p.InputOutput = append(p.InputOutput, InOutMapping{Input: k, Output: v})
	}
	sort.Slice(p.InputOutput, func(i, j int) bool {
		return p.InputOutput[i].Input < p.InputOutput[j].Input
	})
	return p
}

func (p *InferenceProc) FIELD_MAP(m map[string]string) *InferenceProc {
	if p.FieldMapping != nil {
		panic("multiple FIELD_MAPPING calls")
	}
	p.FieldMapping = make([]FieldMapping, 0, len(m))
	for k, v := range m {
		p.FieldMapping = append(p.FieldMapping, FieldMapping{Document: k, Model: v})
	}
	sort.Slice(p.FieldMapping, func(i, j int) bool {
		return p.FieldMapping[i].Document < p.FieldMapping[j].Document
	})
	return p
}

func (p *InferenceProc) INFERENCE_CONFIG(m map[string]any) *InferenceProc {
	if p.Config != nil {
		panic("multiple INFERENCE_CONFIG calls")
	}
	p.Config = m
	return p
}

func (p *InferenceProc) IGNORE_MISSING(t bool) *InferenceProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *InferenceProc) Render(dst io.Writer) error {
	if p.ModelID == "" {
		return fmt.Errorf("no model ID for INFERENCE %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.InputOutput != nil && (p.TargetField != nil || p.FieldMapping != nil) {
		return fmt.Errorf("using input_output with target_field/field_map for INFERENCE %s:%d: %s", p.file, p.line, p.Tag)
	}
	htmlStripTemplate := template.Must(template.New("inference").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- inference:` +
		preamble + `
    model_id: {{yaml_string .ModelID}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .InputOutput}}
    input_output:{{range .}}
      - input_field: {{yaml_string .Input}}
        output_field: {{yaml_string .Output}}{{end}}
{{- end -}}
{{- with .FieldMapping}}
    field_map:{{range .}}
      {{yaml_string .Document}}: {{yaml_string .Model}}{{end}}
{{- end -}}
{{- with .Config}}
{{yaml 4 2 "inference_config" .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return htmlStripTemplate.Execute(dst, p)
}
