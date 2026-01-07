package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
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

	for _, f := range node.Decls {
		switch g := f.(type) {
		case *ast.GenDecl:
			fmt.Printf("%T GenDecl (const, var, type, import)\n", g)
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
				fmt.Printf("Это структура, делаем грязь\n")
				// if g.Doc == nil {
				// fmt.Printf("SKIP struct %#v doesnt have comments\n", currType.Name.Name)
				// continue
				// }
			}
		case *ast.FuncDecl:
			fmt.Printf("%T FuncDecl (func Foo(a int) int P{}, func (r *R) Bar() {})\n", g.Name.Name)
			if g.Doc == nil {
				fmt.Printf("SKIP struct %#v doesnt have comments\n", g.Name.Name)
				continue
			}
			needCodegen := false
			for _, comment := range g.Doc.List {
				needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// apigen:api")
			}
			if !needCodegen {
				fmt.Printf("SKIP, %#v не имеет префикса apiget:gen\n", g.Name.Name)
				continue
			}
			fmt.Printf("Это функция, делаем функциональную грязь\n")
		default:
			fmt.Printf("SKIP %#T is not *ast.GenDecl/*ast.FuncDecl\n", f)
			continue
		}
	}
}
