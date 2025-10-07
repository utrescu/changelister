package template

import (
	"os"
	"path/filepath"
	"text/template"
)

func ProcessTemplate(fileTemplate string, outputFile string, data interface{}) error {

	tmpl, err := template.New(filepath.Base(fileTemplate)).ParseFiles(fileTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		return err
	}

	return nil
}
