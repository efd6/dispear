package dispear

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

// GEOIP adds a geoip processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/geoip-processor.html.
func GEOIP(dst, src string) *IPLocationProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &IPLocationProc{Field: src, TargetField: pDst, Flavour: "geoip"}
	p.recDecl()
	p.Tag = "geoip_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

// IP_LOCATION adds an ip_location processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/ip-location-processor.html.
func IP_LOCATION(dst, src string) *IPLocationProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &IPLocationProc{Field: src, TargetField: pDst, Flavour: "ip_location"}
	p.recDecl()
	p.Tag = "ip_location_" + PathCleaner.Replace(src)
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type IPLocationProc struct {
	shared[*IPLocationProc]

	Flavour                    string
	Field                      string
	TargetField                *string
	IgnoreMissing              *bool
	DatabaseFile               *string
	DownloadOnPipelineCreation *bool
	FirstOnly                  *bool
	Properties                 []string
}

func (p *IPLocationProc) IGNORE_MISSING(t bool) *IPLocationProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *IPLocationProc) DATABASE_FILE(s string) *IPLocationProc {
	if p.DatabaseFile != nil {
		panic("multiple DATABASE_FILE calls")
	}
	p.DatabaseFile = &s
	return p
}

func (p *IPLocationProc) DOWNLOAD_ON_PIPELINE_CREATION(t bool) *IPLocationProc {
	if p.DownloadOnPipelineCreation != nil {
		panic("multiple DOWNLOAD_ON_PIPELINE_CREATION calls")
	}
	p.DownloadOnPipelineCreation = &t
	return p
}

func (p *IPLocationProc) FIRST_ONLY(t bool) *IPLocationProc {
	if p.FirstOnly != nil {
		panic("multiple FIRST_ONLY calls")
	}
	p.FirstOnly = &t
	return p
}

func (p *IPLocationProc) PROPERTIES(s ...string) *IPLocationProc {
	if p.Properties != nil {
		panic("multiple PROPERTIES calls")
	}
	p.Properties = s
	return p
}

func (p *IPLocationProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for %s %s:%d: %s", strings.ToUpper(p.Flavour), p.file, p.line, p.Tag)
	}
	ipLocationTemplate := template.Must(template.New(p.Flavour).Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Flavour}}:` +
		preamble + `
    field: {{yaml_string .Field}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .DatabaseFile}}
    database_file: {{yaml_string .}}
{{- end -}}
{{- with .DownloadOnPipelineCreation}}
    download_database_on_pipeline_creation: {{.}}
{{- end -}}
{{- with .FirstOnly}}
    first_only: {{.}}
{{- end -}}
{{- with .Properties}}
    properties:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return ipLocationTemplate.Execute(dst, p)
}
