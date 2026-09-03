/*
Copyright 2023 The cert-manager Authors.

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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cert-manager/boilersuite/internal/boilersuite"
)

const templateDir = "boilerplate-templates"

func Test_Templates(t *testing.T) {
	dirEntries, err := os.ReadDir(templateDir)
	if err != nil {
		t.Fatalf("failed to walk dir %q: %s", templateDir, err)
	}

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(templateDir, entry.Name())

		if !strings.HasPrefix(entry.Name(), "boilerplate") {
			t.Errorf("missing 'boilerplate' prefix on template file %q", path)
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read %q: %s", path, err)
			continue
		}

		if !strings.Contains(string(contents), boilersuite.YearMarker) {
			t.Errorf("couldn't find marker %s in %q", boilersuite.YearMarker, path)
			continue
		}

		if !strings.Contains(string(contents), boilersuite.AuthorMarker) {
			t.Errorf("couldn't find marker %s in %q", boilersuite.AuthorMarker, path)
			continue
		}

		if bytes.Contains(contents, []byte("\r")) {
			t.Errorf("template %q has Windows style line endings. Unix style are required", path)
			continue
		}
	}
}
