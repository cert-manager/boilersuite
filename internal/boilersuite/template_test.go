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
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode"
)

const tmplHash = "#header\n#Copyright <<YEAR>> by <<AUTHOR>>\n#footer"
const tmplTrim = "  \n// header\n// Copyright <<YEAR>> by <<AUTHOR>>\n//\n// footer\n\n\n"
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

func TestValidate(t *testing.T) {
	tmplSh := load(t, tmplHash, "sh")
	tmplGo := load(t, tmplTrim, "go")
	tmplC := load(t, tmplOneline, "c")
	tests := []struct {
		name  string
		tmpl  Template
		txt   string
		msg   string
		patch string
	}{
		{tmpl: tmplSh,
			name:  "correct",
			txt:   "#header\n#Copyright 2000 by Unittest\n#footer\n\nfoo\n",
			msg:   "",
			patch: "",
		},
		{tmpl: tmplGo,
			name:  "skip check",
			txt:   "package foo\n// +skip_license_check\n",
			msg:   "",
			patch: "",
		},
		{tmpl: tmplSh,
			name:  "do not edit",
			txt:   "#!/bin/sh\n#DO NOT EDIT.\n",
			msg:   "",
			patch: "",
		},
		{tmpl: tmplSh,
			name:  "empty file",
			txt:   "",
			msg:   "missing boilerplate",
			patch: "+#header\n+#Copyright YEAR by Unittest\n+#footer\n",
		},
		{tmpl: tmplSh,
			name:  "minimal text",
			txt:   "foo\nbar\nbaz\n",
			msg:   "missing boilerplate",
			patch: "+#header\n+#Copyright YEAR by Unittest\n+#footer\n+\n foo\n bar\n baz\n",
		},
		{tmpl: tmplSh,
			name:  "leading newlines no header",
			txt:   "\n\n\nfoo\nbar\nbaz\n",
			msg:   "missing boilerplate",
			patch: "+#header\n+#Copyright YEAR by Unittest\n+#footer\n \n-\n-\n foo\n bar\n baz\n",
		},
		{tmpl: tmplC,
			name:  "leading newlines with header",
			txt:   "\n\n/*Copyright 2000 by Unittest*/\n\n\nfoo\n",
			msg:   "incorrect boilerplate",
			patch: "-\n-\n /*Copyright 2000 by Unittest*/\n \n-\n foo\n",
		},
		{tmpl: tmplSh,
			name:  "missing author",
			txt:   "#header\n#Copyright 2000\n#footer\n\nfoo\n",
			msg:   "incorrect boilerplate",
			patch: " #header\n-#Copyright 2000\n+#Copyright 2000 by Unittest\n #footer\n \n foo\n",
		},
		{tmpl: tmplSh,
			name:  "bad year",
			txt:   "#header\n#Copyright something\n#footer\n\nfoo\n",
			msg:   "incorrect boilerplate",
			patch: " #header\n-#Copyright something\n+#Copyright YEAR by Unittest\n #footer\n \n foo\n",
		},
		{tmpl: tmplSh,
			name:  "skip shebang",
			txt:   "#!/bin/sh\n#header\n#Copyright 2000 by Unittest\n#footer\nfoo\n",
			msg:   "incorrect boilerplate",
			patch: " #!/bin/sh\n+\n #header\n #Copyright 2000 by Unittest\n #footer\n+\n foo\n",
		},
		{tmpl: tmplGo,
			name:  "skip gobuild",
			txt:   "// +build linux\n//go:build linux\npackage foo\n",
			msg:   "missing boilerplate",
			patch: " // +build linux\n //go:build linux\n+\n+// header\n+// Copyright YEAR by Unittest\n+//\n+// footer\n+\n package foo\n",
		},
		{tmpl: tmplSh,
			name:  "add template using crlf",
			txt:   "#!/bin/bash\r\n\r\nfoo\r\n",
			msg:   "missing boilerplate",
			patch: " #!/bin/bash\r\n \r\n+#header\r\n+#Copyright YEAR by Unittest\r\n+#footer\r\n+\r\n foo\r\n",
		},
		{tmpl: tmplSh,
			name:  "fix template to use crlf",
			txt:   "#header\n#Copyright 2000 by Unittest\n#footer\n\nfoo\r\n",
			msg:   "incorrect boilerplate",
			patch: "-#header\n-#Copyright 2000 by Unittest\n-#footer\n-\n+#header\r\n+#Copyright 2000 by Unittest\r\n+#footer\r\n+\r\n foo\r\n",
		},
	}
	thisyear := fmt.Sprintf("%d", time.Now().Year())
	for _, c := range tests {
		if err := c.tmpl.validateContent(c.txt, c.name, true); err != nil {
			// Check the diagnostic part
			s := fmt.Sprintf("%s", err)
			msg := s[:strings.Index(s, "\n")]
			if !strings.Contains(msg, c.msg) {
				t.Fatalf("case %q: bad message, want %q\n%s", c.name, c.msg, s)
			}
			// Check the patch part, ignoring the header to simplify test maintenance
			patch := s[(strings.Index(s, "@@\n") + 3):]
			cpatch := strings.ReplaceAll(c.patch, "YEAR", thisyear)
			if patch != cpatch {
				t.Fatalf("case %q: bad patch:\n>>>>expected:\n%s>>>>got:\n%s>>>>", c.name, cpatch, patch)
			}
		} else {
			if c.msg != "" {
				t.Fatalf("case %q: should have failed with %q", c.name, c.msg)
			}
		}
	}
}
