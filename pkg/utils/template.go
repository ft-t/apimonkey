package utils

import (
	"bytes"
	"text/template"

	"github.com/cockroachdb/errors"
)

func ExecuteTemplate(input string, cfg any) (string, error) {
	if input == "" {
		return input, nil
	}

	parsed, err := template.New("any").Parse(input)
	if err != nil {
		return "", errors.Wrap(err, "can not parse template")
	}

	var buf bytes.Buffer
	if err = parsed.Execute(&buf, cfg); err != nil {
		return "", errors.Wrap(err, "can not execute template")
	}

	return buf.String(), nil
}
