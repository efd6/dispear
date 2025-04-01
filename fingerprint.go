package dispear

import (
	"fmt"
	"io"
	"slices"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/fingerprint-processor.html.
func FINGERPRINT(dst string, src ...string) *FingerprintProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &FingerprintProc{Fields: src, TargetField: pDst}
	p.recDecl()
	p.Tag = "fingerprint"
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type FingerprintProc struct {
	shared[*FingerprintProc]

	Fields        []string
	TargetField   *string
	Method        *string
	Salt          *string
	IgnoreMissing *bool
}

func (p *FingerprintProc) METHOD(s string) *FingerprintProc {
	if p.Method != nil {
		panic("multiple METHOD calls")
	}
	p.Method = &s
	return p
}

func (p *FingerprintProc) SALT(s string) *FingerprintProc {
	if p.Method != nil {
		panic("multiple SALT calls")
	}
	p.Method = &s
	return p
}

func (p *FingerprintProc) IGNORE_MISSING(t bool) *FingerprintProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *FingerprintProc) Render(dst io.Writer) error {
	if len(p.Fields) == 0 {
		return fmt.Errorf("no src for FINGERPRINT %s:%d: %s", p.file, p.line, p.Tag)
	}
	if slices.Contains(p.Fields, "") {
		return fmt.Errorf("empty src element for FINGERPRINT %s:%d: %s", p.file, p.line, p.Tag)
	}
	fingerprintTemplate := template.Must(template.New("fingerprint").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- fingerprint:` +
		preamble + `
{{- with .Fields}}
    fields:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .Method}}
    method: {{yaml_string .}}
{{- end -}}
{{- with .Salt}}
    salt: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return fingerprintTemplate.Execute(dst, p)
}
