package reset

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/mrPTqp/metrics/internal/generate/reset/generate"
)

type pkg struct {
	name string
	path string
}

func WalkGoFiles(root string, handler func(path string, file *ast.File) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err 
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		return handler(path, file)
	})
}

func CollectStructs(root string) (map[pkg][]structDecl, error) {
	pkgToStructs := make(map[pkg][]structDecl)

	err := WalkGoFiles(root, func(filePath string, file *ast.File) error {
		structs := getStructsToGenerateReset(file)
		if len(structs) == 0 {
			return nil
		}

		pkg := pkg{
			name: file.Name.Name,
			path: filepath.Dir(filePath),
		}
		pkgToStructs[pkg] = append(pkgToStructs[pkg], structs...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return pkgToStructs, nil
}

func GeneratePackageCode(pkgName string, structs []structDecl) ([]byte, error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "package %s\n\n", pkgName)

	for _, s := range structs {
		generate.GenerateResetMethod(&buf, s.name, s.structType)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format generated code: %w", err)
	}

	return formatted, nil
}

func WriteResetFile(outputPath string, data []byte) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", outputPath, err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to %s: %w", outputPath, err)
	}

	return nil
}

func Generate(root string) {
	structsByPackage, err := CollectStructs(root)
	if err != nil {
		log.Fatalln("Error collecting structs:", err)
	}

	for pkg, structs := range structsByPackage {
		code, err := GeneratePackageCode(pkg.name, structs)
		if err != nil {
			log.Printf("Error generating code for package %s: %v", pkg.name, err)
			continue
		}

		outputPath := filepath.Join(pkg.path, "reset.gen.go")
		err = WriteResetFile(outputPath, code)
		if err != nil {
			log.Printf("Error writing file %s: %v", outputPath, err)
			continue
		}

		log.Printf("Generated %s", outputPath)
	}
}