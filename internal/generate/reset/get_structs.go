package reset

import (
	"go/ast"
	"go/token"
	"strings"
)

type structDecl struct {
	name       string
	structType *ast.StructType
}

func hasGenearateComment(genDecl *ast.GenDecl) bool {
	if genDecl.Doc != nil {
		for _, comment := range genDecl.Doc.List {
			if strings.TrimSpace(comment.Text) == "// generate:reset" {
				return true
			}
		}
	}
	return false
}

func getStructsToGenerateReset(file *ast.File) []structDecl {
	var structs []structDecl
	for _, declaration := range file.Decls {
		genDecl, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if hasGenearateComment(genDecl) {
				structs = append(structs, structDecl{typeSpec.Name.Name, structType})
			}
		}
	}
	return structs
}
