package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background())
	defer cancel()
	commands := []string{
		"go run ../../../ undgen plain -v --dir ../testtargets --pkg ./...",
		"go run ../../../ undgen validator -v --dir ../testtargets --pkg ./...",
		"go run ../../../ undgen patch -v --dir ../testtargets --pkg ./all ...",
		"go run ../../../ undgen patch -v --dir ../testtargets --pkg ./deeplynested ...",
		"go run ../../../ undgen patch -v --dir ../testtargets --pkg ./ignored ...",
		"go run ../../../ undgen patch -v --dir ../testtargets --pkg ./typeparam ...",
	}
	for _, command := range commands {
		splitted := strings.Split(command, " ")
		cmd := exec.CommandContext(context.WithoutCancel(ctx), splitted[0], splitted[1:]...)
		piped, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}
		go func() {
			_, _ = io.Copy(os.Stdout, piped)
		}()
		err = cmd.Run()
		if err != nil {
			fmt.Printf("\n\ncommand %q failed: %v\n\n", command, err)
		} else {
			fmt.Printf("\n\ncommands %q succeeded\n\n", command)
		}
	}
}
