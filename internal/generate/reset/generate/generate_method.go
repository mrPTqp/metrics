package generate

import (
	"fmt"
	"go/ast"
	"io"
	"strings"
)

const receiverName = "rcvr"

var numberTypes = []string{"int", "float", "complex"}

const funcDeclarationTemplate = `
func (rcvr *%s) Reset() {
	%s
}
`

func GenerateResetMethod(w io.Writer, structName string, structType *ast.StructType) {
	funcBody := generateResetMethodBody(structType)
	funcDeclaration := fmt.Sprintf(funcDeclarationTemplate, structName, funcBody)
	fmt.Fprintln(w, funcDeclaration)
}

func generateResetMethodBody(structType *ast.StructType) string {
	funcBodySlice := []string{}
	for _, field := range structType.Fields.List {
		for _, fieldName := range field.Names {
			fieldSetString := generateFieldReset(receiverName, fieldName.Name, field.Type)
			funcBodySlice = append(funcBodySlice, fieldSetString)
		}
	}
	return strings.Join(funcBodySlice, "\n")
}

// type notResetable struct{}
// type resetable struct {
// }

// func (r *resetable) Reset() {
// }

// // generate:reset
// type strct struct {
// 	iint, iint2   int
// 	bl            bool
// 	i64           int64
// 	str           string
// 	intP          *int
// 	intPP         **int
// 	strP          *string
// 	slc           []*string
// 	slcP          *[]*string
// 	mp            map[string]string
// 	mpP           *map[string]string
// 	resetable     resetable
// 	notResetable  notResetable
// 	resetableP    *resetable
// 	notResetableP *notResetable
// }
