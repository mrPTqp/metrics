package generate

import "strings"

func getDefautValue(typeName string) (string, bool) {
	for _, numberType := range numberTypes {
		if strings.Contains(typeName, numberType) || strings.Contains(typeName, "int") {
			return "0", true
		}
	}
	if typeName == "bool" {
		return "false", true
	}
	if typeName == "string" {
		return `""`, true
	}
	return `""`, false
}
