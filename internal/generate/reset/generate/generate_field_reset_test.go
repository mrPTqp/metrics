package generate

import (
	"go/ast"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func normalizeSpaces(s string) string {
	re := regexp.MustCompile(`\s+`)
	s = strings.TrimSpace(s)
	return re.ReplaceAllString(s, " ")
}

func TestGenerateFieldReset(t *testing.T) {
	reciver := "s"
	tests := []struct {
		name      string
		fieldName string
		fieldType ast.Expr
		want      string
	}{
		{
			name:      "Basic type",
			fieldName: "Age",
			fieldType: &ast.Ident{Name: "int"},
			want:      "s.Age = 0",
		},
		{
			name:      "Custom type",
			fieldName: "Obj",
			fieldType: &ast.Ident{Name: "MyType"},
			want: `
var Obj any = s.Obj
if resetter, ok := Obj.(interface{ Reset() }); ok {
	resetter.Reset()
}
`,
		},
		{
			name:      "Pointer to basic type",
			fieldName: "Count",
			fieldType: &ast.StarExpr{
				X: &ast.Ident{Name: "int"},
			},
			want: `
if s.Count != nil {
	*s.Count = 0
}`,
		},
		{
			name:      "Slice",
			fieldName: "Items",
			fieldType: &ast.ArrayType{
				Elt: &ast.Ident{Name: "int"},
			},
			want: "s.Items = (s.Items)[:0]",
		},
		{
			name:      "Map",
			fieldName: "Dict",
			fieldType: &ast.MapType{
				Key:   &ast.Ident{Name: "string"},
				Value: &ast.Ident{Name: "int"},
			},
			want: "clear(s.Dict)",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := generateFieldReset(reciver, testCase.fieldName, testCase.fieldType)
			assert.Equal(
				t,
				normalizeSpaces(testCase.want),
				normalizeSpaces(got),
			)
		})
	}
}
