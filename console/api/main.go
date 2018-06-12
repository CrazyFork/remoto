package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func init() {
	http.HandleFunc("/api/templates", handleTemplates())
	http.HandleFunc("/api/templates/", handleRenderTemplate())
}

// handleTemplates gets a list of available templates.
// 	GET /api/templates
func handleTemplates() http.HandlerFunc {
	var init sync.Once
	var err error
	type template struct {
		Name string `json:"name"`
	}
	var response struct {
		Templates []template `json:"templates"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		init.Do(func() {
			err = filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil // skip directories
				}
				path, err = filepath.Rel("templates", path)
				if err != nil {
					return err
				}
				if !strings.Contains(path, "/") {
					return nil // skip root files
				}
				if filepath.Ext(path) == ".plush" {
					path = path[0 : len(path)-len(".plush")]
				}
				response.Templates = append(response.Templates, template{
					Name: path,
				})
				return nil
			})
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleRenderTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		//templatePath := r.URL.Path[len("/api/"):] + ".plush"
		// b, err := ioutil.ReadFile(templatePath)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusBadRequest)
		// 	return
		// }

		// TODO: parse this with the Remoto generator when it can
		// accept files.

	}
}
