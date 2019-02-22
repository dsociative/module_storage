package handlers

import (
	"log"
	"net/http"

	"html/template"

	"github.com/dsociative/module_storage/store"
	"github.com/julienschmidt/httprouter"
)

var (
	ModulesTemplate = template.Must(template.New("modules").Parse(
		`
<html>
<ul>
{{ range $id, $module := .}}
<li><a href="/module/{{ $id }}">{{ $id }}</a></li>
{{ end }}
</ul>

<form action="/add_module/" method="post">
	<input type="text" name="id" />
	<input type="submit" value="add" />
</form>
</html>
		`,
	))
	VersionsTemplate = template.Must(template.New("versions").Parse(
		`
<html>
<h3>Upload Version</h3>
<form enctype="multipart/form-data" action="/module/{{ .ID }}" method="post">
	<input type="file" name="file" />
	<input type="submit" value="upload" />
</form>

<h3>Edit Metadata</h3>
<form action="/module/{{ .ID }}/set_meta" method="post">
<input type="text" value="{{ .Module.Meta.Name }}"  name="name" placeholder="name"/>
<input type="text" value="{{ .Module.Meta.Package }}" name="package" placeholder="package"/>
<input type="text" value="{{ .Module.Meta.Description }}"  name="description" placeholder="description"/>
<input type="submit" value="edit"/>
</form>

<h3>Versions</h3>
<table>
{{ $activeVersion := .Module.ActiveVersion }}
{{ $name := .ID }}
{{ range $version, $date := .Module.Versions}}
	<tr><td><b>{{ $version }}</b></td><td>{{ $date }}</td><td>{{ if (ne $activeVersion $version) }}<a href="/module/{{ $name }}/version/{{ $version }}/set_active">SetActive</a>{{ else }} Active {{ end }}</td><tr>
{{ end }}
</table>

<a href="/">modules</a>
</html>
		`,
	))
)

type GetterFun func(*store.Store, httprouter.Params) (interface{}, error)

type TemplateHandler struct {
	Store    *store.Store
	Template *template.Template
	GetFun   GetterFun
}

func NewTemplateHandler(
	s *store.Store,
	tmpl *template.Template,
	getFun GetterFun,
) TemplateHandler {
	return TemplateHandler{s, tmpl, getFun}
}

func (h TemplateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	modules, err := h.GetFun(h.Store, p)
	if err == nil {
		err = h.Template.Execute(w, modules)
	}
	if err != nil {
		log.Println(err)
	}
}
