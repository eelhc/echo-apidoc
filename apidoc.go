package apidoc

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/labstack/echo"
)

const docTmplFmt = `
# Group {{.Group | toTitle}}
{{range $path, $rsc := .Resources}}{{range $index, $req := $rsc}}{{if eq $index 0}}
## {{$req.Path}}
{{if ne $req.Parameters ""}}
+ Parameters

    {{$req.Parameters | alignIndent 1}}
{{end}}{{end}}
### {{$req.HandlerName}} [{{$req.Method}}]
{{$req.HandlerDesc}}

+ Request
{{if ne $req.Header ""}}
    + Header

            {{$req.Header | alignIndent 3}}
{{end}}
{{if ne $req.Body ""}}
    + Body

            {{$req.Body | alignIndent 3}}
{{end}}
{{range $index, $resp := $req.Responses}}
+ Response {{$resp.StatusCode}}
{{if ne $resp.Header ""}}
    + Header

            {{$resp.Header | alignIndent 3}}
{{end}}
{{if ne $resp.Body ""}}
    + Body

            {{$resp.Body | alignIndent 3}}
{{end}}
{{end}}
{{end}}{{end}}
`

var (
	errFailRenderAPIDoc = errors.New("fail to render api doc")
	defaultIndent       = "    "
	defaultTmpl         = template.Must(
		template.New("").Funcs(template.FuncMap{
			"toTitle": strings.Title,
			"alignIndent": func(args ...interface{}) string {
				if len(args) != 2 {
					panic("invalid arguments count")
				}
				cnt, ok := args[0].(int)
				if !ok {
					panic("invalid arguments types: type assertion fail: align indent count")
				}
				input, ok := args[1].(string)
				if !ok {
					panic("invalid arguments types: type assertion fail: input string")
				}
				var alignIndent string
				for i := 0; i < cnt; i++ {
					alignIndent += defaultIndent
				}
				return strings.Replace(input, "\n", "\n"+alignIndent, -1)
			},
		}).Parse(docTmplFmt),
	)
)

type APIDoc struct {
	Resources map[string][]*docRequest
	pkgDoc    *PkgDoc
}

func New() *APIDoc {
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}
	return &APIDoc{
		Resources: map[string][]*docRequest{},
		pkgDoc:    NewPkgDoc(dir),
	}
}

func (doc *APIDoc) Recorder() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := doc.newRequest(c.Request(), c.Path(), next)
			resp := newDocResponse()
			c.Response().Writer = newRespWriterWrapper(c.Response(), resp)
			if err := next(c); err != nil {
				return err
			}
			req.addResponse(resp)
			return nil
		}
	}
}

func (doc *APIDoc) Write(path string) error {
	fi, err := os.OpenFile(
		filepath.Join(doc.pkgDoc.AbsPath, doc.pkgDoc.PkgName+".apib"),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return err
	}
	defer fi.Close()
	_, err = fi.WriteString(doc.Render())
	return err
}

func (doc *APIDoc) Render() string {
	buff := bytes.NewBuffer([]byte{})
	if err := defaultTmpl.Execute(buff, doc); err != nil {
		return fmt.Sprintf("%s\n%s", errFailRenderAPIDoc, err)
	}
	return buff.String()
}

func (doc *APIDoc) Group() string {
	return strings.Title(doc.pkgDoc.PkgName)
}

func (doc *APIDoc) newRequest(req *http.Request, path string, fn echo.HandlerFunc) *docRequest {
	dreq := newDocRequest(req, path)
	requests, ok := doc.Resources[dreq.Path()]
	if !ok {
		requests = []*docRequest{}
		doc.Resources[dreq.Path()] = requests
	}

	var find bool
	for _, r := range requests {
		if dreq.Method == r.Method {
			find = true
			dreq = r
			break
		}
	}

	if !find {
		dreq.HandlerName = FuncName(fn)
		dreq.HandlerDesc = doc.pkgDoc.FuncDoc(dreq.HandlerName)
	}

	doc.Resources[dreq.Path()] = append(requests, dreq)

	return dreq
}
