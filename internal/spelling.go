package internal

import "slices"

func SpellCheckerTool() Tool {
	var args = [8]string{
		"exec",
		"--yes",
		"cspell",
		"--",
		"--config",
		"cspell.json",
		"--quiet",
		"lint",
	}
	return NewCommandLineTool("npm", []string{"--version"}, func(files []string) []string {
		if slices.Contains(files, "dictionary.txt") {
			return append(args[:], "go.mod", "**/*.go")
		}
		return append(args[:], files...)
	}, nil)
}
