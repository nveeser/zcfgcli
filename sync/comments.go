package sync

import (
	"fmt"
	"github.com/mitchellh/go-wordwrap"
	"io"
	"strings"
	"text/template"
	"zcfgcli/meta"
)

var commentTmpl = `
--------------------------
Name: {{ $.Name }}
--------------------------
{{ range $d := wrap $.Description -}}
{{ $d }}
{{ end -}}
Type: {{ $.Type }}
{{ if $.Values -}}
Values: {{ printValues $.Values }}
{{ else -}}
Range: ({{ $.ValueMin }}, {{ $.ValueMax }})
{{ end -}}
{{ range $preset := $.Presets -}}
Preset: {{ $preset.Name }} => {{ $preset.Value }} {{ printf "%q" $preset.Description }}
{{ end -}}
`

type commentBuilder struct {
	width uint
}

func (b *commentBuilder) WriteString(entity meta.Entity) (string, error) {
	var buf strings.Builder
	if err := b.Write(&buf, entity); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (b *commentBuilder) Write(w io.Writer, entity meta.Entity) error {
	fMap := template.FuncMap{
		"wrap":        b.wrapDescription,
		"printValues": b.printValues,
	}
	mapTemplate, err := template.New("map").Funcs(fMap).Parse(commentTmpl)
	if err != nil {
		return err
	}
	return mapTemplate.Execute(w, entity)
}

func (b *commentBuilder) wrapDescription(s string) []string {
	return strings.Split(wordwrap.WrapString(s, b.width), "\n")
}
func (b *commentBuilder) printValues(values []string) string {
	return fmt.Sprintf("%q", values)
	//var quoted []string
	//for _, v := range values {
	//
	//}
	//return strings.Split(wordwrap.WrapString(s, b.width), "\n")
}
