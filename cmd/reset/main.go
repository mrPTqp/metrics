package main

import (
	"github.com/mrPTqp/metrics/internal/generate/reset"
)

func main() {
	root := "."
	reset.Generate(root)
}