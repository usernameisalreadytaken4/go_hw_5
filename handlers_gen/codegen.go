package main

import (
	"bytes"
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
func {{.FuncName}}(r *http.Request) ({{.StructName}}, error) {
    var data {{.StructName}}
	`))
	template.Execute(out, tpl{FuncName: funcName, StructName: structName})

}

func CloseForm(out *os.File) {
	template := template.Must(template.New("formTpl").Parse(`
	return data, nil
}
	`))
	template.Execute(out, tpl{})

}

type ApiError struct {
	HTTPStatus int
	Err        error
}

func FillJobTemplate(funcName, body string) string {
	var buf bytes.Buffer
	jobTemplate.Execute(&buf, tpl{FuncName: funcName, Body: body})
	return buf.String()
}

func writeServeHTTP(out *os.File, currentApiName string, metaSlice []APIMeta) {
	var cases string
	for _, meta := range metaSlice {
		var buf bytes.Buffer
		caseTpl.Execute(&buf, tpl{
			Body:     strconv.Quote(meta.URL),
			FuncName: ToCamelCase([]string{meta.Entity, meta.Action}),
		})
		cases += string(buf.String())
	}
	serveTpl.Execute(out, tpl{
		ApiName: currentApiName,
		Body:    cases,
	})
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
	if !field.Required {
		return
	}
	template := template.Must(template.New("requiredFieldTpl").Parse(`
	// required
	data.{{.FieldName}} = r.URL.Query().Get("{{.ParamName}}")
	if data.{{.FieldName}} == "" {
		return data, &ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err: errors.New("{{.ParamName}} must me not empty"),
		}
	}
	`))
	template.Execute(out, tpl{
		FieldName: field.Name,
		ParamName: strings.ToLower(field.ParamName),
	})
}

func (field *FieldMeta) EnumCheck(out *os.File) {
	if len(field.Enum) == 0 {
		return
	}
	template := template.Must(template.New("enumFieldTpl").Parse(`
	// enum
	enums := []string{"{{.Body}}"}
	data.{{.FieldName}} = r.URL.Query().Get("{{.ParamName}}")
	if !slices.Contains(enums, data.{{.FieldName}}) {
		return data, &ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err: errors.New("{{.ParamName}} must be one of [{{.Body}}]"),
		}
	}
	`))
	template.Execute(out, tpl{
		FieldName: field.Name,
		ParamName: strings.ToLower(field.ParamName),
		Body:      strings.Join(field.Enum, ", "),
	})
}

func (field *FieldMeta) DefaultCheck(out *os.File) {
	if field.Default == "" {
		return
	}
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
	if limit == 0 {
		return
	}
	template := template.Must(template.New("minMaxFieldTpl").Parse(`
	// min/max
	data.{{.FieldName}} = r.URL.Query().Get("{{.ParamName}}")
	if data.{{.FieldName}} != "" && len(data.{{.FieldName}}) {{.Body}} {{.IntValue}} {
	    return data, &ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err: 		errors.New()"{{.ParamName}} must be {{.Body}}= {{.IntValue}}"),
		}
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
	ApiName     string
	FuncName    string
	StructName  string
	FieldName   string
	ParamName   string
	Body        string
	IntValue    int
	StringValue string
}

var (
	importTpl = template.Must(template.New("importTpl").Parse(`
import (
	{{.Body}}
)
	`))
	serveTpl = template.Must(template.New("serveTpl").Parse(`
func (h *{{.ApiName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    {{.Body}}
    default:
		resp := map[string]interface{}{
			"error": "unknown method",
		}
        w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
    }
}
	`))
	caseTpl = template.Must(template.New("caseTpl").Parse(`
	case {{.Body}}:
		h.{{.FuncName}}(w, r)
	`))
	methodTpl = template.Must(template.New("methodTpl").Parse(`
func (h *{{.ApiName}}) {{.FuncName}}(w http.ResponseWriter, r *http.Request){
{{.Body}}
}
`))
	jobTemplate = template.Must(template.New("jobTpl").Parse(`
	resp := map[string]interface{}{
		"error": "unknown method",
	}
	in, err := {{.Body}}(r)
	if err != nil {
		resp["error"] = err.Error()
		if apiErr, ok := err.(ApiError); ok {
			w.WriteHeader(apiErr.HTTPStatus)
		} else {
			resp["error"] = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}
		jsonRaw, _ := json.Marshal(resp)
		w.Write([]byte(jsonRaw))
		return
	}

	ctx := r.Context()
	data, err := h.{{.FuncName}}(ctx, in)
	if err != nil {
		resp["error"] = err.Error()
		if apiErr, ok := err.(ApiError); ok {
			w.WriteHeader(apiErr.HTTPStatus)
		} else {
			resp["error"] = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}
		jsonRaw, _ := json.Marshal(resp)
		w.Write([]byte(jsonRaw))
		return
	}
	resp["response"] = data

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp) 
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return

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

	var importSlice []string
	for _, name := range []string{"net/http", "slices", "errors", "encoding/json"} {
		importSlice = append(importSlice, strconv.Quote(name))
	}

	importTpl.Execute(out, tpl{
		Body: strings.Join(importSlice, "\n\t"),
	})

	var currentApiName string
	var currentStructName string

	validatorMap := make(map[string]string)
	var metaSlice []APIMeta

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
					if currentApiName != currType.Name.Name && len(metaSlice) > 0 {
						// Новое API подразумевает, что закончили со старым
						writeServeHTTP(out, currentApiName, metaSlice)
						metaSlice = nil
					}
					currentApiName = currType.Name.Name
				} else {
					currentStructName = currType.Name.Name
				}
				if strings.Contains(currType.Name.Name, "Params") {
					apiValidatorText := "apivalidator:"

					fmt.Println("Создаем валидатор для:", currentStructName)
					methodName := strings.TrimSuffix(currentStructName, "Params")
					validatorMap[strings.ToLower(methodName)] = currentStructName + "Validator"
					OpenForm(out, ToCamelCase([]string{currentStructName, "Validator"}), currentStructName)

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
								}

								if strings.Contains(param, "=") {
									paramSlice := strings.Split(param, "=")
									if len(paramSlice) == 2 {
										if paramSlice[0] == "default" {
											fieldMeta.Default = paramSlice[1]

										}
										if paramSlice[0] == "enum" {
											fieldMeta.Enum = strings.Split(paramSlice[1], "|")

										}
										if paramSlice[0] == "max" {
											value, err := strconv.Atoi(paramSlice[1])
											if err != nil {
												fieldMeta.Max = value
											}
										}
										if paramSlice[0] == "min" {
											value, err := strconv.Atoi(paramSlice[1])
											if err != nil {
												fieldMeta.Min = value
											}
										}
										if paramSlice[0] == "paramname" {
											fieldMeta.ParamName = paramSlice[1]

										}
									}
								}

							}
							fieldMeta.RequiredCheck(out)
							fieldMeta.DefaultCheck(out)
							fieldMeta.EnumCheck(out)
							fieldMeta.MaxCheck(out)
							fieldMeta.MinCheck(out)

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
				metaSlice = append(metaSlice, meta)
				var body string
				if meta.Auth {
					body = authCheck
				} else {
					body = ""
				}
				fmt.Println(meta)
				apiSuffix := strings.Trim(currentApiName, "Api")
				if apiSuffix == "My" {
					apiSuffix = ToCamelCase([]string{validatorMap[meta.Action]})
				} else {
					apiSuffix = ToCamelCase([]string{apiSuffix, validatorMap[meta.Action]})
				}
				body += FillJobTemplate(ToCamelCase([]string{meta.Action}), apiSuffix)

				fmt.Println("Создаем хэндлер для:", currentApiName, meta.Entity, meta.Action)
				methodTpl.Execute(out, tpl{
					ApiName:  currentApiName,
					FuncName: ToCamelCase([]string{meta.Entity, meta.Action}),
					Body:     body,
				})
			}
		default:
			fmt.Printf("SKIP %#T is not *ast.GenDecl/*ast.FuncDecl\n", f)
			continue
		}
	}
	// последняя проверка
	if len(metaSlice) > 0 {
		writeServeHTTP(out, currentApiName, metaSlice)
	}
}
