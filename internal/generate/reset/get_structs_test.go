package reset

import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testSrc = `
package main

// generate:reset
func main() {
}

// generate:reset
type inter interface {

}

type structNoReset struct {
}

// generate:reset
type structToBeReset1 struct {
}

// generate:reset
// ...
type structToBeReset2 struct {
}

// generate:reset
const cnst = ":)"


`

var wantNames = []string{"structToBeReset1", "structToBeReset2"}

func TestGetStructsToGenerateReset(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", testSrc, parser.ParseComments)
	assert.NoError(t, err, "test src was corrupted")
	structs := getStructsToGenerateReset(file)
	var structNames []string
	fmt.Println(structs)

	for _, strct := range structs {
		structNames = append(structNames, strct.name)
	}

	assert.Equal(t, wantNames, structNames)
}
