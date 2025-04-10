package dispear

import (
	"fmt"
	"io"
	"text/template"
)

// GEO_GRID adds a geo_grid processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-geo-grid-processor.html.
func GEO_GRID(dst, src, typ string) *GeoGridProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &GeoGridProc{Field: src, TileType: typ, TargetField: pDst}
	p.recDecl()
	p.Tag = "geo_grid_" + PathCleaner.Replace(src) + "_as_" + typ
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type GeoGridProc struct {
	shared[*GeoGridProc]

	Field            string
	TileType         string
	TargetField      *string
	IgnoreMissing    *bool
	ChildrenField    *string
	NonChildrenField *string
	ParentField      *string
	PrecisionField   *string
	TargetFormat     *string
}

func (p *GeoGridProc) IGNORE_MISSING(t bool) *GeoGridProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *GeoGridProc) CHILDREN_FIELD(s string) *GeoGridProc {
	if p.ChildrenField != nil {
		panic("multiple CHILDREN_FIELD calls")
	}
	p.ChildrenField = &s
	return p
}

func (p *GeoGridProc) NON_CHILDREN_FIELD(s string) *GeoGridProc {
	if p.NonChildrenField != nil {
		panic("multiple NON_CHILDREN_FIELD calls")
	}
	p.NonChildrenField = &s
	return p
}

func (p *GeoGridProc) PARENT_FIELD(s string) *GeoGridProc {
	if p.ParentField != nil {
		panic("multiple PARENT_FIELD calls")
	}
	p.ParentField = &s
	return p
}

func (p *GeoGridProc) PRECISION_FIELD(s string) *GeoGridProc {
	if p.PrecisionField != nil {
		panic("multiple PRECISION_FIELD calls")
	}
	p.PrecisionField = &s
	return p
}

func (p *GeoGridProc) TARGET_FORMAT(s string) *GeoGridProc {
	if p.TargetFormat != nil {
		panic("multiple TARGET_FORMAT calls")
	}
	p.TargetFormat = &s
	return p
}

func (p *GeoGridProc) Render(dst io.Writer) error {
	if p.Field == "" {
		return fmt.Errorf("no src for GEO_GRID %s:%d: %s", p.file, p.line, p.Tag)
	}
	if p.TileType == "" {
		return fmt.Errorf("no tile type for GEO_GRID %s:%d: %s", p.file, p.line, p.Tag)
	}
	geoGridTemplate := template.Must(template.New("geo_grid").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- geo_grid:` +
		preamble + `
    field: {{yaml_string .Field}}
    tile_type: {{yaml_string .TileType}}
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .TargetFormat}}
    target_format: {{yaml_string .}}
{{- end -}}
{{- with .ChildrenField}}
    children_field: {{yaml_string .}}
{{- end -}}
{{- with .NonChildrenField}}
    non_children_field: {{yaml_string .}}
{{- end -}}
{{- with .ParentField}}
    parent_field: {{yaml_string .}}
{{- end -}}
{{- with .PrecisionField}}
    precision_field: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return geoGridTemplate.Execute(dst, p)
}
