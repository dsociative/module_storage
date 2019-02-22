package main

import (
	"flag"
	"net/http"

	"github.com/dsociative/module_storage/handlers"
	"github.com/dsociative/module_storage/store"
	"github.com/julienschmidt/httprouter"
)

var (
	path = flag.String("path", "/tmp/modules", "modules dir")
	bind = flag.String("bind", ":8080", "bind addr")
)

func main() {
	flag.Parse()
	s := store.NewStore(*path)
	router := httprouter.New()
	handlers := handlers.NewHandlers(s)
	router.GET(
		"/",
		handlers.ModuleList.ServeHTTP,
	)
	router.GET(
		"/module/:module",
		handlers.Module.ServeHTTP,
	)
	router.POST(
		"/module/:module",
		handlers.AddModuleVersion,
	)
	router.POST(
		"/module/:module/set_meta",
		handlers.SetModuleMeta,
	)
	router.GET(
		"/module/:module/active",
		handlers.ActiveVersion,
	)
	router.POST(
		"/add_module",
		handlers.NewModule,
	)
	router.POST(
		"/sync",
		handlers.Sync,
	)
	router.GET(
		"/module/:module/version/:version/set_active",
		handlers.SetModuleVersion,
	)
	http.ListenAndServe(*bind, router)
}
