package internal

func LinterTool() Tool {
	return NewCommandLineTool("golangci-lint", []string{"version"}, []string{"run"})
}
