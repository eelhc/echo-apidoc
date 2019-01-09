package apidoc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type docRequest struct {
	Method      string
	path        string
	body        []byte
	HandlerName string
	HandlerDesc string
	Responses   map[int]*docResponse
	requestURI  string
	header      http.Header
	values      url.Values
}

func newDocRequest(httpReq *http.Request, path string) *docRequest {
	reqBody := []byte{}
	if httpReq.Body != nil {
		reqBody, _ = ioutil.ReadAll(httpReq.Body)
		httpReq.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
	}

	return &docRequest{
		Method:     httpReq.Method,
		path:       path,
		body:       reqBody,
		Responses:  map[int]*docResponse{},
		requestURI: httpReq.URL.RequestURI(),
		header:     httpReq.Header,
		values:     httpReq.URL.Query(),
	}
}

func (req *docRequest) Path() string {
	tokns := strings.Split(req.path, "/")
	for i := 0; i < len(tokns); i++ {
		if strings.HasPrefix(tokns[i], ":") {
			tokns[i] = "{" + strings.Replace(tokns[i], ":", "", 1) + "}"
		}
	}
	path := strings.Join(tokns, "/")
	if len(req.values) > 0 {
		valNames := []string{}
		path += "{?"
		for k, _ := range req.values {
			valNames = append(valNames, k)
		}
		path += strings.Join(valNames, ",") + "}"
	}
	return path
}

func (req *docRequest) Parameters() string {
	var params string

	reqPathTokns := strings.Split(req.requestURI, "/")
	pathTokns := strings.Split(req.path, "/")

	for i := 0; i < len(pathTokns); i++ {
		if strings.HasPrefix(pathTokns[i], ":") &&
			len(reqPathTokns) > i {
			params += fmt.Sprintf(
				"+ %s (%s)\n",
				strings.Replace(pathTokns[i], ":", "", 1),
				valueType(reqPathTokns[i]),
			)
		}
	}

	for queryKey, queryVals := range req.values {
		if len(queryVals) < 1 {
			continue
		}
		params += fmt.Sprintf(
			"+ %s: `%s` (%s, optional)\n",
			queryKey, queryVals[0], valueType(queryVals[0]),
		)
	}

	return params
}

func (req *docRequest) Header() string {
	var header string
	for hkey, hval := range req.header {
		if hkey == "User-Agent" ||
			hkey == "Content-Length" {
			continue
		}
		header += hkey + ": " + strings.Join(hval, ", ") + "\n"
	}
	return header
}

func (req *docRequest) Body() string {
	if len(req.body) > 0 {
		return bytesToJSONPretty(req.body, defaultIndent)
	}
	return ""
}

func (req *docRequest) addResponse(resp *docResponse) {
	if _, ok := req.Responses[resp.status]; !ok {
		req.Responses[resp.status] = resp
	}
}
