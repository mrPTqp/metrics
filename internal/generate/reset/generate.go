package reset

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mrPTqp/metrics/internal/generate/reset/generate"
)

type pkg struct {
	name string
	path string
}

func getWalkDirFunc(pkgToStructs map[pkg][]structDecl) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}
		pkgName := file.Name.Name
		structs := getStructsToGenerateReset(file)
		pkg := pkg{
			name: pkgName,
			path: filepath.Dir(path),
		}
		if structs != nil {
			pkgToStructs[pkg] = append(pkgToStructs[pkg], structs...)
		}
		return nil
	}

}

func Generate(root string) {
	pathToStructs := make(map[pkg][]structDecl)
	filepath.WalkDir(root, getWalkDirFunc(pathToStructs))
	for pkg, structs := range pathToStructs {
		generatedSource := &bytes.Buffer{}

		pkgDeclaration := fmt.Sprintf("package %s\n\n", pkg.name)
		fmt.Fprintln(generatedSource, pkgDeclaration)
		for _, strct := range structs {
			generate.GenerateResetMethod(generatedSource, strct.name, strct.structType)
		}
		formatedSource, err := format.Source(generatedSource.Bytes())
		if err != nil {
			log.Fatalln(err.Error())
		}

		filename := path.Join(pkg.path, "reset.gen.go")
		generatedFile, err := os.Create(filename)
		if err != nil {
			log.Fatalln(err.Error())
		}
		generatedFile.Write(formatedSource)
		defer generatedFile.Close()

	}
}
