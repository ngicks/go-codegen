package suffixprinter

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

func CheckGoimports() error {
	_, err := exec.LookPath("goimports")
	if err != nil {
		return fmt.Errorf(
			"exec.LookPath failed. "+
				"If you have not installed goimports, "+
				"install it with \"go install golang.org/x/tools/cmd/goimports@latest\": %w",
			err,
		)
	}
	return nil
}

func ApplyGoimports(ctx context.Context, buf []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "goimports")
	cmd.Stdin = bytes.NewBuffer(buf)
	formatted := new(bytes.Buffer) // pool buf?
	stderr := new(bytes.Buffer)
	cmd.Stdout = formatted
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("goimports failed: input = %s\nerr = %v, msg = %s", buf, err, stderr.Bytes())
	}
	return formatted.Bytes(), nil
}
