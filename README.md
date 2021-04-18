# grobotstxt

[![Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Build Status](https://github.com/jimsmart/grobotstxt/actions/workflows/main.yml/badge.svg)](https://github.com/jimsmart/grobotstxt/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/jimsmart/grobotstxt/branch/master/graph/badge.svg)](https://codecov.io/gh/jimsmart/grobotstxt)
[![Go Report Card](https://goreportcard.com/badge/github.com/jimsmart/grobotstxt?cache-buster)](https://goreportcard.com/report/github.com/jimsmart/grobotstxt)
[![Used By](https://img.shields.io/sourcegraph/rrc/github.com/jimsmart/grobotstxt.svg)](https://sourcegraph.com/github.com/jimsmart/grobotstxt)
[![Godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/jimsmart/grobotstxt)

grobotstxt is a native Go port of [Google's robots.txt parser and matcher C++
library](https://github.com/google/robotstxt).

- Direct function-for-function conversion/port
- Preserves all behaviour of original library
- All 100% of original test suite functionality
- Minor language-specific cleanups
- Added a helper to extract Sitemap URIs
- Super simple API

As per Google's original library, we include a small standalone binary executable,
for webmasters, that allows testing a single URL and user-agent against
a robots.txt. Ours is called `icanhasrobot`, and its inputs and outputs
are compatible with the original tool.

## About

Quoting the README from Google's robots.txt parser and matcher repo:

> The Robots Exclusion Protocol (REP) is a standard that enables website owners to control which URLs may be accessed by automated clients (i.e. crawlers) through a simple text file with a specific syntax. It's one of the basic building blocks of the internet as we know it and what allows search engines to operate.
>
> Because the REP was only a de-facto standard for the past 25 years, different implementers implement parsing of robots.txt slightly differently, leading to confusion. This project aims to fix that by releasing the parser that Google uses.
>
> The library is slightly modified (i.e. some internal headers and equivalent symbols) production code used by Googlebot, Google's crawler, to determine which URLs it may access based on rules provided by webmasters in robots.txt files. The library is released open-source to help developers build tools that better reflect Google's robots.txt parsing and matching.

Package grobotstxt aims to be a faithful conversion, from C++ to Go, of Google's robots.txt parser and matcher.

## Quickstart

### Installation

#### For developers

Get the package (only needed if not using modules):

```bash
go get github.com/jimsmart/grobotstxt
```

Use the package within your code (see examples below):

```go
import "github.com/jimsmart/grobotstxt"
```

#### For webmasters

Assumes Go is installed, and its environment is already set up.

Fetch the package:

```bash
go get github.com/jimsmart/grobotstxt
```

Build and install the standalone binary executable:

```bash
go install github.com/jimsmart/grobotstxt/...
```

By default, the resulting binary executable will be `~/go/bin/icanhasrobot` (assuming no customisation has been made to `$GOPATH` or `$GOBIN`).

Use the tool:

```bash
$ icanhasrobot ~/local/path/to/robots.txt YourBot https://example.com/url
user-agent 'YourBot' with URI 'https://example.com/url': ALLOWED
```

If `$GOBIN` is not included in your environment's `$PATH`, use the full path `~/go/bin/icanhasrobot` when invoking the executable.

### Example Code

#### `AgentAllowed`

```go
import "github.com/jimsmart/grobotstxt"

// Contents of robots.txt file.
robotsTxt := `
    # robots.txt with restricted area

    User-agent: *
    Disallow: /members/*

    Sitemap: http://example.net/sitemap.xml
`

// Target URI.
uri := "http://example.net/members/index.html"


// Is bot allowed to visit this page?
ok := grobotstxt.AgentAllowed(robotsTxt, "FooBot/1.0", uri)

```

See also `AgentsAllowed`.

#### `Sitemaps`

Additionally, one can also extract all Sitemap URIs from a given robots.txt file:

```go
sitemaps := grobotstxt.Sitemaps(robotsTxt)
```

## Documentation

GoDocs [https://godoc.org/github.com/jimsmart/grobotstxt](https://godoc.org/github.com/jimsmart/grobotstxt)

## Testing

To run the tests execute `go test` inside the project folder.

For a full coverage report, try:

```bash
go test -coverprofile=coverage.out && go tool cover -html=coverage.out
```

## Notes

The original library required that the URI passed to the
`AgentAllowed` and `AgentsAllowed` functions, or to the URI parameter
of the standalone binary tool, should follow the encoding/escaping format specified by RFC3986, because the library itself does not perform URI normalisation.

In Go, with its native UTF-8 strings, this requirement is not in line with other commonly used APIs, and is therefore somewhat of a surprising/unexpected behaviour to Go developers.

Because of this, the Go API presented here has been ammended to automatically handle UTF-8 URIs, and performs any necessary normalisation internally.

This is the only behavioural change between grobotstxt and the original C++ library.

## License

Like the original library, package grobotstxt is licensed under the terms of the
Apache License, Version 2.0.

See [LICENSE](LICENSE) for more information.

## Links

- Original project:
    [Google robots.txt parser and matcher library](https://github.com/google/robotstxt)

## History

- v1.0.0 (2021-04-18) Tagged as stable.
- v0.2.1 (2021-01-16) Expose more methods of RobotsMatcher as public. Thanks to [anatolym](https://github.com/anatolym)
- v0.2.0 (2020-04-24) Removed requirement for pre-encoded RFC3986 URIs on front-facing API.
- v0.1.0 (2020-04-23) Initial release.
