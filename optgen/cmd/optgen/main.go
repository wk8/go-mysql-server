package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/tools/imports"

	"github.com/dolthub/go-mysql-server/optgen/cmd/support"
)

//go:generate go run main.go -out ../../../sql/expression/function/aggregation/unary_aggs.og.go -pkg aggregation aggs

var (
	errInvalidArgCount     = errors.New("invalid number of arguments")
	errUnrecognizedCommand = errors.New("unrecognized command")
)

var (
	pkg = flag.String("pkg", "aggregation", "package name used in generated files")
	out = flag.String("out", "", "output file name of generated code")
)

const useGoFmt = true

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		exit(errInvalidArgCount)
	}

	var defs support.GenDefs
	var err error
	cmd := args[0]
	switch cmd {
	case "aggs":
		absPath, _ := filepath.Abs(path.Join("..", "source", "unary_aggs.yaml"))
		defs, err = support.DecodeUnaryAggDefs(absPath)
		if err != nil {
			exit(err)
		}
	case "frame":
	case "frameFactory":
	case "framer":
	case "memo":
		absPath, _ := filepath.Abs(path.Join("..", "source", "memo.yaml"))
		defs, err = support.DecodeMemoExprs(absPath)
		if err != nil {
			exit(err)
		}
	default:
		flag.Usage()
		exit(errUnrecognizedCommand)
	}

	var writer io.Writer
	if *out != "" {
		file, err := os.Create(*out)
		if err != nil {
			exit(err)
		}

		defer file.Close()
		writer = file
	} else {
		writer = os.Stderr
	}

	switch cmd {
	case "aggs":
		err = generateAggs(defs, writer)
	case "frame":
		err = generateFrames(nil, writer)
	case "frameFactory":
		err = generateFramesFactory(nil, writer)
	case "framer":
		err = generateFramers(nil, writer)
	case "memo":
		err = generateMemo(defs, writer)
	}

	if err != nil {
		exit(err)
	}
}

// usage is a replacement usage function for the flags package.
func usage() {
	fmt.Fprintf(os.Stderr, "Optgen is a tool for generating optimizer code.\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")

	fmt.Fprintf(os.Stderr, "\toptgen command [flags] sources...\n\n")

	fmt.Fprintf(os.Stderr, "The commands are:\n\n")
	fmt.Fprintf(os.Stderr, "\taggs generates aggregation definitions and functions\n")
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintf(os.Stderr, "Flags:\n")

	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, "\n")
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	os.Exit(2)
}

func generateAggs(defines support.GenDefs, w io.Writer) error {
	var gen support.AggGen
	return generate(defines, w, gen.Generate)
}

func generateFrames(defines support.GenDefs, w io.Writer) error {
	var gen support.FrameGen
	return generate(defines, w, gen.Generate)
}

func generateFramesFactory(defines support.GenDefs, w io.Writer) error {
	var gen support.FrameFactoryGen
	return generate(defines, w, gen.Generate)
}

func generateFramers(defines support.GenDefs, w io.Writer) error {
	var gen support.FramerGen
	return generate(defines, w, gen.Generate)
}

func generateMemo(defines support.GenDefs, w io.Writer) error {
	var gen support.MemoGen
	return generate(defines, w, gen.Generate)
}

func generate(defines support.GenDefs, w io.Writer, genFunc func(defines support.GenDefs, w io.Writer)) error {
	var buf bytes.Buffer

	buf.WriteString("// Code generated by optgen; DO NOT EDIT.\n\n")
	fmt.Fprintf(&buf, "  package %s\n\n", *pkg)

	genFunc(defines, &buf)

	var b []byte
	var err error

	if useGoFmt {
		b, err = imports.Process("github.com/dolthub/go-mysql-server", buf.Bytes(), nil)
		if err != nil {
			// Write out incorrect source for easier debugging.
			b = buf.Bytes()
		}
	} else {
		b = buf.Bytes()
	}

	w.Write(b)
	return err
}
