package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	http.HandleFunc("/", handleAny)
	log.Fatal(http.ListenAndServe(":"+httpPort(), nil))
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

	if strings.HasPrefix(req.Header.Get("Origin"), "ws://") { // TODO: how to detect? scheme is null
		handleWs(res, req, path, pathInfo)
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

		if strings.Contains(req.Header.Get("Accept"), "html") {
			renderDirHtml(res, path.Name(), fileInfos)
		} else {
			renderDirJson(res, fileInfos)
		}
	} else if pathInfo.Mode().IsRegular() {
		io.Copy(res, path)
	} else {
		handleError(res, req, nil,
			http.StatusBadRequest,
			"Invalid file type for GET. Only directories and regular files are supported.")
	}
}

func handlePost(res http.ResponseWriter, req *http.Request, path *os.File, pathInfo os.FileInfo) {
	resFlusherWriter := flushWriterWrapper{res.(flushWriter)}
	execCmd(res, req, path, pathInfo, req.Body, resFlusherWriter, false)
}

func handleWs(res http.ResponseWriter, req *http.Request, path *os.File, pathInfo os.FileInfo) {
	handler := func(ws *websocket.Conn) {
		execCmd(res, req, path, pathInfo, ws, ws, true)
	}

	websocket.Handler(handler).ServeHTTP(res, req)
}

func execCmd(res http.ResponseWriter, req *http.Request, path *os.File, pathInfo os.FileInfo, in io.Reader, out io.Writer, interactive bool) {
	var cmd *exec.Cmd

	if pathInfo.Mode().IsDir() {
		if interactive {
            // TODO: allow interactive session to have a prompt, support heredocs, handle arrow keys, and generally act like a real terminal
			// cmd = exec.Command("bin/pseudo-interactive-bash") // fixes `bash -i` echoing problem, but breaks on heredocs and probably other bash special cases
			cmd = exec.Command("bash", "-i") // double echos (ws + bash -i) and displays arrow character
		} else {
			cmd = exec.Command("bash")
		}
		cmd.Dir = path.Name()
	} else if pathInfo.Mode().IsRegular() && pathInfo.Mode()&0110 != 0 /* is executable for user or group */ {
		cmd = exec.Command(path.Name())
		cmd.Dir = path.Name()[0:strings.LastIndex(path.Name(), string(os.PathSeparator))]
	} else {
		handleError(res, req, nil,
			http.StatusBadRequest,
			"Invalid file type for POST. Only directories and regular executable file are supported")
		return
	}

	cmd.Env = cgiEnv(req)
	if interactive {
		cmd.Env = append(cmd.Env, "PS1=\\[\\033[01;34m\\]\\w\\[\\033[00m\\] \\[\\033[01;32m\\]$ \\[\\033[00m\\]")
	}
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		// error already sent to client. log only
		log.Printf("method=%s path=%q message=%q", req.Method, req.URL.Path, err)
	}
}

func renderDirHtml(res http.ResponseWriter, pathName string, fileInfos []os.FileInfo) {
	res.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(res, "<pre><ul>")
	for _, fi := range fileInfos {
		if !strings.HasSuffix(pathName, "/") {
			pathName += "/"
		}

		label := fi.Name()
		if fi.IsDir() {
			label = "/" + label
		}

		fmt.Fprintf(res, "<li><a href='%s%s'>%s</a></li>", pathName, fi.Name(), html.EscapeString(label))
	}
	fmt.Fprintf(res, "</ul></pre>")
}

func renderDirJson(res http.ResponseWriter, fileInfos []os.FileInfo) {
	res.Header().Set("Content-Type", "application/json")
	fileResponses := map[string]fileInfoDetails{}
	for _, fi := range fileInfos {
		fileResponses[fi.Name()] = toFileInfoDetails(fi)
	}
	fileResponsesJson, err := json.Marshal(fileResponses)
	if err != nil {
		panic(err)
	}
	res.Write(fileResponsesJson)
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

func httpPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "5000"
}

// copied from net/http/cgi/host.go
func cgiEnv(req *http.Request) []string {
	env := []string{
		"SERVER_SOFTWARE=go",
		"SERVER_NAME=" + req.Host,
		"SERVER_PROTOCOL=HTTP/1.1",
		"HTTP_HOST=" + req.Host,
		"GATEWAY_INTERFACE=CGI/1.1",
		"REQUEST_METHOD=" + req.Method,
		"QUERY_STRING=" + req.URL.RawQuery,
		"REQUEST_URI=" + req.URL.RequestURI(),
		"PATH_INFO=" + req.URL.Path,
		"SCRIPT_NAME=" + req.URL.Path,
		"SCRIPT_FILENAME=" + req.URL.Path,
		"REMOTE_ADDR=" + req.RemoteAddr,
		"REMOTE_HOST=" + req.RemoteAddr,
		"SERVER_PORT=" + httpPort(),
	}

	if req.TLS != nil {
		env = append(env, "HTTPS=on")
	}

	for k, v := range req.Header {
		k = strings.Map(upperCaseAndUnderscore, k)
		joinStr := ", "
		if k == "COOKIE" {
			joinStr = "; "
		}
		env = append(env, "HTTP_"+k+"="+strings.Join(v, joinStr))
	}

	if req.ContentLength > 0 {
		env = append(env, fmt.Sprintf("CONTENT_LENGTH=%d", req.ContentLength))
	}
	if ctype := req.Header.Get("Content-Type"); ctype != "" {
		env = append(env, "CONTENT_TYPE="+ctype)
	}

	return env
}

func upperCaseAndUnderscore(r rune) rune {
	switch {
	case r >= 'a' && r <= 'z':
		return r - ('a' - 'A')
	case r == '-':
		return '_'
	case r == '=':
		// Maybe not part of the CGI 'spec' but would mess up
		// the environment in any case, as Go represents the
		// environment as a slice of "key=value" strings.
		return '_'
	}
	// TODO: other transformations in spec or practice?
	return r
}
