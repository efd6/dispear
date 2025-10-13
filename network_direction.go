package dispear

import (
	"io"
	"text/template"
)

// NETWORK_DIRECTION adds a network_direction processor to the global context.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/network-direction-processor.html.
func NETWORK_DIRECTION(dst, srcip, dstip string) *NetworkDirectionProc {
	var pDst, pSrcip, pDstip *string
	if dst != "" {
		pDst = &dst
	}
	if srcip != "" {
		pSrcip = &srcip
	}
	if dstip != "" {
		pDstip = &dstip
	}
	p := &NetworkDirectionProc{TargetField: pDst, SourceIP: pSrcip, DestinationIP: pDstip}
	p.recDecl()
	p.Tag = "network_direction"
	if dst != "" {
		p.Tag += "_into_" + PathCleaner.Replace(dst)
	}
	p.template = networkDirectionTemplate
	p.parent = p
	ctx.Add(p)
	return p
}

type NetworkDirectionProc struct {
	shared[*NetworkDirectionProc]

	TargetField           *string
	SourceIP              *string
	DestinationIP         *string
	InternalNetworksField *string
	InternalNetworks      []string
	IgnoreMissing         *bool
}

func (p *NetworkDirectionProc) Name() string { return "network_direction" }

func (p *NetworkDirectionProc) INTERNAL_NETWORKS_FIELD(s string) *NetworkDirectionProc {
	if p.InternalNetworksField != nil {
		panic("multiple INTERNAL_NETWORKS_FIELD calls")
	}
	p.InternalNetworksField = &s
	return p
}

func (p *NetworkDirectionProc) INTERNAL_NETWORKS(s ...string) *NetworkDirectionProc {
	if p.InternalNetworks != nil {
		panic("multiple INTERNAL_NETWORKS calls")
	}
	p.InternalNetworks = s
	return p
}

func (p *NetworkDirectionProc) IGNORE_MISSING(t bool) *NetworkDirectionProc {
	if p.IgnoreMissing != nil {
		panic("multiple IGNORE_MISSING calls")
	}
	p.IgnoreMissing = &t
	return p
}

func (p *NetworkDirectionProc) Render(dst io.Writer, notag bool) error {
	oldNotag := p.parent.SemanticsOnly
	p.parent.SemanticsOnly = notag
	err := p.template.Execute(dst, p)
	p.parent.SemanticsOnly = oldNotag
	return err
}

var networkDirectionTemplate = template.Must(template.New("network_direction").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- {{.Name}}:` +
	preamble + `
{{- with .TargetField}}
    target_field: {{yaml_string .}}
{{- end -}}
{{- with .SourceIP}}
    source_ip: {{yaml_string .}}
{{- end -}}
{{- with .DestinationIP}}
    destination_ip: {{yaml_string .}}
{{- end -}}
{{- with .InternalNetworksField}}
    internal_networks_field: {{yaml_string .}}
{{- end -}}
{{- with .InternalNetworks}}
    internal_networks:{{range .}}
      - {{yaml_string .}}{{end}}
{{- end -}}
{{- with .IgnoreMissing}}
    ignore_missing: {{.}}
{{- end -}}` +
	postamble,
))
