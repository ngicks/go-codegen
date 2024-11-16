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
		"go run ../../../ undgen plain -v --dir ../testtargets --pkg ./...",
		"go run ../../../ undgen validator -v --dir ../testtargets --pkg ./...",
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
				"go run ../../../ undgen patch -v --dir ../testtargets --pkg ./%s ...",
				name,
			),
		)
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
