package handlers

// import (
// 	"log"
// 	"net/http"
// 	"path/filepath"
// 	"simpleMSs/files"

// 	"github.com/gorilla/mux"
// )

// type Files struct {
// 	log   *log.Logger
// 	store files.Storage
// }

// func NewFiles(s files.Storage, l *log.Logger) *Files {
// 	return &Files{store: s, log: l}
// }

// func (f *Files) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	id := vars["id"]
// 	fn := vars["filename"]

// 	f.log.Printf("Handle POST, id: %s, filename: %s", id, fn)

// 	// по сути валидировать файлнейм не надо, ведь там regex, но ???
// 	if id == "" || fn == "" {
// 		http.Error(w, "Invalid URI", http.StatusBadRequest)
// 	}
// 	f.saveFile(id, fn, w, r)
// }

// func (f *Files) saveFile(id string, path string, w http.ResponseWriter, r *http.Request) {
// 	f.log.Println("Save file for product ", id) // , fn !!!rewrite

// 	fp := filepath.Join(id, path)
// 	err := f.store.Save(fp, r.Body)
// 	if err != nil {
// 		f.log.Println("Unable to save file")
// 		http.Error(w, "Unable to save file", http.StatusInternalServerError)
// 	}//
