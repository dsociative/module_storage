package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dsociative/module_storage/store"
	"github.com/julienschmidt/httprouter"
)

type Handlers struct {
	Store      *store.Store
	Module     TemplateHandler
	ModuleList TemplateHandler
}

func NewHandlers(
	s *store.Store,
) Handlers {
	moduleListHandler := NewTemplateHandler(
		s,
		ModulesTemplate,
		func(s *store.Store, _ httprouter.Params) (interface{}, error) {
			return s.Modules()
		})
	moduleHandler := NewTemplateHandler(
		s,
		VersionsTemplate,
		func(s *store.Store, p httprouter.Params) (interface{}, error) {
			m, err := s.Modules()
			if err == nil {
				id := p.ByName("module")
				data := struct {
					ID     string
					Module store.ModuleMetadata
				}{ID: id, Module: m[p.ByName("module")]}
				return data, err
			}
			return nil, err
		})
	return Handlers{
		s,
		moduleHandler,
		moduleListHandler,
	}
}

func (h Handlers) AddModuleVersion(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	req.ParseMultipartForm(700000)
	file, _, err := req.FormFile("file")
	if err == nil {
		err = h.Store.AddModuleVersion(p.ByName("module"), time.Now(), file)
	}
	if err != nil {
		log.Println(err)
	}
	h.Module.ServeHTTP(w, req, p)
}

func (h Handlers) SetModuleMeta(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	req.ParseForm()
	err := h.Store.SetModuleMeta(
		p.ByName("module"),
		req.Form.Get("name"),
		req.Form.Get("package"),
		req.Form.Get("description"),
	)
	if err != nil {
		log.Println(err)
	}
	h.ModuleList.ServeHTTP(w, req, p)
}

func (h Handlers) SetModuleVersion(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	version, err := strconv.Atoi(p.ByName("version"))
	if err != nil {
		log.Println(err)
	} else {
		h.Store.SetModuleVersion(p.ByName("module"), version)
	}
	http.Redirect(w, req, fmt.Sprintf("/module/%s", p.ByName("module")), http.StatusTemporaryRedirect)
}

func (h Handlers) NewModule(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	req.ParseForm()
	err := h.Store.NewModule(req.Form.Get("id"))
	if err != nil {
		log.Println(err)
	}
	h.ModuleList.ServeHTTP(w, req, p)
}

func (h Handlers) ActiveVersion(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	id := p.ByName("module")
	module, reader, err := h.Store.ActiveVersion(id)
	if err == nil {
		w.Header().Set("ID", id)
		w.Header().Set("VERSION", strconv.Itoa(module.ActiveVersion))
		w.Header().Set("NAME", module.Meta.Name)
		w.Header().Set("DESCRIPTION", module.Meta.Description)
		w.Header().Set("PACKAGE", module.Meta.Package)
		_, err = io.Copy(w, reader)
	}
	if err != nil {
		log.Println(err)
	}
}

type Module struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type SyncRequest struct {
	InstalledModules []Module `json:"installedModules"`
}

type SyncResponse struct {
	UpdatedModules []Module `json:"updatedModules"`
}

func (h Handlers) Sync(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	var sRequest SyncRequest
	var err error
	if err = json.NewDecoder(req.Body).Decode(&sRequest); err == nil {
		var m store.MapModuleMetadata
		if m, err = h.Store.Modules(); err == nil {
			sResponse := SyncResponse{UpdatedModules: []Module{}}
			for _, installedModule := range sRequest.InstalledModules {
				if module, ok := m[installedModule.ID]; ok {
					if module.ActiveVersion != installedModule.Version {
						sResponse.UpdatedModules = append(
							sResponse.UpdatedModules,
							Module{installedModule.ID, module.ActiveVersion},
						)
					}
				}
			}
			err = json.NewEncoder(w).Encode(&sResponse)
		}
	}
	if err != nil {
		log.Println(err)
	}
}
