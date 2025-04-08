package dispear

import (
	"bytes"
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
func yamlValue(v any) (string, error) {
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
	err = enc.Encode(n.Content[0])
	if err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}
