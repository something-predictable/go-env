package internal

func TestTool() Tool {
	return NewCommandLineTool("go", []string{}, []string{"test", "./..."})
}
