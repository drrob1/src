package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"src/liveproject/secretapp/filestore"
	"src/liveproject/secretapp/types"
	//"github.com/amitsaha/manning-go-advanced-beginners/project-series-2/mini-project-1/milestone1-code/filestore"
	//"github.com/amitsaha/manning-go-advanced-beginners/project-series-2/mini-project-1/milestone1-code/types"
)

func secretHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		createSecret(w, r)
	case "GET":
		getSecret(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func getHash(plainText string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(plainText)))
}

func createSecret(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}
	p := types.CreateSecretPayload{}
	err = json.Unmarshal(bytes, &p)
	if err != nil || len(p.PlainText) == 0 {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	digest := getHash(p.PlainText)
	resp := types.CreateSecretResponse{Id: digest}

	s := types.SecretData{Id: resp.Id, Secret: p.PlainText}
	err = filestore.FileStoreConfig.Fs.Write(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jd, err := json.Marshal(&resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jd)
}

func getSecret(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path
	id = strings.TrimPrefix(id, "/")
	if len(id) == 0 {
		http.Error(w, "No Secret ID specified", http.StatusBadRequest)
		return
	}
	resp := types.GetSecretResponse{}
	v, err := filestore.FileStoreConfig.Fs.Read(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Data = v
	jd, err := json.Marshal(&resp)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	if len(resp.Data) == 0 {
		w.WriteHeader(404)
	}
	w.Write(jd)
}
