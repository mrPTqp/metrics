package generate

import (
	"fmt"
	"go/ast"
)

func generateFieldReset(rcvr, fieldName string, fieldType ast.Expr) string {
	switch t := fieldType.(type) {
	case *ast.Ident:
		defaultValue, isBasic := getDefautValue(t.Name)
		if isBasic {
			return fmt.Sprintf("%s.%s = %s", rcvr, fieldName, defaultValue)
		} else {
			return fmt.Sprintf(`
var %s any = %s.%s
if resetter, ok := %s.(interface{ Reset() }); ok {
	resetter.Reset()
}
`, fieldName, rcvr, fieldName, fieldName)
		}
	case *ast.StarExpr:
		newRcvr := "*" + rcvr
		fieldReset := generateFieldReset(newRcvr, fieldName, t.X)
		return fmt.Sprintf(`if %s.%s != nil {
			%s
			}`, rcvr, fieldName, fieldReset)

	case *ast.ArrayType:
		return fmt.Sprintf("%s.%s = (%s.%s)[:0]", rcvr, fieldName, rcvr, fieldName)
	case *ast.MapType:
		return fmt.Sprintf("clear(%s.%s)", rcvr, fieldName)
	default:
		return fmt.Sprintf("// type of %s doesn't support ", fieldName)
	}
}
