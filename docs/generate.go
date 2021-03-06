package main

//go:generate go run generate.go

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/aws-cloudformation/rain/cmd"
	"github.com/spf13/cobra/doc"
)

var tmpl *template.Template

func init() {
	var err error

	tmpl = template.New("README.tmpl")

	tmpl = tmpl.Funcs(template.FuncMap{
		"pad": func(s string, n int) string {
			return strings.Repeat(" ", n-len(s))
		},
	})
	if err != nil {
		panic(err)
	}

	tmpl, err = tmpl.ParseFiles("README.tmpl")
	if err != nil {
		panic(err)
	}
}

func emptyStr(s string) string {
	return ""
}

func identity(s string) string {
	if s == "rain.md" {
		return "index.md"
	}

	return s
}

func main() {
	err := doc.GenMarkdownTreeCustom(cmd.Root, "./", emptyStr, identity)
	if err != nil {
		panic(err)
	}

	err = os.Rename("rain.md", "index.md")
	if err != nil {
		panic(err)
	}

	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, cmd.Root)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("../README.md", buf.Bytes(), 0644)

	cmd.Root.GenBashCompletionFile("bash_completion.sh")
	cmd.Root.GenZshCompletionFile("zsh_completion.sh")
}
