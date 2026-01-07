package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
	"unicode"
)

func URLToCamelCase(path string) string {
	path = strings.Trim(path, "/")

	var pathSlice []string

	for index, route := range strings.Split(path, "/") {
		if index == 0 {
			pathSlice = append(pathSlice, route)
		} else {
			runes := []rune(strings.ToLower(route))
			runes[0] = unicode.ToUpper(runes[0])
			pathSlice = append(pathSlice, string(runes))
		}
	}
	return strings.Join(pathSlice, "")
}

type APIMeta struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type tpl struct {
	FuncName string
	Auth     string
}

var (
	baseTpl = template.Must(template.New("baseTpl").Parse(`
func {{.FuncName}}(w http.ResponseWriter, r *http.Request) {
{{.Auth}}
	// some shit
}
`))
	authCheck = `
	auth := r.Header.Get("X-Auth")
	if auth != "100500" {
	    http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
`
	validatorTpl = template.Must(template.New("validatorTpl").Parse(`
`))
)

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import "net/http"`)

	var currentApiName string
	var currentFuncName string

	for _, f := range node.Decls {
		switch g := f.(type) {
		case *ast.GenDecl:
			for _, spec := range g.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					fmt.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
					continue
				}
				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					fmt.Printf("SKIP %T, не структура", currStruct)
					continue
				}
				fmt.Printf("Это структура %v, делаем грязь\n", currType.Name.Name)
				if strings.Contains(strings.ToLower(currType.Name.Name), "api") {
					currentApiName = currType.Name.Name
				}
				if strings.Contains(currType.Name.Name, "Params") {
					apiValidatorText := "apivalidator:"
					for _, field := range currStruct.Fields.List {
						if field.Tag != nil && strings.Contains(field.Tag.Value, apiValidatorText) {
							tag := reflect.StructTag(field.Tag.Value[len(apiValidatorText)+1:])
							fmt.Println(tag)
							// продолжим тут, надо сделать форму
						}
					}
				}
			}
		case *ast.FuncDecl:
			if g.Doc == nil {
				fmt.Printf("SKIP структура %#v не имеет комментов\n", g.Name.Name)
				continue
			}
			needCodegen := false
			var methodJson string
			for _, comment := range g.Doc.List {
				needCodegen = strings.HasPrefix(comment.Text, "// apigen:api ")
				sepIdx := strings.Index(comment.Text, "{")
				if sepIdx == -1 {
					continue
				}
				methodJson = comment.Text[sepIdx:]

				fmt.Printf("SKIP, %#v не имеет префикса apigen:api\n", g.Name.Name)
				continue
			}
			var meta APIMeta
			err := json.Unmarshal([]byte(methodJson), &meta)
			if err != nil {
				panic(err)
			}
			if needCodegen {
				currentFuncName = currentApiName + meta.URL
				var auth string
				if meta.Auth {
					auth = authCheck
				} else {
					auth = ""
				}
				baseTpl.Execute(out, tpl{URLToCamelCase(currentFuncName), auth})
			}
		default:
			fmt.Printf("SKIP %#T is not *ast.GenDecl/*ast.FuncDecl\n", f)
			continue
		}
	}
}
