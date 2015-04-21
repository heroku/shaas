package main

//import (
//	"net/http"
//	"net/http/httptest"
//	"net/url"
//	"strings"
//	"testing"
//)
//
//const ContentType = "Content-Type"
//
//const base = "http://example.com"
//const text = "hello, world"
//const plainTextType = "text/plain; charset=utf-8"
//const textBase64 = "aGVsbG8sIHdvcmxk"
//
//func uri(contentType string, data string) string {
//	return (base + "?uri=" + url.QueryEscape("data:"+contentType+";base64,"+data))
//}
//
//func TestHandleGet(t *testing.T) {
//	req, _ := http.NewRequest("GET", uri(plainTextType, textBase64), nil)
//	res := httptest.NewRecorder()
//	handleGet(res, req)
//
//	assertEquals(t, res.Body.String(), text)
//	assertEquals(t, res.Header().Get(ContentType), plainTextType)
//	assertEquals(t, res.Code, 200)
//}
//
//func TestHandleGetCustomStatus(t *testing.T) {
//	req, _ := http.NewRequest("GET", uri(plainTextType, textBase64)+"&status=403", nil)
//	res := httptest.NewRecorder()
//	handleGet(res, req)
//
//	assertEquals(t, res.Body.String(), text)
//	assertEquals(t, res.Header().Get(ContentType), plainTextType)
//	assertEquals(t, res.Code, 403)
//}
//
//func TestHandleGetCustomStatusUnparsable(t *testing.T) {
//	req, _ := http.NewRequest("GET", uri(plainTextType, textBase64)+"&status=X", nil)
//	res := httptest.NewRecorder()
//	handleGet(res, req)
//
//	assertEquals(t, res.Body.String(), "Error parsing status code to integer\n")
//	assertEquals(t, res.Header().Get(ContentType), plainTextType)
//	assertEquals(t, res.Code, 400)
//}
//
//func TestHandleGetNoContentType(t *testing.T) {
//	req, _ := http.NewRequest("GET", uri("", textBase64), nil)
//	res := httptest.NewRecorder()
//	handleGet(res, req)
//
//	assertEquals(t, res.Body.String(), text)
//	assertEquals(t, res.Header().Get(ContentType), "")
//	assertEquals(t, res.Code, 200)
//}
//
//func TestHandleGetNoUri(t *testing.T) {
//	req, _ := http.NewRequest("GET", "/path", nil)
//	res := httptest.NewRecorder()
//	handleGet(res, req)
//	assertEquals(t, res.Code, 400)
//}
//
//func TestHandleGetBadUri(t *testing.T) {
//	req, _ := http.NewRequest("GET", "/path?uri=junk", nil)
//	res := httptest.NewRecorder()
//	handleGet(res, req)
//	assertEquals(t, res.Code, 400)
//}
//
//func TestHandlePost(t *testing.T) {
//	req, _ := http.NewRequest("POST", base, strings.NewReader(text))
//	res := httptest.NewRecorder()
//	handlePost(res, req)
//
//	assertEquals(t, res.Body.String(), uri(plainTextType, textBase64)+"\n")
//	assertEquals(t, res.Header().Get(ContentType), "text/uri-list; charset=utf-8")
//	assertEquals(t, res.Code, 200)
//}
//
//func TestHandlePostCustomStatus(t *testing.T) {
//	req, _ := http.NewRequest("POST", base+"?status=403", strings.NewReader(text))
//	res := httptest.NewRecorder()
//	handlePost(res, req)
//
//	assertEquals(t, res.Body.String(), uri(plainTextType, textBase64)+"\n")
//	assertEquals(t, res.Header().Get(ContentType), "text/uri-list; charset=utf-8")
//	assertEquals(t, res.Code, 403)
//}
//
//func assertEquals(t *testing.T, actual interface{}, expected interface{}) {
//	if actual != expected {
//		t.Errorf("Actual: '%v'; Expected: '%v'", actual, expected)
//	}
//}
