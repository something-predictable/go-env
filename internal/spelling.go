package internal

import "slices"

func SpellCheckerTool() Tool {
	return NewCommandLineTool("npm", []string{"--version"}, func(files []string) []string {
		if slices.Contains(files, "dictionary.txt") {
			return []string{"exec", "cspell", "go.mod", "**/*.go"}
		}
		return append([]string{"exec", "cspell"}, files...)
	})
}
