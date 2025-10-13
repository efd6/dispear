package dispear

import (
	"bytes"
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

func yamlString(s string) (string, error) {
	var n yaml.Node
	n.SetString(s)
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	err := enc.Encode(&n)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}

// There must be a better way to do this, but given that this is YAML,
// probably not.
//
// ¯\_(ツ)_/¯
func yamlValue(pre, in int, k string, v any) (string, error) {
	v = map[string]any{k: v}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return "", err
	}
	dec := yaml.NewDecoder(&buf)
	var n yaml.Node
	err = dec.Decode(&n)
	if err != nil {
		return "", err
	}
	buf.Reset()
	enc = yaml.NewEncoder(&buf)
	enc.SetIndent(in)
	err = enc.Encode(n.Content[0])
	if err != nil {
		return "", err
	}
	return indent(strings.TrimRight(buf.String(), "\n"), pre), nil
}

func (p *shared[P]) setSemantics() error {
	var buf bytes.Buffer
	p.SemanticsOnly = true
	err := p.template.Execute(&buf, p.parent)
	p.SemanticsOnly = false
	if err != nil {
		return err
	}
	var m []map[string]any
	err = yaml.Unmarshal(buf.Bytes(), &m)
	if err != nil {
		return err
	}
	b, err := json.Marshal(m[0])
	if err != nil {
		return err
	}
	if p.semantic == nil {
		p.semantic = &semantic{}
	}
	p.semantic.data = b
	return nil
}
