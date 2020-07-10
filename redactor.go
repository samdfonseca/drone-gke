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

func (rw *SecretsWriter) Write(p []byte) (n int, err error) {
	for _, pp := range bytes.Split(p, []byte("---\n")) {
		if pp == nil || len(pp) == 0 {
			continue
		}
		var v map[string]interface{}
		if err := yaml.Unmarshal(pp, &v); err != nil {
			return n, err
		}
		if v["kind"].(string) == "Secret" {
			data, ok := v["data"]
			if ok {
				data, ok := data.(map[string]interface{})
				if ok {
					for k := range data {
						data[k] = "<REDACTED>"
					}
				}
			}
		}
		b, err := yaml.Marshal(v)
		if err != nil {
			return n, err
		}
		nn, err := fmt.Fprintf(rw.out, "---\n%s", b)
		n = n + nn
		if err != nil {
			return n, err
		}
	}
	return n, nil
}
