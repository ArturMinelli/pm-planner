package project

import (
	"fmt"
	"io"
	"os"
)

func Clean(root string, out io.Writer) error {
	root, err := ResolveRoot(root)
	if err != nil {
		return err
	}
	if out == nil {
		out = io.Discard
	}

	for _, target := range CleanTargets(root) {
		if err := os.RemoveAll(target); err != nil {
			return err
		}
		fmt.Fprintf(out, "removed %s\n", target)
	}
	return nil
}
