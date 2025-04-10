package dispear

import (
	"io"
	"text/template"
)

// COMMUNITY_ID adds a community_id processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/community-id-processor.html.
func COMMUNITY_ID(dst string) *CommunityIDProc {
	var pDst *string
	if dst != "" {
		pDst = &dst
	}
	p := &CommunityIDProc{TargetField: pDst}
	p.recDecl()
	p.Tag = "community_id"
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type CommunityIDProc struct {
	shared[*CommunityIDProc]

	TargetField     *string
	SourceIP        *string
	SourcePort      *string
	DestinationIP   *string
	DestinationPort *string
	IANANumber      *string
	ICMPCode        *string
	ICMPType        *string
	Transport       *string
	Seed            *uint16
	IgnoreMissing   *bool
}

func (p *CommunityIDProc) SOURCE_ID_FIELD(s string) *CommunityIDProc {
	if p.SourceIP != nil {
		panic("multiple SOURCE_ID_FIELD calls")
	}
	p.SourceIP = &s
	return p
}

func (p *CommunityIDProc) SOURCE_PORT_FIELD(s string) *CommunityIDProc {
	if p.SourcePort != nil {
		panic("multiple SOURCE_PORT_FIELD calls")
	}
	p.SourcePort = &s
	return p
}

func (p *CommunityIDProc) DESTINATION_ID_FIELD(s string) *CommunityIDProc {
	if p.DestinationIP != nil {
		panic("multiple DESTINATION_ID_FIELD calls")
	}
	p.DestinationIP = &s
	return p
}

func (p *CommunityIDProc) DESTINATION_PORT_FIELD(s string) *CommunityIDProc {
	if p.DestinationPort != nil {
		panic("multiple DESTINATION_PORT_FIELD calls")
	}
	p.DestinationPort = &s
	return p
}

func (p *CommunityIDProc) IANA_NUMBER_FIELD(s string) *CommunityIDProc {
	if p.IANANumber != nil {
		panic("multiple IANA_NUMBER_FIELD calls")
	}
	p.IANANumber = &s
	return p
}

func (p *CommunityIDProc) ICMP_CODE_FIELD(s string) *CommunityIDProc {
	if p.ICMPCode != nil {
		panic("multiple ICMP_CODE_FIELD calls")
	}
	p.ICMPCode = &s
	return p
}

func (p *CommunityIDProc) ICMP_TYPE_FIELD(s string) *CommunityIDProc {
	if p.ICMPType != nil {
		panic("multiple ICMP_TYPE_FIELD calls")
	}
	p.ICMPType = &s
	return p
}

func (p *CommunityIDProc) TRANSPORT_FIELD(s string) *CommunityIDProc {
	if p.Transport != nil {
		panic("multiple TRANSPORT_FIELD calls")
	}
	p.Transport = &s
	return p
}

func (p *CommunityIDProc) SEED(i uint16) *CommunityIDProc {
	if p.Seed != nil {
		panic("multiple SEED calls")
	}
	p.Seed = &i
	return p
}

func (p *CommunityIDProc) IGNORE_MISSING(t bool) *CommunityIDProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *CommunityIDProc) Render(dst io.Writer) error {
	communityIDTemplate := template.Must(template.New("community_id").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- community_id:` +
		preamble + `
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .SourceIP}}
    source_ip: {{yaml_string .}}
{{- end -}}
{{- with .SourcePort}}
    source_port: {{yaml_string .}}
{{- end -}}
{{- with .DestinationIP}}
    destination_ip: {{yaml_string .}}
{{- end -}}
{{- with .DestinationPort}}
    destination_port: {{yaml_string .}}
{{- end -}}
{{- with .IANANumber}}
    iana_number: {{yaml_string .}}
{{- end -}}
{{- with .ICMPCode}}
    icmp_code: {{yaml_string .}}
{{- end -}}
{{- with .ICMPType}}
    icmp_type: {{yaml_string .}}
{{- end -}}
{{- with .Seed}}
    seed: {{.}}
{{- end -}}
{{- with .Transport}}
    transport: {{yaml_string .}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
		postamble,
	))
	return communityIDTemplate.Execute(dst, p)
}
