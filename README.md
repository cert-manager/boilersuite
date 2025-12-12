# boilersuite

Boilersuite is a tool for checking license boilerplate in cert-manager projects. It was ported to Golang from a Python script originally written for Kubernetes. That Python script was also used [in cert-manager itself](https://github.com/cert-manager/cert-manager/blob/v1.11.0/hack/verify_boilerplate.py).

It uses boilerplate template files to specify how boilerplate should look in a variety of file formats including Makefiles, go files, Dockerfiles and bash scripts.

Since it's written in Go, it's easy to install in projects which already use Go (as is the case for cert-manager). Plus, it doesn't add a dependency on Python allowing for more compact CI environments which don't need to include Python.

## Boilerplate Templates

All templates are in `boilerplate-templates/` and can be changed as needed. The templates are embedded into the built Go binary to ensure portability.

Templates will be interpreted as:

- "suffix type" (e.g. `boilerplate.go.boilertmpl` will be used for `*.go` files)
- "prefix type" (e.g. `boilerplate.Dockerfile.boilertmpl` will be used for `Dockerfile` or `Dockerfile.*`)

All templates can be both types, e.g. `boilerplate.go.boilertmpl` will match `main.go` and `go.mod`. There are built-in
exceptions made for several files, including `go.mod`, `go.sum`, git directories and others.

## Validation Process

We start by loading all templates, checking their validity, replacing the `<<AUTHOR>>` marker, etc.
Then, for each file found in the input directory:

1. Match the file name to one of the loaded templates (skip file if no match)
2. Read file content
3. Skip file if its content marks it as generated or `skip_license_check`
4. Find location and year of existing boilerplate
    * A comment block with a "copyright" string
    * Skiping the shebang or go build constraint if applicable
5. Build the expected content
    * Set year in the template
    * Set lf/crlf newline type
    * Set `expected = original_header + empty_line + boiler_plate + empty_line + original_body`
6. If original doesn't match expected, return an error string and if requested a unified diff

## Running

```console
boilersuite [--skip "paths to skip"] [--author "example"] [--verbose] [--patch] <path-to-validate>
```

|arg               |meaning  |default  |
|------------------|---------|---------|
|`path-to-validate`|Directory or single file to validate|Mandatory|
|`--author`        |Substitute for `<<AUTHOR>>` in the templates|`cert-manager`|
|`--skip`          |Space-separated list of directories that should not be validated (adds to default)|`.git _bin bin node_modules vendor third_party staging`|
|`--verbose`       |Print validated/skipped files and other details|false|
|`--patch`         |Print a unified diff suitable for piping into `patch -p0` to fix bad boilerplate|false|

## Building

```console
make build
```

After this, boilersuite will be available at `_bin/boilersuite`.

## Testing

```console
make test
make validate-local-boilerplate
```
