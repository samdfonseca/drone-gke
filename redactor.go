package main

import (
	"bytes"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"
)

type SecretsWriter struct {
	out io.Writer
}

func NewSecretsWriter(out io.Writer) *SecretsWriter {
	return &SecretsWriter{out: out}
}

// Write parses resource YAML bytes p, replaces the value of all data and stringData fields in Secret resources, then
// writes the redacted resource out to the underlying io.Writer. If p is not valid YAML, an error is returned.
func (rw *SecretsWriter) Write(p []byte) (n int, err error) {
	vv, err := unmarshalMultiYaml(p)
	if err != nil {
		return n, err
	}
	for i, v := range vv {
		if v["kind"].(string) == "Secret" {
			for _, k := range []string{"data", "stringData"} {
				data, ok := v[k]
				if ok {
					data, ok := data.(map[string]interface{})
					if ok {
						for kk := range data {
							data[kk] = "<REDACTED>"
						}
						vv[i][k] = data
					}
				}
			}
		}
	}
	b, err := marshalMultiYaml(vv)
	if err != nil {
		return n, err
	}
	return fmt.Fprint(rw.out, string(b))
}

func unmarshalMultiYaml(y []byte) ([]map[string]interface{}, error) {
	var o []map[string]interface{}
	for _, b := range bytes.Split(y, []byte("---\n")) {
		if b == nil || len(b) == 0 {
			continue
		}
		var v map[string]interface{}
		if err := yaml.Unmarshal(b, &v); err != nil {
			return nil, err
		}
		o = append(o, v)
	}
	return o, nil
}

func marshalMultiYaml(y []map[string]interface{}) ([]byte, error) {
	var o []byte
	for _, v := range y {
		b, err := yaml.Marshal(v)
		if err != nil {
			return nil, err
		}
		o = append(o, "---\n"...)
		o = append(o, b...)
	}
	return o, nil
}
