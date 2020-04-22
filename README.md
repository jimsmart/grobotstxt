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

The original package includes a standalone binary but that has not yet been ported as part of this package.

## Installation
```bash
$ go get github.com/jimsmart/grobotstxt
```

```go
import "github.com/jimsmart/grobotstxt"
```

### Dependencies

- Standard library.
- [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) if you wish to run the tests.

## Examples

```go
import "github.com/jimsmart/grobotstxt"

// Fetched robots.txt file.
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
of the robots tool, follows the format specified by RFC3986, since this library
will not perform full normalization of those URI parameters. Only if the URI is
in this format, the matching will be done according to the REP specification.

## License

Package grobotstxt is licensed under the terms of the
Apache license. See [LICENSE](LICENSE) for more information.

## Links

*   Original project
    [Google robots.txt parser and matcher library](https://github.com/google/robotstxt)
