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
