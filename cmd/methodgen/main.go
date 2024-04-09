package main

import (
	"flag"
	"os"
	"strings"

	"github.com/ngicks/go-iterator/cmd/methodgen/helper"
)

var (
	inputDir = flag.String("i", ".", "input dir.")
	targetTy = flag.String("ty", "def.DeIterator,def.SeIterator", "target type. comma-separated")
	// ignore   = flag.String("ignore", "lenner.go", "ignored filename list. comma-separated")
	outFilename = flag.String("o", "", "out filename. stdout if empty.")
	debug       = flag.Bool("debug", false, "debug prints")
)

func main() {
	flag.Parse()

	var targetTypeNames []string
	if *targetTy == "" {
		targetTypeNames = []string{"def.DeIterator", "def.SeIterator"}
	} else {
		targetTypeNames = strings.Split(*targetTy, ",")
	}

	var p helper.Parser
	p.Debug = *debug

	if err := p.ParseDir(
		*inputDir,
		targetTypeNames,
		[]string{"sizehint"},
	); err != nil {
		panic(err)
	}

	g := helper.Generator{
		PkgName: p.PkgName,
		Imports: helper.GetDefaultImports(),
	}
	g.GenPreamble()

	g.GenSizeHint(p.Info)

	if err := p.ParseDir(
		*inputDir,
		targetTypeNames,
		[]string{"reverse"},
	); err != nil {
		panic(err)
	}

	g.GenReverse(p.Info)

	// open file after parsing.
	var outFile *os.File
	if *outFilename == "" {
		outFile = os.Stdout
	} else {
		f, err := os.Create(*outFilename)
		if err != nil {
			panic(err)
		}
		outFile = f
		defer func() {
			f.Close()
		}()
	}

	if err := g.Write(outFile); err != nil {
		panic(err)
	}
}
