package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

var short = true

type EnvVar struct {
	Name  string
	Value string
}

// main pretty prints env vars:
//   - sorted by name
//   - cutting long values to begin...end
//   - with padding separating names and values
func main() {
	envp := []EnvVar{}
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		varValue := parts[1]
		if short && len(varValue) > 100 {
			varValue = varValue[:49] + "..." + varValue[len(varValue)-49:]
		}
		envp = append(envp, EnvVar{
			Name:  parts[0],
			Value: varValue,
		})
	}
	sort.Slice(envp, func(i, j int) bool {
		return envp[i].Name < envp[j].Name
	})
	maxLen := 0
	for _, envVar := range envp {
		if len(envVar.Name) > maxLen {
			maxLen = len(envVar.Name)
		}
	}
	pad := strings.Repeat(" ", maxLen)
	for _, envVar := range envp {
		fmt.Printf("%s%s = %s\n", envVar.Name, pad[:maxLen-len(envVar.Name)], envVar.Value)
	}
}
