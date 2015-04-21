package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	Start(os.Getenv("PORT"))
}

func Start(port string) {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			handleGet(res, req)
			//		case "POST":
			//			handlePost(res, req)
		default:
			http.Error(res, "Only GET and POST supported", http.StatusMethodNotAllowed)
		}
	})

	if port == "" {
		port = "5000"
	}
	log.Println("listening on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type fileInfoResponse struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Type    string    `json:"type"`
	Perm    int       `json:"permission"`
	ModTime time.Time `json:"updated_at"`
}

func toFileInfoResponse(fi os.FileInfo) fileInfoResponse {
	return fileInfoResponse{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Type:    string(fi.Mode().String()[0]),
		Perm:    int(fi.Mode().Perm()),
		ModTime: fi.ModTime(),
	}
}

func handleGet(res http.ResponseWriter, req *http.Request) {
	path, err := os.Open(req.URL.Path)
	if err != nil {
		http.Error(res, "Error reading path: "+err.Error(), http.StatusBadRequest)
		return
	}

	pathInfo, err := path.Stat()
	if err != nil {
		http.Error(res, "Error reading path info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if pathInfo.Mode().IsDir() {
		fileInfos, err := ioutil.ReadDir(req.URL.Path)
		if err != nil {
			http.Error(res, "Error reading dir: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fileResponses := []fileInfoResponse{}
		for _, fi := range fileInfos {
			fileResponses = append(fileResponses, toFileInfoResponse(fi))
		}

		res.Header().Set("Content-Type", "application/json") // TODO allow others
		fileResponsesJson, err := json.Marshal(fileResponses)
		if err != nil {
			http.Error(res, "Error converting JSON: "+err.Error(), http.StatusInternalServerError)
			return
		}
		res.Write(fileResponsesJson)
	} else if pathInfo.Mode().IsRegular() {
		io.Copy(res, path)
	} else {
		http.Error(res, "Invalid file type: "+string(pathInfo.Mode().String()[0]), http.StatusBadRequest)
	}
}

//func handlePost(res http.ResponseWriter, req *http.Request) {
//	scheme := "http"
//	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
//		scheme = "https"
//	}
//
//	data, err := ioutil.ReadAll(req.Body)
//	if err != nil {
//		log.Println("post.error.body:", err)
//		http.Error(res, "Error reading request body: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	if err := handleStatusParam(res, req); err != nil {
//		return
//	}
//
//	res.Header().Set("Content-Type", "text/uri-list; charset=utf-8")
//	contentType := req.Header.Get("Content-Type")
//	uri := shaas.CreateUri(scheme, req.Host, req.URL.Path, contentType, data)
//	fmt.Fprintln(res, uri)
//}
