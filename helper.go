package apidoc

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

type PkgDoc struct {
	AbsPath string
	PkgName string
	funcMap map[string]*doc.Func
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
		AbsPath: absPath,
		PkgName: pkg.Name,
		funcMap: map[string]*doc.Func{},
	}

	goDoc := doc.New(pkg, absPath, doc.AllDecls)

	for _, t := range goDoc.Types {
		for _, m := range t.Methods {
			pd.funcMap[t.Name+"-"+m.Name] = m
		}
	}

	for _, fn := range goDoc.Funcs {
		pd.funcMap[fn.Name] = fn
	}

	return pd
}

func (p *PkgDoc) FuncDoc(fname string, tname ...string) string {
	if tname != nil && len(tname) > 0 && tname[0] != "" {
		fname = tname[0] + "-" + fname
	}
	if fdoc, ok := p.funcMap[fname]; ok {
		return fdoc.Doc
	}
	return ""
}

func FuncName(fn interface{}) (packageName, structName, fnName string) {
	t := reflect.TypeOf(fn)
	if t != nil && t.Kind() == reflect.Func {
		path := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
		paths := strings.Split(path, "/")
		fnPaths := strings.Split(paths[len(paths)-1:][0], ".")

		packageName = strings.Join(paths[:len(paths)-1], "/")
		packageName += "/" + fnPaths[0]
		fnName = fnPaths[len(fnPaths)-1:][0]

		if strings.HasSuffix(fnName, "-fm") {
			fnName = strings.Replace(fnName, "-fm", "", 1)
		}

		if len(fnPaths) > 2 {
			structName = fnPaths[1]
		}
	}
	return
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
