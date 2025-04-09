package dispear

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

// See https://www.elastic.co/guide/en/elasticsearch/reference/current/reroute-processor.html.
func REROUTE(namespace, dataset, destination string) *RerouteProc {
	var pNamespace, pDataset, pDestination *string
	if namespace != "" {
		pNamespace = &namespace
	}
	if dataset != "" {
		pDataset = &dataset
	}
	if destination != "" {
		pDestination = &destination
	}
	p := &RerouteProc{Namespace: pNamespace, Dataset: pDataset, Destination: pDestination}
	p.recDecl()
	p.Tag = "reroute"
	if namespace != "" || dataset != "" {
		p.Tag += "_to"
	}
	if namespace != "" {
		p.Tag += "_" + PathCleaner.Replace(namespace)
	}
	if dataset != "" {
		p.Tag += "_" + PathCleaner.Replace(dataset)
	}
	if destination != "" {
		p.Tag += "_to_" + PathCleaner.Replace(destination)
	}
	p.parent = p
	ctx.Add(p)
	ctx.tags[p.Tag] = append(ctx.tags[p.Tag], &p.shared)
	return p
}

type RerouteProc struct {
	shared[*RerouteProc]

	Namespace   *string
	Dataset     *string
	Destination *string
}

func (p *RerouteProc) Render(dst io.Writer) error {
	if (p.Destination != nil) == (p.Namespace != nil || p.Dataset != nil) {
		return fmt.Errorf("no destination provided with namespace or dataset for REROUTE %s:%d: %s", p.file, p.line, p.Tag)
	}
	if !isValidRerouteName(p.Namespace) {
		return fmt.Errorf("invalid namespace name for REROUTE %s:%d: %s", p.file, p.line, p.Tag)
	}
	if !isValidRerouteName(p.Dataset) {
		return fmt.Errorf("invalid dataset name for REROUTE %s:%d: %s", p.file, p.line, p.Tag)
	}
	rerouteTemplate := template.Must(template.New("reroute").Funcs(templateHelpers).Parse(`
{{with .Comment}}{{comment .}}
{{end}}- reroute:` +
		preamble + `
{{- with .Namespace}}
    namespace: {{yaml_string .}}
{{- end -}}
{{- with .Dataset}}
    dataset: {{yaml_string .}}
{{- end -}}
{{- with .Destination}}
    destination: {{yaml_string .}}
{{- end -}}` +
		postamble,
	))
	return rerouteTemplate.Execute(dst, p)
}

func isValidRerouteName(p *string) bool {
	if p == nil {
		return true
	}
	s := *p
	switch {
	case s == "", len(s) > 100, s == ".", s == "..",
		strings.ContainsAny(s[:1], "-_+"), strings.ContainsAny(s, `-\/*?"<>| ,#`):
		return false
	default:
		return true
	}
}
