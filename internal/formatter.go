package internal

import (
	"fmt"
	"io"
	"strings"
)

// spell-checker: ignore gofmt

func FormattingTool() Tool {
	return NewCommandLineTool("gofmt", []string{}, func(files []string) []string {
		goFiles := Filter(files, func(f string) bool {
			return strings.HasSuffix(f, ".go")
		})
		return append([]string{"-e", "-l", "-s"}, goFiles...)
	}, func(stdout string, _ int, diagnostics io.Writer) (bool, error) {
		if len(stdout) != 0 {
			for line := range strings.SplitSeq(stdout, "\n") {
				if len(line) == 0 {
					continue
				}
				// spell-checker: ignore malformatted
				_, err := fmt.Fprintf(diagnostics, "%s malformatted\n", line)
				if err != nil {
					return false, err //nolint:wrapcheck
				}
			}
			return false, nil
		}
		return true, nil
	})
}
