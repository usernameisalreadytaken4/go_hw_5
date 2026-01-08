package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

func ToCamelCase(stringSlice []string) string {

	var pathSlice []string

	for _, route := range stringSlice {

		runes := []rune(route)
		runes[0] = unicode.ToUpper(runes[0])
		pathSlice = append(pathSlice, string(runes))
	}
	return strings.Join(pathSlice, "")
}

type APIMeta struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`

	Entity string
	Action string
}

func OpenForm(out *os.File, funcName, structName string) {
	template := template.Must(template.New("formTpl").Parse(`
func {{.FuncName}}(r *http.Request) error {
    var data {{.Body}}
	`))
	template.Execute(out, tpl{FuncName: funcName, Body: structName})

}

func CloseForm(out *os.File) {
	template := template.Must(template.New("formTpl").Parse(`
	return nil
}
	`))
	template.Execute(out, tpl{})

}

type FieldMeta struct {
	// * `required` - поле не должно быть пустым (не должно иметь значение по-умолчанию)
	// * `paramname` - если указано - то брать из параметра с этим именем, иначе `lowercase` от имени
	// * `enum` - "одно из"
	// * `default` - если указано и приходит пустое значение (значение по-умолчанию) - устанавливать то что написано указано в `default`
	// * `min` - >= X для типа `int`, для строк `len(str)` >=
	// * `max` - <= X для типа `int`
	Name      string
	Required  bool
	ParamName string
	Enum      []string
	Default   string
	Min       int
	Max       int
}

func (field *FieldMeta) RequiredCheck(out *os.File) {
	template := template.Must(template.New("requiredFieldTpl").Parse(`
	// required
	data.{{.FieldName}} = r.URL.Query().Get("{{.ParamName}}")
	if data.{{.FieldName}} == "" {
		return errors.New("{{.ParamName}} must me not empty")
	}
	`))
	template.Execute(out, tpl{
		FieldName: field.Name,
		ParamName: strings.ToLower(field.ParamName),
	})
}

func (field *FieldMeta) EnumCheck(out *os.File) {
	template := template.Must(template.New("enumFieldTpl").Parse(`
	// enum
	enums := []string{"{{.Body}}"}
	data.{{.FieldName}} = r.URL.Query().Get("{{.ParamName}}")
	if !slices.Contains(enums, data.{{.FieldName}}) {
		return errors.New("{{.ParamName}} must be one of [{{.Body}}]")
	}
	`))
	template.Execute(out, tpl{
		FieldName: field.Name,
		ParamName: strings.ToLower(field.ParamName),
		Body:      strings.Join(field.Enum, ", "),
	})
}

func (field *FieldMeta) DefaultCheck(out *os.File) {
	template := template.Must(template.New("defaultFieldTpl").Parse(`
	// default
	data.{{.FieldName}} = r.URL.Query().Get("{{.ParamName}}")
	if data.{{.FieldName}} == "" {
		data.{{.FieldName}} = "{{.Body}}"
	}
	`))
	template.Execute(out, tpl{
		FieldName: field.Name,
		ParamName: strings.ToLower(field.ParamName),
		Body:      field.Default,
	})
}

func (field *FieldMeta) MinMaxCheck(out *os.File, op string, limit int) {
	template := template.Must(template.New("minMaxFieldTpl").Parse(`
	// min/max
	data.{{.FieldName}} = r.URL.Query().Get("{{.ParamName}}")
	if data.{{.FieldName}} != "" && len(data.{{.FieldName}}) {{.Body}} {{.IntValue}} {
	    return errors.New("{{.ParamName}} must be {{.Body}}= {{.IntValue}}")
	}
	`))

	msg := map[string]string{">": "<=", "<": ">="}[op]

	template.Execute(out, tpl{
		FieldName:   field.Name,
		ParamName:   strings.ToLower(field.ParamName),
		IntValue:    limit,
		StringValue: msg,
		Body:        op,
	})
}

func (field *FieldMeta) MinCheck(out *os.File) {
	field.MinMaxCheck(out, "<", field.Min)
}

func (field *FieldMeta) MaxCheck(out *os.File) {
	field.MinMaxCheck(out, ">", field.Max)
}

type tpl struct {
	FuncName    string
	FieldName   string
	ParamName   string
	Body        string
	IntValue    int
	StringValue string
}

var (
	baseTpl = template.Must(template.New("baseTpl").Parse(`
func {{.FuncName}}(w http.ResponseWriter, r *http.Request) {
{{.Body}}
	// some shit
}
`))
	authCheck = `
	auth := r.Header.Get("X-Auth")
	if auth != "100500" {
	    http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
`
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
	fmt.Fprintln(out, `import "slices"`)
	fmt.Fprintln(out, `import "errors"`)

	var currentApiName string
	var currentStructName string

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
				if strings.Contains(strings.ToLower(currType.Name.Name), "api") {
					currentApiName = currType.Name.Name
				} else {
					currentStructName = currType.Name.Name
				}
				if strings.Contains(currType.Name.Name, "Params") {
					apiValidatorText := "apivalidator:"

					fmt.Println("Создаем валидатор для:", currentStructName)
					OpenForm(out, currentStructName+"Validator", currentStructName)

					for _, field := range currStruct.Fields.List {
						fieldName := field.Names[0].Name
						if field.Tag != nil && strings.Contains(field.Tag.Value, apiValidatorText) {
							param, _ := strconv.Unquote(field.Tag.Value)
							param = param[len(apiValidatorText):]
							param, _ = strconv.Unquote(param)
							params := strings.Split(param, ",")
							fieldMeta := FieldMeta{}
							fieldMeta.Name = fieldName
							fieldMeta.ParamName = fieldName
							for _, param := range params {
								if param == "required" {
									fieldMeta.Required = true
									fieldMeta.RequiredCheck(out)
								}

								if strings.Contains(param, "=") {
									paramSlice := strings.Split(param, "=")
									if len(paramSlice) == 2 {
										if paramSlice[0] == "default" {
											fieldMeta.Default = paramSlice[1]
											fieldMeta.DefaultCheck(out)

										}
										if paramSlice[0] == "enum" {
											fieldMeta.Enum = strings.Split(paramSlice[1], "|")
											fieldMeta.EnumCheck(out)

										}
										if paramSlice[0] == "max" {
											value, err := strconv.Atoi(paramSlice[1])
											if err != nil {
												fieldMeta.Max = value
												fieldMeta.MaxCheck(out)
											}
										}
										if paramSlice[0] == "min" {
											value, err := strconv.Atoi(paramSlice[1])
											if err != nil {
												fieldMeta.Min = value
												fieldMeta.MinCheck(out)
											}
										}
										if paramSlice[0] == "paramname" {
											fieldMeta.ParamName = paramSlice[1]

										}

									}
								}
							}

						}
					}
					CloseForm(out)

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
			path := strings.Trim(meta.URL, "/")
			parts := strings.Split(path, "/")
			if len(parts) >= 2 {
				meta.Entity = parts[0]
				meta.Action = parts[1]
			}

			if needCodegen {
				var auth string
				if meta.Auth {
					auth = authCheck
				} else {
					auth = ""
				}

				fmt.Println("Создаем хэндлер для:", currentApiName, meta.Entity, meta.Action)
				baseTpl.Execute(out, tpl{
					FuncName: ToCamelCase([]string{currentApiName, meta.Entity, meta.Action}),
					Body:     auth,
				})
			}
		default:
			fmt.Printf("SKIP %#T is not *ast.GenDecl/*ast.FuncDecl\n", f)
			continue
		}
	}
}
