package apidoc

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

type PkgDoc struct {
	AbsPath  string
	PkgName  string
	funcsMap map[string]*doc.Func
}

func NewPkgDoc(absPath string) *PkgDoc {
	pkgs, err := parser.ParseDir(
		token.NewFileSet(),
		absPath,
		func(fi os.FileInfo) bool { return !strings.Contains(fi.Name(), "_test") },
		parser.ParseComments,
	)
	if err != nil || len(pkgs) < 1 {
		return nil
	}

	var pkg *ast.Package
	for _, v := range pkgs {
		pkg = v
		break
	}
	if pkg == nil {
		return nil
	}

	pd := &PkgDoc{
		AbsPath:  absPath,
		PkgName:  pkg.Name,
		funcsMap: map[string]*doc.Func{},
	}

	goDoc := doc.New(pkg, absPath, doc.AllDecls)
	for _, fn := range goDoc.Funcs {
		pd.funcsMap[fn.Name] = fn
	}

	return pd
}

func (p *PkgDoc) FuncDoc(fname string) string {
	if fdoc, ok := p.funcsMap[fname]; ok {
		return fdoc.Doc
	}
	return ""
}

func FuncName(fn interface{}) string {
	t := reflect.TypeOf(fn)
	if t == nil && t.Kind() != reflect.Func {
		return ""
	}

	fnFullPath := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return strings.Replace(filepath.Ext(fnFullPath), ".", "", 1)
}

func bytesToJSONPretty(data []byte, indent string) string {
	var err error
	tmp := map[string]interface{}{}
	if err = jsoniter.Unmarshal(data, &tmp); err != nil {
		if len(data) > 0 {
			return string(data)
		}
		return ""
	}

	jsonData, _ := jsoniter.MarshalIndent(&tmp, "", indent)
	return string(jsonData)
}

func valueType(val string) string {
	if isBool(val) {
		return "boolean"
	} else if isNumber(val) {
		return "number"
	}
	return "string"
}

func isBool(val string) bool {
	lowers := strings.ToLower(val)
	if lowers == "true" ||
		lowers == "false" {
		return true
	}
	return false
}

func isNumber(val string) bool {
	if _, err := strconv.Atoi(val); err != nil {
		return false
	}
	return true
}
