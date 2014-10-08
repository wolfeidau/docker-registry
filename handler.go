package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/wolfeidau/docker-registry/uuid"
)

type HttpRouteHandler func(http.ResponseWriter, *http.Request, [][]string)
type HttpAuthHandler func(http.ResponseWriter, *http.Request) bool

type Mapping struct {
	Method        string
	Regexp        *regexp.Regexp
	Authenticator HttpAuthHandler
	Handler       HttpRouteHandler
}

type Handler struct {
	DataDir, Namespace string
	Auth               UserAuth
	Mappings           []*Mapping
}

func (h *Handler) WriteJsonHeader(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
}

func (h *Handler) WriteEndpointsHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Docker-Endpoints", r.Host)
}

func (h *Handler) GetPing(w http.ResponseWriter, r *http.Request, p [][]string) {
	logger.Infof("GetPing %s", p)

	w.Header().Add("X-Docker-Registry-Version", "0.6.0")
	w.WriteHeader(200)

	fmt.Fprint(w, "pong")
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request, p [][]string) {
	w.WriteHeader(200)
	fmt.Fprint(w, "OK")
}

func (h *Handler) PostUsers(w http.ResponseWriter, r *http.Request, p [][]string) {
	logger.Printf("p %v", p)
	w.WriteHeader(201)
	fmt.Fprint(w, "OK")
}

func (h *Handler) GetRepositoryImages(w http.ResponseWriter, r *http.Request, p [][]string) {

	repo := &Repository{h.DataDir + "/repositories/" + p[0][2]}

	if images, err := repo.Images(); err == nil {
		h.WriteJsonHeader(w)
		h.WriteEndpointsHeader(w, r)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, string(images))
	} else {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusNotFound)
	}

}

func (h *Handler) GetImageAncestry(w http.ResponseWriter, r *http.Request, p [][]string) {
	idPrefix := p[0][2]

	globPath := h.DataDir + "/images/" + idPrefix + "*"
	logger.Printf("GetImageAncestry %s", globPath)

	if paths, err := filepath.Glob(globPath); err == nil {
		if len(paths) > 0 {
			image := &Image{paths[0]}
			if out, err := json.Marshal(image.Ancestry()); err == nil {
				h.WriteJsonHeader(w)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, string(out))
				return
			}
		}
	}

	http.NotFound(w, r)
}

func (h *Handler) GetImageLayer(w http.ResponseWriter, r *http.Request, p [][]string) {
	idPrefix := p[0][2]

	globPath := h.DataDir + "/images/" + idPrefix + "*"
	logger.Printf("GetImageLayer %s", globPath)

	if paths, err := filepath.Glob(globPath); err == nil {
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

func (h *Handler) GetImageJson(w http.ResponseWriter, r *http.Request, p [][]string) {
	idPrefix := p[0][2]

	globPath := h.DataDir + "/images/" + idPrefix + "*"
	logger.Printf("GetImageJson %s", globPath)

	if paths, err := filepath.Glob(globPath); err == nil {
		if len(paths) > 0 {
			image := &Image{paths[0]}
			file, err := os.Open(image.Dir + "/json")
			if err == nil {
				if file, err := os.Open(image.LayerPath()); err == nil {
					if stat, err := file.Stat(); err == nil {
						w.Header().Add("X-Docker-Size", fmt.Sprintf("%d", stat.Size()))
					}
				}
				io.Copy(w, file)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) GetRepositoryTags(w http.ResponseWriter, r *http.Request, p [][]string) {

	repo := &Repository{h.DataDir + "/repositories/" + p[0][2]}
	tagsJson, err := json.Marshal(repo.Tags())

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, err.Error())
		return
	}

	h.WriteJsonHeader(w)
	h.WriteEndpointsHeader(w, r)
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, string(tagsJson))

	logger.Infof("tags %s", string(tagsJson))
}

func (h *Handler) PutImageResource(w http.ResponseWriter, r *http.Request, p [][]string) {
	imageId := p[0][2]
	tagName := p[0][3]

	err := writeFile(h.DataDir+"/images/"+imageId+"/"+tagName, r.Body)

	if err != nil {
		logger.Error(err.Error())
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) PutRepositoryTags(w http.ResponseWriter, r *http.Request, p [][]string) {

	repoName := p[0][2]
	path := h.DataDir + "/repositories/" + repoName + "/tags/" + p[0][3]

	err := writeFile(path, r.Body)
	if err != nil {
		logger.Error(err.Error())
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) PutRepositoryImages(w http.ResponseWriter, r *http.Request, p [][]string) {

	repoName := p[0][2]
	repo := &Repository{h.DataDir + "/repositories/" + repoName}

	err := writeFile(repo.ImagesPath(), r.Body)
	if err != nil {
		logger.Error(err.Error())
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) PutRepository(w http.ResponseWriter, r *http.Request, p [][]string) {

	repoName := p[0][2]

	h.WriteJsonHeader(w)
	h.WriteEndpointsHeader(w, r)
	w.WriteHeader(http.StatusOK)

	repo := &Repository{h.DataDir + "/repositories/" + repoName}

	err := writeFile(repo.IndexPath(), r.Body)

	if err != nil {
		logger.Error(err.Error())
	}
}

func (h *Handler) RepoAuthenticator(w http.ResponseWriter, r *http.Request) bool {

	// if the Authorization header is present
	if _, ok := r.Header["Authorization"]; ok {
		session, err := h.Auth.CheckAuth(r)
		if err != nil {
			w.Header().Add("WWW-Authenticate", `Basic realm="docker-registry"`)
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}

		logger.Infof("session %s %s %d", session.Login, session.Token, session.Status)

		switch session.Status {
		case SessionNew:
			w.Header().Add("WWW-Authenticate", `Token signature=123abc,repository="dynport/test",access=write`)
			w.Header().Add("X-Docker-Token", session.Token)
		}
	}
	return true
}

func (h *Handler) NoopAuthenticator(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (h *Handler) Map(t, re string, authenticator HttpAuthHandler, handler HttpRouteHandler) {
	h.Mappings = append(h.Mappings, &Mapping{t, regexp.MustCompile("/v(\\d+)/" + re), authenticator, handler})
}

func (h *Handler) doHandle(w http.ResponseWriter, r *http.Request) (ok bool) {

	for _, mapping := range h.Mappings {
		if r.Method != mapping.Method {
			continue
		}
		if res := mapping.Regexp.FindAllStringSubmatch(r.URL.String(), -1); len(res) > 0 {
			if ok := mapping.Authenticator(w, r); ok {
				mapping.Handler(w, r, res)
			}
			return true
		}
	}

	return false
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	started := time.Now()
	uuid := uuid.NewUUID()

	w.Header().Add("X-Request-ID", uuid)

	// logger.Info(fmt.Sprintf("%s got request %s %s", uuid, r.Method, r.URL.String()))
	// logger.Info(spew.Sprintf("headers %v", r.Header))

	if ok := h.doHandle(w, r); !ok {
		http.NotFound(w, r)
	}
	l
	ogger.Info(fmt.Sprintf("%s finished request in %.06f", uuid, time.Now().Sub(started).Seconds()))
}

func NewHandler(dataDir, namespace string, auth UserAuth) (handler *Handler) {
	handler = &Handler{DataDir: dataDir, Namespace: namespace, Mappings: make([]*Mapping, 0), Auth: auth}

	// dummies
	handler.Map("GET", "_ping", handler.NoopAuthenticator, handler.GetPing)
	handler.Map("GET", "users", handler.RepoAuthenticator, handler.GetUsers)
	handler.Map("POST", "users/$", handler.NoopAuthenticator, handler.PostUsers)

	// images
	handler.Map("GET", "images/(.*?)/ancestry", handler.RepoAuthenticator, handler.GetImageAncestry)

	handler.Map("GET", "images/(.*?)/layer", handler.RepoAuthenticator, handler.GetImageLayer)
	handler.Map("GET", "images/(.*?)/json", handler.RepoAuthenticator, handler.GetImageJson)
	handler.Map("PUT", "images/(.*?)/(.*)", handler.RepoAuthenticator, handler.PutImageResource)

	// repositories
	handler.Map("GET", fmt.Sprintf("repositories/%s/(.*?)/tags", namespace), handler.RepoAuthenticator, handler.GetRepositoryTags)
	handler.Map("GET", fmt.Sprintf("repositories/%s/(.*?)/images", namespace), handler.RepoAuthenticator, handler.GetRepositoryImages)
	handler.Map("PUT", fmt.Sprintf("repositories/%s/(.*?)/tags/(.*)", namespace), handler.RepoAuthenticator, handler.PutRepositoryTags)
	handler.Map("PUT", fmt.Sprintf("repositories/%s/(.*?)/images", namespace), handler.RepoAuthenticator, handler.PutRepositoryImages)
	handler.Map("PUT", fmt.Sprintf("repositories/%s/(.*?)/$", namespace), handler.RepoAuthenticator, handler.PutRepository)
	return
}
