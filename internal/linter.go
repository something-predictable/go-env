package internal

func LinterTool() Tool {
	return NewCommandLineTool("golangci-lint", []string{"version"}, func(_ []string) []string {
		return []string{"run"}
	}, nil)
}
