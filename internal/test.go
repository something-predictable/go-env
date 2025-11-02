package internal

import "strings"

func TestTool() Tool {
	return NewCommandLineTool("go", []string{}, func(files []string) []string {
		sourceFiles := Filter(files, func(f string) bool {
			return strings.HasSuffix(f, ".go") && !strings.HasSuffix(f, "_test.go") || strings.HasPrefix(f, "/testdata/")
		})
		if len(sourceFiles) != 0 {
			return []string{"test", "-race", "./..."}
		}

		testFiles := Filter(files, func(f string) bool {
			return strings.HasSuffix(f, "_test.go")
		})
		if len(testFiles) != 0 {
			return append([]string{"test", "-race"}, testFiles...)
		}

		return nil
	})
}
