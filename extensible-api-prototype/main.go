package main

import (
	"fmt"
	"net/http"
	"plugin"
    "os"
	"path/filepath"
	"strings"
	"encoding/json"

	"github.com/gorilla/mux"
    "github.com/praetorian-inc/microservices-golang/nicohttp"
)

var (
	server *nicohttp.NicoServer
	loadedModules = make(map[string]string)
)

func main() {
	b := nicohttp.GetBuilder()
	b = b.WithDefaults().WithBaseFlags()
	server, _ = b.Create("extensible-api-service", 8080)
	server.Mux().HandleFunc("/load", load).Methods("POST").Queries("path", "{path}", "name", "{name}").Name("LoadAPIExtension")
	server.Mux().HandleFunc("/loadAll", loadAll).Methods("POST").Name("LoadAllModules")
	server.Mux().HandleFunc("/modules", list).Methods("GET").Name("GetLoadedModules")
	server.Mux().HandleFunc("/unload", load).Methods("POST").Queries("path", "{path}", "name", "{name}").Name("UnloadAPIExtension")

	fmt.Printf("%v\n", b.Props())
	server.Start()
}



func load(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	if path == "" {
		http.Error(w, "path not defined", http.StatusBadRequest)
		return
	}
	name := mux.Vars(r)["name"]
	if name == "" {
		http.Error(w, "name not defined", http.StatusBadRequest)
		return
	}

	// load module
	mod := fmt.Sprintf("%s/%s.so", path, name)
	plug, err := plugin.Open(mod)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not load module %s", mod), http.StatusBadRequest)
		return
	}

	// verify symbol
	sym, err := plug.Lookup("APIExtension")
	if err != nil {
		http.Error(w, fmt.Sprintf("could not verify symbol %s", "APIExtension"), http.StatusBadRequest)
		return
	}

	// Type assert loaded symbol is of desired type
	var extender APIExtension
	extender, ok := sym.(APIExtension)
	if !ok {
		http.Error(w, fmt.Sprintf("unexpected type from module symbol %s", "APIExtension"), http.StatusBadRequest)
		return
	}
	routeName := extender.ERoute(server.Mux())
	loadedModules[mod] = routeName
}


func unload(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	if path == "" {
		http.Error(w, "path not defined", http.StatusBadRequest)
		return
	}
	name := mux.Vars(r)["name"]
	if name == "" {
		http.Error(w, "name not defined", http.StatusBadRequest)
		return
	}

	mod := fmt.Sprintf("%s/%s.so", path, name)
	//var rn string
	if _, ok := loadedModules[mod]; !ok {
		http.Error(w, fmt.Sprintf("module %s not currently loaded", mod), http.StatusBadRequest)
		return
	}
	delete(loadedModules, mod)
	//Swap Router - TBD
}


func loadAll(w http.ResponseWriter, r *http.Request) {
	mods, err := getModules()
	if err != nil {
		http.Error(w, fmt.Sprintf("error loading modules from directory %s: error %s", "./modules", err), http.StatusInternalServerError)
		return
	}

	moduleStatus := make(map[string]string)
	for _, m := range mods {
		if _, ok := loadedModules[m]; ok {
			moduleStatus[m] = fmt.Sprintf("FAILED: Module already loaded")
			continue
		}
		// load module
		plug, err := plugin.Open(m)
		if err != nil {
			moduleStatus[m] = fmt.Sprintf("FAILED: %s", err)
			continue
		}

		// verify symbol
		sym, err := plug.Lookup("APIExtension")
		if err != nil {
			moduleStatus[m] = fmt.Sprintf("FAILED: %s", err)
			continue
		}	

		// Type assert loaded symbol is of desired type
		var extender APIExtension
		extender, ok := sym.(APIExtension)
		if !ok {
			moduleStatus[m] = fmt.Sprintf("FAILED: %s", err)
			continue
		}
		routeName := extender.ERoute(server.Mux())
		moduleStatus[m] = fmt.Sprintf("SUCCESS")
		loadedModules[m] = routeName
	}

	js, err := json.Marshal(moduleStatus)
	if (err != nil) {
		http.Error(w, fmt.Sprintf("JSON marshalling error: %s", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


func list(w http.ResponseWriter, r *http.Request) {
	
	js, err := json.MarshalIndent(loadedModules, "", "\t")
	if (err != nil) {
		http.Error(w, fmt.Sprintf("JSON marshalling error: %s", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}



func getModules() ([]string, error) {
	var files []string

    root := "./modules"
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".so") {
			files = append(files, path)
		}
        return nil
	})
	return files, err
}
