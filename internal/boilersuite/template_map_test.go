/*
Copyright 2025 The cert-manager Authors.

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

package boilersuite

import (
	"io"
	"log"
	"strings"
	"testing"
)

func TestTemplateFor(t *testing.T) {
	// Create a fake TemplateMap, abusing the text field to identify each template
	tm := TemplateMap{
		"go":         {text: "go"},
		"Dockerfile": {text: "Dockerfile"},
		"bash":       {text: "bash"},
	}
	// List of filenames and the template they should match
	// "" matches the default Template.text returned when no match is found
	cases := map[string]string{
		"main.cpp":                     "",
		"main.go go go.other":         "go",
		"Dockerfile Dockerfile.linux": "Dockerfile",
		"foo.bash":                    "bash",
	}
	for names, res := range cases {
		for _, name := range strings.Split(names, " ") {
			tmpl, _ := tm.TemplateFor(name)
			if res != tmpl.text {
				t.Fatalf("expected file %q to match template %q, matched %q", name, res, tmpl.text)
			}
		}
	}
}
