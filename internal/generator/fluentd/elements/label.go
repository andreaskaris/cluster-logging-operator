package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

type FromLabel struct {
	framework.InLabel
	SubElements []framework.Element
	Desc        string
}

func (f FromLabel) Name() string {
	return "pipeline"
}

func (f FromLabel) Template() string {
	return `{{define "` + f.Name() + `"  -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
<label {{.InLabel}}>
{{compose .Elements | indent 2}}
</label>
{{end}}`
}

func (f FromLabel) Elements() []framework.Element {
	return f.SubElements
}

type Pipeline = FromLabel
