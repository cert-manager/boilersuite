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
	"strings"
	"testing"
	"unicode"
)

const tmplHash = "#header\n#Copyright <<YEAR>> by <<AUTHOR>>\n#footer"
const tmplTrim = "  \n// header\n// Copyright <<YEAR>> by <<AUTHOR>>\n//\n//\n// footer\n\n\n"
const tmplOneline = "/*Copyright <<YEAR>> by <<AUTHOR>>*/"
const tmplNoCopyright = "# <<YEAR>> by <<AUTHOR>>"
const tmplNoYear = "# Copyright <<YEAH>> by <<AUTHOR>>"
const tmplNoAuthor = "# Copyright <<YEAR>> by <<AUTH>>"

func load(t *testing.T, content string, name string) Template {
	tmpl, err := NewTemplate(content, name, "Unittest")
	if err != nil {
		t.Fatalf("failed to load test %q template: %s", name, err)
	}
	return tmpl
}

func TestNewGoodTemplate(t *testing.T) {
	for _, content := range []string{tmplHash, tmplTrim, tmplOneline} {
		tmpl := load(t, content, "sh")
		txt := tmpl.text
		if !strings.Contains(txt, "Unittest") || strings.Contains(txt, "<<AUTHOR>>") {
			t.Fatalf("loaded test template didn't replace author: %q", txt)
		}

		if unicode.IsSpace(rune(txt[0])) || txt[len(txt)-1] != '\n' || unicode.IsSpace(rune(txt[len(txt)-2])) {
			t.Fatalf("loaded test template has bad trim: %q", txt)
		}
	}
}

func TestNewBadTemplate(t *testing.T) {
	for _, content := range []string{tmplNoYear, tmplNoAuthor, tmplNoCopyright} {
		_, err := NewTemplate(content, "sh", "Unittest")
		if err == nil {
			t.Fatalf("should reject test template %q", content)
		}
	}
}
