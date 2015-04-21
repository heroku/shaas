package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func main() {
	Start(os.Getenv("PORT"))
}

func Start(port string) {
	http.HandleFunc("/", handleAny)

	if port == "" {
		port = "5000"
	}
	log.Println("listening on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleAny(res http.ResponseWriter, req *http.Request) {
	log.Printf("method=%s path=%q", req.Method, req.URL.Path)

	path, err := os.Open(req.URL.Path)
	if err != nil {
		if os.IsNotExist(err) {
			handleError(res, req, err, http.StatusNotFound, "File not found")
			return
		}

		handleError(res, req, err, http.StatusBadRequest, "Error reading path")
		return
	}
	defer path.Close()

	pathInfo, err := path.Stat()
	if err != nil {
		handleError(res, req, err, http.StatusInternalServerError, "Error reading path info")
		return
	}

	switch req.Method {
	case "GET":
		handleGet(res, req, path, pathInfo)
	case "POST":
		handlePost(res, req, path, pathInfo)
	default:
		http.Error(res, "Only GET and POST supported", http.StatusMethodNotAllowed)
	}
}

func handleGet(res http.ResponseWriter, req *http.Request, path *os.File, pathInfo os.FileInfo) {
	if pathInfo.Mode().IsDir() {
		fileInfos, err := path.Readdir(0)
		if err != nil {
			handleError(res, req, err, http.StatusInternalServerError, "Error reading directory")
			return
		}

		fileResponses := map[string]fileInfoDetails{}
		for _, fi := range fileInfos {
			fileResponses[fi.Name()] = toFileInfoDetails(fi)
		}

		res.Header().Set("Content-Type", "application/json")
		fileResponsesJson, err := json.Marshal(fileResponses)
		if err != nil {
			handleError(res, req, err, http.StatusInternalServerError, "Error serializing response")
			return
		}
		res.Write(fileResponsesJson)
	} else if pathInfo.Mode().IsRegular() {
		io.Copy(res, path)
	} else {
		handleError(res, req,
			fmt.Errorf("Invalid file type for GET. Only directories and regular files are supported."),
			http.StatusBadRequest, "Invalid file type")
	}
}

func handlePost(res http.ResponseWriter, req *http.Request, path *os.File, pathInfo os.FileInfo) {
	if pathInfo.Mode().IsDir() {
		resFlusherWriter := flushWriterWrapper{res.(flushWriter)}
		cmd := exec.Command("sh")
		cmd.Dir = path.Name()
		cmd.Env = []string{}
		cmd.Stdin = req.Body
		cmd.Stdout = resFlusherWriter
		cmd.Stderr = resFlusherWriter
		if err := cmd.Run(); err != nil {
			handleError(res, req, err, http.StatusInternalServerError, "Error running command")
			return
		}
	} else {
		handleError(res, req,
			fmt.Errorf("Invalid file type for POST. Only directories are supported"),
			http.StatusBadRequest, "Invalid file type")
	}
}

type fileInfoDetails struct {
	Size    int64     `json:"size"`
	Type    string    `json:"type"`
	Perm    int       `json:"permission"`
	ModTime time.Time `json:"updated_at"`
}

func toFileInfoDetails(fi os.FileInfo) fileInfoDetails {
	return fileInfoDetails{
		Size:    fi.Size(),
		Type:    string(fi.Mode().String()[0]),
		Perm:    int(fi.Mode().Perm()),
		ModTime: fi.ModTime(),
	}
}

type stdError struct {
	Message string `json:"message"`
	Cause   error  `json:"cause"`
}

func handleError(res http.ResponseWriter, req *http.Request, err error, httpStatus int, message string) {
	stdError := stdError{
		Message: message,
		Cause:   err,
	}

	log.Printf("method=%s path=%q message=%q cause=%q", req.Method, req.URL.Path, stdError.Message, stdError.Cause)

	stdErrorJson, err := json.Marshal(stdError)
	if err != nil {
		panic(err)
	}

	res.Header().Set("Content-Type", "application/json")
	http.Error(res, string(stdErrorJson), httpStatus)
}

type flushWriter interface {
	Flush()
	Write(buf []byte) (int, error)
}

type flushWriterWrapper struct {
	fw flushWriter
}

func (fww flushWriterWrapper) Write(p []byte) (n int, err error) {
	n, err = fww.fw.Write(p)
	fww.fw.Flush()
	return
}
