package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/wolfeidau/docker-registry/uuid"
)

type HttpService struct {
	DataDir string
	router  *mux.Router
}

func (h *HttpService) GetDataDir() string {
	return h.DataDir
}

func NewHttpService(dataDir string) *HttpService {
	h := &HttpService{}
	h.DataDir = dataDir

	r := mux.NewRouter()

	r.PathPrefix("/v1")

	// ping
	r.HandleFunc("/_ping", h.GetPing).Methods("GET")

	// users
	r.HandleFunc("/users", h.GetUsers).Methods("GET")
	r.HandleFunc("/users", h.NewUserSession).Methods("POST") // authentication

	// images
	r.HandleFunc("/images/{imageId}/ancestry", h.GetImageAncestry).Methods("GET")
	r.HandleFunc("/images/{imageId}/layer", h.GetImageLayer).Methods("GET")
	r.HandleFunc("/images/{imageId}/json", h.GetImageJson).Methods("GET")
	r.HandleFunc("/images/{imageId}/{tagName}", h.PutImageResource).Methods("PUT")

	// repositories
	r.HandleFunc("/repositories/{repoName}/tags", h.GetRepositoryTags).Methods("GET")
	r.HandleFunc("/repositories/{repoName}/images", h.GetRepositoryImages).Methods("GET")
	r.HandleFunc("/repositories/{repoName}/tags/{tagName}", h.PutRepositoryTags).Methods("PUT")
	r.HandleFunc("/repositories/{repoName}/images", h.PutRepositoryImages).Methods("PUT")
	r.HandleFunc("/repositories/{repoName}", h.PutRepository).Methods("PUT")

	h.router = r

	return h
}

func (h *HttpService) GetPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Docker-Registry-Version", "0.6.0")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "pong")
}

func (h *HttpService) GetUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (h *HttpService) NewUserSession(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (h *HttpService) GetImageAncestry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["imageId"]
	if paths, err := filepath.Glob(h.GetDataDir() + "/images/" + imageID + "*"); err == nil {
		if len(paths) > 0 {
			image := &Image{paths[0]}
			if out, err := json.Marshal(image.Ancestry()); err == nil {
				writeJSONHeader(w)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, string(out))
				return
			}
		}
	}
	http.NotFound(w, r)
}

func (h *HttpService) GetImageLayer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["imageId"]
	if paths, err := filepath.Glob(h.GetDataDir() + "/images/" + imageID + "*"); err == nil {
		image := &Image{paths[0]}
		file, err := os.Open(image.LayerPath())
		if err == nil {
			w.Header().Add("Content-Type", "application/x-xz")
			w.WriteHeader(http.StatusOK)
			io.Copy(w, file)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func (h *HttpService) GetImageJson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["imageId"]
	if paths, err := filepath.Glob(h.GetDataDir() + "/images/" + imageID + "*"); err == nil {
		if len(paths) > 0 {
			image := &Image{paths[0]}
			file, err := os.Open(image.Dir + "/json")
			if err == nil {
				if file, err := os.Open(image.LayerPath()); err == nil {
					if stat, err := file.Stat(); err == nil {
						w.Header().Add("X-Docker-Size", fmt.Sprintf("%d", stat.Size()))
					}
				}
				w.WriteHeader(http.StatusOK)
				io.Copy(w, file)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (h *HttpService) PutImageResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["imageId"]
	tagName := vars["tagName"]

	err := writeFile(h.GetDataDir()+"/images/"+imageID+"/"+tagName, r.Body)
	if err != nil {
		logger.Error(err.Error())
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *HttpService) GetRepositoryTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repoName := vars["repoName"]
	repo := &Repository{h.GetDataDir() + "/repositories/" + repoName}
	tagsJSON, err := json.Marshal(repo.Tags())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, err.Error())
		return
	}
	writeJSONHeader(w)
	writeEndpointsHeader(w, r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(tagsJSON))
	return
}

func (h *HttpService) GetRepositoryImages(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	repoName := vars["repoName"]

	repo := &Repository{h.GetDataDir() + "/repositories/" + repoName}
	if images, err := repo.Images(); err == nil {
		writeJSONHeader(w)
		writeEndpointsHeader(w, r)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, string(images))
	} else {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	return
}

func (h *HttpService) PutRepositoryTags(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	repoName := vars["repoName"]
	tagName := vars["tagName"]

	path := h.GetDataDir() + "/repositories/" + repoName + "/tags/" + tagName
	err := writeFile(path, r.Body)
	if err != nil {
		logger.Error(err.Error())
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *HttpService) PutRepositoryImages(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	repoName := vars["repoName"]

	repo := &Repository{h.GetDataDir() + "/repositories/" + repoName}
	err := writeFile(repo.ImagesPath(), r.Body)
	if err != nil {
		logger.Error(err.Error())
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *HttpService) PutRepository(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	repoName := vars["repoName"]

	writeJSONHeader(w)
	writeEndpointsHeader(w, r)
	w.Header().Add("WWW-Authenticate", `Token signature=123abc,repository="dynport/test",access=write`)
	w.Header().Add("X-Docker-Token", "token")
	w.WriteHeader(http.StatusOK)
	repo := &Repository{h.GetDataDir() + "/repositories/" + repoName}
	err := writeFile(repo.IndexPath(), r.Body)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (h *HttpService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	uuid := uuid.NewUUID()
	w.Header().Add("X-Request-ID", uuid)
	logger.Info(fmt.Sprintf("%s got request %s %s", uuid, r.Method, r.URL.String()))
	h.router.ServeHTTP(w, r)
	logger.Info(fmt.Sprintf("%s finished request in %.06f", uuid, time.Now().Sub(started).Seconds()))
}

func writeJSONHeader(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
}

func writeEndpointsHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Docker-Endpoints", r.Host)
}
