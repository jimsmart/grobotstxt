package grobotstxt_test

import (
	"fmt"

	"github.com/jimsmart/grobotstxt"
)

func ExampleAgentAllowed() {

	robotsTxt := `
	# robots.txt with restricted area

	User-agent: *
	Disallow: /members/*
`
	ok := grobotstxt.AgentAllowed(robotsTxt, "FooBot/1.0", "http://example.net/members/index.html")
	fmt.Println(ok)

	// Output:
	// false
}

func ExampleSitemaps() {

	robotsTxt := `
	# robots.txt with sitemaps

	User-agent: *
	Disallow: /members/*

	Sitemap: http://example.net/sitemap.xml
	Sitemap: http://example.net/sitemap2.xml
`
	sitemaps := grobotstxt.Sitemaps(robotsTxt)
	fmt.Println(sitemaps)

	// Output:
	// [http://example.net/sitemap.xml http://example.net/sitemap2.xml]
}
