// Copyright 2022 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"
)

//go:generate go run ./main.go -out ../../setup/setup_data.sg.go -pkg setup setup ../../setup/scripts

var (
	errInvalidArgCount     = errors.New("invalid number of arguments")
	errUnrecognizedCommand = errors.New("unrecognized command")
)

var (
	pkg = flag.String("pkg", "queries", "package name used in generated files")
	out = flag.String("out", "", "output file name of generated code")
)

const useGoFmt = true

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
		exit(errInvalidArgCount)
	}

	cmd := args[0]
	switch cmd {
	case "setup":

	default:
		flag.Usage()
		exit(errUnrecognizedCommand)
	}

	var buf bytes.Buffer
	buf.WriteString("// Code generated by scriptgen; DO NOT EDIT.\n\n")
	fmt.Fprintf(&buf, "  package %s\n\n", *pkg)

	var err error
	switch cmd {
	case "setup":
		err = generateSetup(args[1], &buf)
	}

	toFile(buf, *out)
	if err != nil {
		exit(err)
	}
}

// usage is a replacement usage function for the flags package.
func usage() {
	fmt.Fprintf(os.Stderr, "Scriptgen is a tool for generating optimizer code.\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")

	fmt.Fprintf(os.Stderr, "\tscriptgen command [flags] sources...\n\n")

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

func generateSetup(setupDir string, buf *bytes.Buffer) error {
	return filepath.WalkDir(setupDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		name := strings.Title(strings.TrimSuffix(d.Name(), ".txt"))
		fmt.Fprintf(buf, "var %sData = []SetupScript{{\n", name)

		s, err := setup.NewFileSetup(path)
		if err != nil {
			log.Fatal(err)
		}
		for {
			_, err := s.Next()
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			hasBacktick := strings.Contains(s.Data().Sql, "`")
			if hasBacktick {
				fmt.Fprintf(buf, "  \"%s\",\n", strings.ReplaceAll(s.Data().Sql, "\n", " "))
			} else {
				fmt.Fprintf(buf, "  `%s`,\n", s.Data().Sql)
			}
		}
		fmt.Fprintf(buf, "}}\n\n")
		return nil
	})
}

func toFile(buf bytes.Buffer, out string) error {
	var w io.Writer
	if out != "" {
		file, err := os.Create(out)
		if err != nil {
			exit(err)
		}

		defer file.Close()
		w = file
	} else {
		w = os.Stderr
	}

	var b []byte
	var err error

	if useGoFmt {
		b, err = format.Source(buf.Bytes())
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
