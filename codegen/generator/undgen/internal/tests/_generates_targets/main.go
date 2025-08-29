package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
)

var (
	excludes = flag.String("e", "", "")
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background())
	defer cancel()
	commands := []string{
		"go run github.com/ngicks/go-codegen/codegen undgen plain -v --ignore-generated --dir ../testtargets --pkg ./...",
		"go run github.com/ngicks/go-codegen/codegen undgen validator -v --ignore-generated --dir ../testtargets --pkg ./...",
	}

	dirents, err := os.ReadDir("../testtargets")
	if err != nil {
		panic(err)
	}
	for _, dirent := range dirents {
		name := dirent.Name()
		if !dirent.IsDir() || slices.Contains(strings.Split(*excludes, ","), name) {
			continue
		}
		commands = append(
			commands,
			fmt.Sprintf(
				"go run github.com/ngicks/go-codegen/codegen undgen patch -v --ignore-generated --dir ../testtargets --pkg ./%s ...",
				name,
			),
		)
	}

	var errors []error
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
			err = fmt.Errorf("command %q failed: %w", command, err)
			errors = append(errors, err)
			fmt.Printf("%v\n\n", err)
		} else {
			fmt.Printf("\n\ncommands %q succeeded\n\n", command)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\n\nfailed commands:\n")
	}
	for _, err := range errors {
		fmt.Printf("\t%v\n", err)
	}
}
