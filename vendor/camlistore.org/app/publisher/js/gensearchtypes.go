// +build never

/*
Copyright 2016 The Camlistore Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"os/exec"
)

type matcher func([]byte) bool

// processDoc runs "go doc" for the given type name,
// filtering the output by deleting lines matching any of the deleters,
// and writes the rest of the output to w.
func processDoc(w io.Writer, name string, deleters ...matcher) error {
	var buf, errBuf bytes.Buffer
	cmd := exec.Command("go", "doc", name)
	cmd.Stdout = &buf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go doc %s error: %v, %s", name, err, errBuf.Bytes())
	}
	scanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
NextLine:
	for scanner.Scan() {
		line := scanner.Bytes()
		for _, d := range deleters {
			if d(line) {
				continue NextLine
			}
		}
		var err error
		if _, err = w.Write(line); err == nil {
			_, err = w.Write([]byte{'\n'})
		}
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

func main() {
	var buf bytes.Buffer

	buf.WriteString(`// generated by gensearchtypes.go; DO NOT EDIT

/*
Copyright 2016 The Camlistore Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
    "net/url"
    "time"

    "camlistore.org/pkg/blob"
    "camlistore.org/pkg/types/camtypes"
)

// Duplicating the search pkg types in here - since we only use them for json
// decoding - , instead of importing them through the search package, which would
// bring in more dependencies, and hence a larger js file.
// To give an idea, the generated publisher.js is ~3.5MB, whereas if we instead import
// camlistore.org/pkg/search to use its types instead of the ones below, we grow to
// ~5.7MB.

`)

	P := func(name string, deleters ...matcher) {
		if err := processDoc(&buf, "camlistore.org/pkg/search."+name, deleters...); err != nil {
			log.Fatal(err)
		}
	}

	matchFuncOrDocText := func(line []byte) bool {
		return bytes.HasPrefix(line, []byte("func")) || bytes.HasPrefix(line, []byte("    "))
	}
	matchDescribeRequest := func(line []byte) bool {
		return bytes.HasPrefix(line, []byte("\tRequest *DescribeRequest"))
	}

	for _, task := range []struct {
		Name  string
		Extra []matcher
	}{
		{"SearchResult", nil},
		{"SearchResultBlob", nil},
		{"DescribeResponse", nil},
		{"MetaMap", nil},

		// stripping DescribeRequest from DescribeBlob because it would pull a lot more of search pkg in
		{"DescribedBlob", []matcher{matchDescribeRequest}},

		{"DescribedPermanode", nil},
	} {
		P(task.Name,
			append([]matcher{matchFuncOrDocText}, task.Extra...)...)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("formatting: %v\n%s", err, buf.String())
	}

	flagOut := flag.String("out", "zsearch.go", "output file name (empty or '-' is stdout)")
	flag.Parse()

	out := os.Stdout
	if *flagOut != "" && *flagOut != "-" {
		if out, err = os.Create(*flagOut); err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := out.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}
	if _, err := out.Write(src); err != nil {
		log.Fatal(err)
	}
}
