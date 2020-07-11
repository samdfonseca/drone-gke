package main

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type writeTestCase struct {
	name string
	in   []byte
	out  []byte
}

// setupSecretsWriterWriteTestCases tries to create a test case for each file in `test-data/redactor/in/` with a `.yaml`
// extension. For every file it finds, a file with the same name in `test-data/redactor/out/` is required, otherwise an
// error is returned. The file in the `out` directory should be the redacted version of the file in the `in` directory.
// Test cases are named after the base filename, minus the `.yaml` extension.
func setupSecretsWriterWriteTestCases() ([]writeTestCase, error) {
	inFiles, err := filepath.Glob(`./test-data/redactor/in/*.yaml`)
	if err != nil {
		return nil, err
	}
	var tests []writeTestCase
	for _, inFile := range inFiles {
		in, err := ioutil.ReadFile(inFile)
		if err != nil {
			return nil, err
		}
		outFile := filepath.Join("./test-data/redactor/out", filepath.Base(inFile))
		outRaw, err := ioutil.ReadFile(outFile)
		if err != nil {
			return nil, err
		}
		// Unmarshal then marshal the contents of the out file since `yaml.Marshal` reorders keys, and unindents lists.
		// Without this, the output is not guaranteed to be textually equivalent.
		outTmp, err := unmarshalMultiYaml(outRaw)
		if err != nil {
			return nil, err
		}
		out, err := marshalMultiYaml(outTmp)
		name := strings.TrimSuffix(filepath.Base(inFile), ".yaml")
		tests = append(tests, writeTestCase{name: name, in: in, out: out})
	}
	return tests, nil
}

func TestSecretsWriter_Write(t *testing.T) {
	tests, err := setupSecretsWriterWriteTestCases()
	assert.NoError(t, err)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			redactedWriter := NewSecretsWriter(buf)
			_, err := redactedWriter.Write(test.in)
			assert.NoError(t, err)
			out := buf.Bytes()
			assert.Equal(t, test.out, out)
		})
	}
}
