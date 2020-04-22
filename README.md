# grobotstxt 

[![Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/travis/jimsmart/grobotstxt/master.svg)](https://travis-ci.org/jimsmart/grobotstxt)
[![codecov](https://codecov.io/gh/jimsmart/grobotstxt/branch/master/graph/badge.svg)](https://codecov.io/gh/jimsmart/grobotstxt)
[![Go Report Card](https://goreportcard.com/badge/github.com/jimsmart/grobotstxt)](https://goreportcard.com/report/github.com/jimsmart/grobotstxt)
[![Used By](https://img.shields.io/sourcegraph/rrc/github.com/jimsmart/grobotstxt.svg)](https://sourcegraph.com/github.com/jimsmart/grobotstxt)
[![Godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/jimsmart/grobotstxt)

grobotstxt is a native Go port of [Google's robots.txt parser and matcher C++ 
library](https://github.com/google/robotstxt).

- Direct function-for-function conversion/port
- Preserves all behaviour of original library
- All 100% of original test suite functionality
- Minor language-specific cleanups

As per Google's original library, we include a small standalone binary, 
for webmasters, that allows testing a single URL and user-agent against 
a robots.txt. Ours is called `icanhasrobot`, and has identical inputs and outputs.

## Installation

### For developers

Get the package (only needed if not using modules):

```bash
$ go get github.com/jimsmart/grobotstxt
```

Use the package within your code (see examples below):

```go
import "github.com/jimsmart/grobotstxt"
```

### For webmasters

Assumes Go is installed, and its environment is already set up.

Fetch the package:

```bash
$ go get github.com/jimsmart/grobotstxt
```

Build and install the standalone binary executable:

```bash
$ go install github.com/jimsmart/grobotstxt/...
```

By default, the resulting binary executable will be `~/go/bin/icanhasrobot` (assuming no customisation has been made to `$GOPATH` or `$GOBIN`).

Use the tool:

```bash
$ icanhasrobot ~/local/path/to/robots.txt YourBot https://example.com/url
user-agent 'YourBot' with URI 'https://example.com/url': ALLOWED
```

If `$GOBIN` is not included in your environment's `$PATH`, use the full path `~/go/bin/icanhasrobot` when invoking the executable.

## Example Code

```go
import "github.com/jimsmart/grobotstxt"

// Coontents of robots.txt file.
robotsTxt := `
    # robots.txt with restricted area

    User-agent: *
    Disallow: /members/*

    Sitemap: http://example.net/sitemap.xml
`

// User-agent of bot.
const userAgent = "FooBot/1.0"

// Target URI.
uri := "http://example.net/members/index.html"


// Is bot allowed to visit this page?
ok := grobotstxt.AgentAllowed(robotsTxt, userAgent, uri)

```

Additionally, one can also extract all Sitemap URIs from a given robots.txt file:

```go
sitemaps := grobotstxt.Sitemaps(robotsTxt)
```

See GoDocs for further information.

## Documentation

GoDocs [https://godoc.org/github.com/jimsmart/grobotstxt](https://godoc.org/github.com/jimsmart/grobotstxt)

## Testing

To run the tests execute `go test` inside the project folder.

For a full coverage report, try:

```bash
$ go test -coverprofile=coverage.out && go tool cover -html=coverage.out
```

## Notes

Parsing of robots.txt files themselves is done exactly as in the production
version of Googlebot, including how percent codes and unicode characters in
patterns are handled. The user must ensure however that the URI passed to the
AgentAllowed and AgentsAllowed functions, or to the URI parameter
of the `icanhasrobot` tool, follows the format specified by RFC3986, since this library
will not perform full normalization of those URI parameters. Only if the URI is
in this format, the matching will be done according to the REP specification.

## License

Package grobotstxt is licensed under the terms of the
Apache license. See [LICENSE](LICENSE) for more information.

## Links

*   Original project
    [Google robots.txt parser and matcher library](https://github.com/google/robotstxt)
