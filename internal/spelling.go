package internal

func SpellCheckerTool() Tool {
	return NewCommandLineTool("npm", []string{"--version"}, []string{"exec", "cspell", "go.mod", "**/*.go"})
}
