// Copyright 2020 Jim Smart
// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file tests the robots.txt parsing and matching code found in robots.cc
// against the current Robots Exclusion Protocol (REP) internet draft (I-D).
// https://tools.ietf.org/html/draft-koster-rep

//

// Converted 2020-04-21, from https://github.com/google/robotstxt/blob/master/robots_test.cc

package grobotstxt_test

import (
	"github.com/jimsmart/grobotstxt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Robots", func() {

	// Line :30
	IsUserAgentAllowed := func(robotstxt, userAgent, url string) bool {
		matcher := grobotstxt.NewRobotsMatcher()
		return matcher.AgentAllowed(robotstxt, userAgent, url)
	}

	EXPECT_TRUE := func(b bool) {
		Expect(b).To(BeTrue())
	}

	EXPECT_FALSE := func(b bool) {
		Expect(b).To(BeFalse())
	}

	// Google-specific: system test.
	It("should GoogleOnly_SystemTest", func() {
		// Line :37
		const robotstxt = "user-agent: FooBot\n" +
			"disallow: /\n"
		// Empty robots.txt: everything allowed.
		EXPECT_TRUE(IsUserAgentAllowed("", "FooBot", ""))

		// Empty user-agent to be matched: everything allowed.
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "", ""))

		// Empty url: implicitly disallowed, see method comment for GetPathParamsQuery
		// in robots.cc.
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", ""))

		// All params empty: same as robots.txt empty, everything allowed.
		EXPECT_TRUE(IsUserAgentAllowed("", "", ""))
	})

	// Rules are colon separated name-value pairs. The following names are
	// provisioned:
	//     user-agent: <value>
	//     allow: <value>
	//     disallow: <value>
	// See REP I-D section "Protocol Definition".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.1
	//
	// Google specific: webmasters sometimes miss the colon separator, but it's
	// obvious what they mean by "disallow /", so we assume the colon if it's
	// missing.
	It("should ID_LineSyntax_Line", func() {
		// Line :65
		const robotstxt_correct = "user-agent: FooBot\n" +
			"disallow: /\n"
		const robotstxt_incorrect = "foo: FooBot\n" +
			"bar: /\n"
		const robotstxt_incorrect_accepted = "user-agent FooBot\n" +
			"disallow /\n"
		const url = "http://foo.bar/x/y"
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_correct, "FooBot", url))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_incorrect, "FooBot", url))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_incorrect_accepted, "FooBot", url))
	})

	// A group is one or more user-agent line followed by rules, and terminated
	// by a another user-agent line. Rules for same user-agents are combined
	// opaquely into one group. Rules outside groups are ignored.
	// See REP I-D section "Protocol Definition".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.1
	It("should ID_LineSyntax_Groups", func() {
		// Line :87
		const robotstxt = "allow: /foo/bar/\n" +
			"\n" +
			"user-agent: FooBot\n" +
			"disallow: /\n" +
			"allow: /x/\n" +
			"user-agent: BarBot\n" +
			"disallow: /\n" +
			"allow: /y/\n" +
			"\n" +
			"\n" +
			"allow: /w/\n" +
			"user-agent: BazBot\n" +
			"\n" +
			"user-agent: FooBot\n" +
			"allow: /z/\n" +
			"disallow: /\n"

		const url_w = "http://foo.bar/w/a"
		const url_x = "http://foo.bar/x/b"
		const url_y = "http://foo.bar/y/c"
		const url_z = "http://foo.bar/z/d"
		const url_foo = "http://foo.bar/foo/bar/"

		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url_x))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url_z))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", url_y))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "BarBot", url_y))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "BarBot", url_w))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "BarBot", url_z))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "BazBot", url_z))

		// // Lines with rules outside groups are ignored.
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", url_foo))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "BarBot", url_foo))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "BazBot", url_foo))
	})

	// REP lines are case insensitive. See REP I-D section "Protocol Definition".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.1
	It("should ID_REPLineNamesCaseInsensitive", func() {
		// Line :128
		const robotstxt_upper = "USER-AGENT: FooBot\n" +
			"ALLOW: /x/\n" +
			"DISALLOW: /\n"
		const robotstxt_lower = "user-agent: FooBot\n" +
			"allow: /x/\n" +
			"disallow: /\n"
		const robotstxt_camel = "uSeR-aGeNt: FooBot\n" +
			"AlLoW: /x/\n" +
			"dIsAlLoW: /\n"
		const url_allowed = "http://foo.bar/x/y"
		const url_disallowed = "http://foo.bar/a/b"

		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_upper, "FooBot", url_allowed))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_lower, "FooBot", url_allowed))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_camel, "FooBot", url_allowed))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_upper, "FooBot", url_disallowed))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_lower, "FooBot", url_disallowed))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_camel, "FooBot", url_disallowed))
	})

	// A user-agent line is expected to contain only [a-zA-Z_-] characters and must
	// not be empty. See REP I-D section "The user-agent line".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.2.1
	It("should ID_VerifyValidUserAgentsToObey", func() {
		// Line :155
		EXPECT_TRUE(grobotstxt.IsValidUserAgentToObey("Foobot"))
		EXPECT_TRUE(grobotstxt.IsValidUserAgentToObey("Foobot-Bar"))
		EXPECT_TRUE(grobotstxt.IsValidUserAgentToObey("Foo_Bar"))

		//   EXPECT_FALSE(grobotstxt.IsValidUserAgentToObey(absl::string_view()));
		EXPECT_FALSE(grobotstxt.IsValidUserAgentToObey(""))
		EXPECT_FALSE(grobotstxt.IsValidUserAgentToObey("ツ"))

		// EXPECT_FALSE(grobotstxt.IsValidUserAgentToObey("Foobot*")) // Allowed by RFC 7231.
		EXPECT_FALSE(grobotstxt.IsValidUserAgentToObey(" Foobot "))
		EXPECT_FALSE(grobotstxt.IsValidUserAgentToObey("Foobot/2.1"))

		EXPECT_FALSE(grobotstxt.IsValidUserAgentToObey("Foobot Bar"))
	})

	// A user-agent line can in fact contain a wider range of characters
	// than the original spec referenced above, according to RFC7231.
	// https://httpwg.org/specs/rfc7231.html#header.user-agent
	It("should ID_VerifyValidUserAgentsToObey", func() {
		EXPECT_TRUE(grobotstxt.IsValidUserAgentToObey("Foo12bot"))
		EXPECT_TRUE(grobotstxt.IsValidUserAgentToObey("Foobot~Bar"))
	})

	// User-agent line values are case insensitive. See REP I-D section "The
	// user-agent line".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.2.1
	It("should ID_UserAgentValueCaseInsensitive", func() {
		// Line :174
		const robotstxt_upper = "User-Agent: FOO BAR\n" +
			"Allow: /x/\n" +
			"Disallow: /\n"
		const robotstxt_lower = "User-Agent: foo bar\n" +
			"Allow: /x/\n" +
			"Disallow: /\n"
		const robotstxt_camel = "User-Agent: FoO bAr\n" +
			"Allow: /x/\n" +
			"Disallow: /\n"
		const url_allowed = "http://foo.bar/x/y"
		const url_disallowed = "http://foo.bar/a/b"

		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_upper, "Foo", url_allowed))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_lower, "Foo", url_allowed))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_camel, "Foo", url_allowed))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_upper, "Foo", url_disallowed))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_lower, "Foo", url_disallowed))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_camel, "Foo", url_disallowed))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_upper, "foo", url_allowed))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_lower, "foo", url_allowed))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_camel, "foo", url_allowed))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_upper, "foo", url_disallowed))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_lower, "foo", url_disallowed))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_camel, "foo", url_disallowed))
	})

	// Google specific: accept user-agent value up to the first space. Space is not
	// allowed in user-agent values, but that doesn't stop webmasters from using
	// them. This is more restrictive than the I-D, since in case of the bad value
	// "Googlebot Images" we'd still obey the rules with "Googlebot".
	// Extends REP I-D section "The user-agent line"
	// https://tools.ietf.org/html/draft-koster-rep#section-2.2.1
	It("should GoogleOnly_AcceptUserAgentUpToFirstSpace", func() {
		// Line :210
		EXPECT_FALSE(grobotstxt.IsValidUserAgentToObey("Foobot Bar"))
		const robotstxt = "User-Agent: *\n" +
			"Disallow: /\n" +
			"User-Agent: Foo Bar\n" +
			"Allow: /x/\n" +
			"Disallow: /\n"
		const url = "http://foo.bar/x/y"

		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "Foo", url))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "Foo Bar", url))
	})

	// If no group matches the user-agent, crawlers must obey the first group with a
	// user-agent line with a "*" value, if present. If no group satisfies either
	// condition, or no groups are present at all, no rules apply.
	// See REP I-D section "The user-agent line".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.2.1
	It("should ID_GlobalGroups_Secondary", func() {
		// Line :229
		const robotstxt_empty = ""
		const robotstxt_global = "user-agent: *\n" +
			"allow: /\n" +
			"user-agent: FooBot\n" +
			"disallow: /\n"
		const robotstxt_only_specific = "user-agent: FooBot\n" +
			"allow: /\n" +
			"user-agent: BarBot\n" +
			"disallow: /\n" +
			"user-agent: BazBot\n" +
			"disallow: /\n"
		const url = "http://foo.bar/x/y"

		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_empty, "FooBot", url))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_global, "FooBot", url))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_global, "BarBot", url))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_only_specific, "QuxBot", url))
	})

	// Matching rules against URIs is case sensitive.
	// See REP I-D section "The Allow and Disallow lines".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.2.2
	It("should ID_AllowDisallow_Value_CaseSensitive", func() {
		// Line :254
		const robotstxt_lowercase_url = "user-agent: FooBot\n" +
			"disallow: /x/\n"
		const robotstxt_uppercase_url = "user-agent: FooBot\n" +
			"disallow: /X/\n"
		const url = "http://foo.bar/x/y"

		EXPECT_FALSE(IsUserAgentAllowed(robotstxt_lowercase_url, "FooBot", url))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt_uppercase_url, "FooBot", url))
	})

	// The most specific match found MUST be used. The most specific match is the
	// match that has the most octets. In case of multiple rules with the same
	// length, the least strict rule must be used.
	// See REP I-D section "The Allow and Disallow lines".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.2.2
	It("should ID_LongestMatch", func() {
		// Line :272

		const url = "http://foo.bar/x/page.html"
		func() {
			const robotstxt = "user-agent: FooBot\n" +
				"disallow: /x/page.html\n" +
				"allow: /x/\n"

			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", url))
		}()
		func() {
			const robotstxt = "user-agent: FooBot\n" +
				"allow: /x/page.html\n" +
				"disallow: /x/\n"

			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/x/"))
		}()
		func() {
			const robotstxt = "user-agent: FooBot\n" +
				"disallow: \n" +
				"allow: \n"
			// In case of equivalent disallow and allow patterns for the same
			// user-agent, allow is used.
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url))
		}()
		func() {
			const robotstxt = "user-agent: FooBot\n" +
				"disallow: /\n" +
				"allow: /\n"
			// In case of equivalent disallow and allow patterns for the same
			// user-agent, allow is used.
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url))
		}()
		func() {
			const url_a = "http://foo.bar/x"
			const url_b = "http://foo.bar/x/"
			const robotstxt = "user-agent: FooBot\n" +
				"disallow: /x\n" +
				"allow: /x/\n"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", url_a))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url_b))
		}()

		func() {
			const robotstxt = "user-agent: FooBot\n" +
				"disallow: /x/page.html\n" +
				"allow: /x/page.html\n"
			// In case of equivalent disallow and allow patterns for the same
			// user-agent, allow is used.
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url))
		}()
		func() {
			const robotstxt = "user-agent: FooBot\n" +
				"allow: /page\n" +
				"disallow: /*.html\n"
			// Longest match wins.
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/page.html"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/page"))
		}()
		func() {
			const robotstxt = "user-agent: FooBot\n" +
				"allow: /x/page.\n" +
				"disallow: /*.html\n"
			// Longest match wins.
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/x/y.html"))
		}()
		func() {
			const robotstxt = "User-agent: *\n" +
				"Disallow: /x/\n" +
				"User-agent: FooBot\n" +
				"Disallow: /y/\n"
			// Most specific group for FooBot allows implicitly /x/page.
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/x/page"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/y/page"))
		}()
	})

	// Octets in the URI and robots.txt paths outside the range of the US-ASCII
	// coded character set, and those in the reserved range defined by RFC3986,
	// MUST be percent-encoded as defined by RFC3986 prior to comparison.
	// See REP I-D section "The Allow and Disallow lines".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.2.2
	//
	// NOTE: It's up to the caller to percent encode a URL before passing it to the
	// parser. Percent encoding URIs in the rules is unnecessary.
	It("should ID_Encoding", func() {
		// Line :372
		const robotstxt = "User-agent: FooBot\n" +
			"Disallow: /\n" +
			"Allow: /foo/bar?qux=taz&baz=http://foo.bar?tar&par\n"
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar?qux=taz&baz=http://foo.bar?tar&par"))
	})

	// 3 byte character: /foo/bar/ツ -> /foo/bar/%E3%83%84
	It("should handle a 3 byte character", func() {
		// Line :385
		const robotstxt = "User-agent: FooBot\n" +
			"Disallow: /\n" +
			"Allow: /foo/bar/ツ\n"
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/%E3%83%84"))
		// // The parser encodes the 3-byte character, but the URL is not %-encoded.
		// EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/ツ"))

		// Departing from Googlebot behaviour:
		// we perform URI normalisation internally now.
		// Due to that, this now returns true here, not false.
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/ツ"))
	})

	// Percent encoded 3 byte character: /foo/bar/%E3%83%84 -> /foo/bar/%E3%83%84
	It("should handle a percent encoded 3 byte character", func() {
		// Line :397
		const robotstxt = "User-agent: FooBot\n" +
			"Disallow: /\n" +
			"Allow: /foo/bar/%E3%83%84\n"
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/%E3%83%84"))
		// EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/ツ"))

		// Departing from Googlebot behaviour:
		// we perform URI normalisation internally now.
		// Due to that, this now returns true here, not false.
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/ツ"))
	})

	// Percent encoded unreserved US-ASCII: /foo/bar/%62%61%7A -> NULL
	// This is illegal according to RFC3986 and while it may work here due to
	// simple string matching, it should not be relied on.
	It("should handle a percent encoded unreserved US-ASCII", func() {
		// Line :410
		const robotstxt = "User-agent: FooBot\n" +
			"Disallow: /\n" +
			"Allow: /foo/bar/%62%61%7A\n"
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/baz"))
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/%62%61%7A"))
	})

	// The REP I-D defines the following characters that have special meaning in
	// robots.txt:
	// # - inline comment.
	// $ - end of pattern.
	// * - any number of characters.
	// See REP I-D section "Special Characters".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.2.3
	It("should ID_SpecialCharacters", func() {
		// Line :429
		func() {
			const robotstxt = "User-agent: FooBot\n" +
				"Disallow: /foo/bar/quz\n" +
				"Allow: /foo/*/qux\n"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/quz"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/quz"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo//quz"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bax/quz"))
		}()
		func() {
			const robotstxt = "User-agent: FooBot\n" +
				"Disallow: /foo/bar$\n" +
				"Allow: /foo/bar/qux\n"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/qux"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar/baz"))
		}()
		func() {
			const robotstxt = "User-agent: FooBot\n" +
				"# Disallow: /\n" +
				"Disallow: /foo/quz#qux\n" +
				"Allow: /\n"
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/bar"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/foo/quz"))
		}()
	})

	// Google-specific: "index.html" (and only that) at the end of a pattern is
	// equivalent to "/".
	It("should GoogleOnly_IndexHTMLisDirectory", func() {
		// Line :473
		const robotstxt = "User-Agent: *\n" +
			"Allow: /allowed-slash/index.html\n" +
			"Disallow: /\n"
		// If index.html is allowed, we interpret this as / being allowed too.
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "foobot", "http://foo.com/allowed-slash/"))
		// Does not exatly match.
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "foobot", "http://foo.com/allowed-slash/index.htm"))
		// Exact match.
		EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "foobot", "http://foo.com/allowed-slash/index.html"))
		EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "foobot", "http://foo.com/anyother-url"))
	})

	// Google-specific: long lines are ignored after 8 * 2083 bytes. See comment in
	// RobotsTxtParser::Parse().
	It("should GoogleOnly_LineTooLong", func() {
		// Line :493
		eolLen := len("\n")
		const maxLineLen = 2083 * 8
		allow := "allow: "
		disallow := "disallow: "

		// Disallow rule pattern matches the URL after being cut off at kMaxLineLen.
		func() {
			robotstxt := "user-agent: FooBot\n"
			longline := "/x/"
			maxLength := maxLineLen - len(longline) - len(disallow) + eolLen
			for len(longline) < maxLength {
				// absl::StrAppend(&longline, "a");
				longline += "a"
			}
			//   absl::StrAppend(&robotstxt, disallow, longline, "/qux\n");
			robotstxt += disallow + longline + "/qux\n"

			// Matches nothing, so URL is allowed.
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fux"))
			// Matches cut off disallow rule.
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar"+longline+"/fux"))
		}()

		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"disallow: /\n"
			longlineA := "/x/"
			longlineB := "/x/"
			maxLength := maxLineLen - len(longlineA) - len(allow) + eolLen
			for len(longlineA) < maxLength {
				longlineA += "a"
				longlineB += "b"
				// absl::StrAppend(&longline_a, "a");
				// absl::StrAppend(&longline_b, "b");
			}
			//   absl::StrAppend(&robotstxt, allow, longline_a, "/qux\n");
			//   absl::StrAppend(&robotstxt, allow, longline_b, "/qux\n");
			robotstxt += allow + longlineA + "/qux\n"
			robotstxt += allow + longlineB + "/qux\n"

			// URL matches the disallow rule.
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/"))
			// Matches the allow rule exactly.
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar"+longlineA+"/qux"))
			// Matches cut off allow rule.
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar"+longlineB+"/fux"))
		}()
	})

	It("should GoogleOnly_DocumentationChecks", func() {
		// Line :545
		// Test documentation from
		// https://developers.google.com/search/reference/robots_txt
		// Section "URL matching based on path values".
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"disallow: /\n" +
					"allow: /fish\n"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/bar"))

			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish.html"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish/salmon.html"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fishheads"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fishheads/yummy.html"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish.html?id=anything"))

			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/Fish.asp"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/catfish"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/?id=fish"))
		}()
		// "/fish*" equals "/fish"
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"disallow: /\n" +
					"allow: /fish*\n"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/bar"))

			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish.html"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish/salmon.html"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fishheads"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fishheads/yummy.html"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish.html?id=anything"))

			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/Fish.bar"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/catfish"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/?id=fish"))
		}()
		// "/fish/" does not equal "/fish"
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"disallow: /\n" +
					"allow: /fish/\n"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/bar"))

			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish/"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish/salmon"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish/?salmon"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish/salmon.html"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish/?id=anything"))

			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish.html"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/Fish/Salmon.html"))
		}()
		// "/*.php"
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"disallow: /\n" +
					"allow: /*.php\n"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/bar"))

			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/filename.php"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/folder/filename.php"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/folder/filename.php?parameters"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar//folder/any.php.file.html"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/filename.php/"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/index?f=filename.php/"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/php/"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/index?php"))

			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/windows.PHP"))
		}()
		// "/*.php$"
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"disallow: /\n" +
					"allow: /*.php$\n"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/bar"))

			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/filename.php"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/folder/filename.php"))

			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/filename.php?parameters"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/filename.php/"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/filename.php5"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/php/"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/filename?php"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/aaaphpaaa"))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar//windows.PHP"))
		}()
		// "/fish*.php"
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"disallow: /\n" +
					"allow: /fish*.php\n"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/bar"))

			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fish.php"))
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/fishheads/catfish.php?parameters"))

			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", "http://foo.bar/Fish.PHP"))
		}()
		// Section "Order of precedence for group-member records".
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"allow: /p\n" +
					"disallow: /\n"
			const url = "http://example.com/page"
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url))
		}()
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"allow: /folder\n" +
					"disallow: /folder\n"
			const url = "http://example.com/folder/page"
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url))
		}()
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"allow: /page\n" +
					"disallow: /*.htm\n"
			const url = "http://example.com/page.htm"
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", url))
		}()
		func() {
			robotstxt :=
				"user-agent: FooBot\n" +
					"allow: /$\n" +
					"disallow: /\n"
			const url = "http://example.com/"
			const url_page = "http://example.com/page.html"
			EXPECT_TRUE(IsUserAgentAllowed(robotstxt, "FooBot", url))
			EXPECT_FALSE(IsUserAgentAllowed(robotstxt, "FooBot", url_page))
		}()
	})

	EXPECT_EQ := func(a, b interface{}) {
		Expect(a).To(Equal(b))
	}

	// Different kinds of line endings are all supported: %x0D / %x0A / %x0D.0A
	It("should ID_LinesNumbersAreCountedCorrectly", func() {
		// Line :796
		report := &robotsStatsReporter{}
		const unixFile = "User-Agent: foo\n" +
			"Allow: /some/path\n" +
			"User-Agent: bar\n" +
			"\n" +
			"\n" +
			"Disallow: /\n"
		grobotstxt.Parse(unixFile, report)
		EXPECT_EQ(4, report.validDirectives)
		EXPECT_EQ(6, report.lastLineSeen)

		const dosFile = "User-Agent: foo\r\n" +
			"Allow: /some/path\r\n" +
			"User-Agent: bar\r\n" +
			"\r\n" +
			"\r\n" +
			"Disallow: /\r\n"
		grobotstxt.Parse(dosFile, report)
		EXPECT_EQ(4, report.validDirectives)
		EXPECT_EQ(6, report.lastLineSeen)

		const macFile = "User-Agent: foo\r" +
			"Allow: /some/path\r" +
			"User-Agent: bar\r" +
			"\r" +
			"\r" +
			"Disallow: /\r"
		grobotstxt.Parse(macFile, report)
		EXPECT_EQ(4, report.validDirectives)
		EXPECT_EQ(6, report.lastLineSeen)

		const noFinalNewline = "User-Agent: foo\n" +
			"Allow: /some/path\n" +
			"User-Agent: bar\n" +
			"\n" +
			"\n" +
			"Disallow: /"
		grobotstxt.Parse(noFinalNewline, report)
		EXPECT_EQ(4, report.validDirectives)
		EXPECT_EQ(6, report.lastLineSeen)

		const mixedFile = "User-Agent: foo\n" +
			"Allow: /some/path\r\n" +
			"User-Agent: bar\n" +
			"\r\n" +
			"\n" +
			"Disallow: /"
		grobotstxt.Parse(mixedFile, report)
		EXPECT_EQ(4, report.validDirectives)
		EXPECT_EQ(6, report.lastLineSeen)
	})

	// BOM characters are unparseable and thus skipped. The rules following the line
	// are used.
	It("should ID_UTF8ByteOrderMarkIsSkipped", func() {
		// Line :856
		report := &robotsStatsReporter{}
		const utf8FileFullBOM = "\xEF\xBB\xBF" +
			"User-Agent: foo\n" +
			"Allow: /AnyValue\n"
		grobotstxt.Parse(utf8FileFullBOM, report)
		EXPECT_EQ(2, report.validDirectives)
		EXPECT_EQ(0, report.unknownDirectives)

		// We allow as well partial ByteOrderMarks.
		const utf8FilePartial2BOM = "\xEF\xBB" +
			"User-Agent: foo\n" +
			"Allow: /AnyValue\n"
		grobotstxt.Parse(utf8FilePartial2BOM, report)
		EXPECT_EQ(2, report.validDirectives)
		EXPECT_EQ(0, report.unknownDirectives)

		const utf8FilePartial1BOM = "\xEF" +
			"User-Agent: foo\n" +
			"Allow: /AnyValue\n"
		grobotstxt.Parse(utf8FilePartial1BOM, report)
		EXPECT_EQ(2, report.validDirectives)
		EXPECT_EQ(0, report.unknownDirectives)

		// If the BOM is not the right sequence, the first line looks like garbage
		// that is skipped (we essentially see "\x11\xBFUser-Agent").
		const utf8FileBrokenBOM = "\xEF\x11\xBF" +
			"User-Agent: foo\n" +
			"Allow: /AnyValue\n"
		grobotstxt.Parse(utf8FileBrokenBOM, report)
		EXPECT_EQ(1, report.validDirectives)
		EXPECT_EQ(1, report.unknownDirectives) // We get one broken line.

		// Some other messed up file: BOMs only valid in the beginning of the file.
		const utf8BOMSomewhereInMiddleOfFile = "User-Agent: foo\n" +
			"\xEF\xBB\xBF" +
			"Allow: /AnyValue\n"
		grobotstxt.Parse(utf8BOMSomewhereInMiddleOfFile, report)
		EXPECT_EQ(1, report.validDirectives)
		EXPECT_EQ(1, report.unknownDirectives)
	})

	// Google specific: the I-D allows any line that crawlers might need, such as
	// sitemaps, which Google supports.
	// See REP I-D section "Other records".
	// https://tools.ietf.org/html/draft-koster-rep#section-2.2.4
	It("should ID_NonStandardLineExample_Sitemap", func() {
		// Line :907
		report := &robotsStatsReporter{}
		func() {
			const sitemap_loc = "http://foo.bar/sitemap.xml"
			robotstxt :=
				"User-Agent: foo\n" +
					"Allow: /some/path\n" +
					"User-Agent: bar\n" +
					"\n" +
					"\n"
			robotstxt += "Sitemap: " + sitemap_loc + "\n"

			grobotstxt.Parse(robotstxt, report)
			EXPECT_EQ(sitemap_loc, report.sitemap)
		}()
		// A sitemap line may appear anywhere in the file.
		func() {
			robotstxt := ""
			const sitemap_loc = "http://foo.bar/sitemap.xml"
			const robotstxt_temp = "User-Agent: foo\n" +
				"Allow: /some/path\n" +
				"User-Agent: bar\n" +
				"\n" +
				"\n"
			robotstxt += "Sitemap: " + sitemap_loc + "\n" + robotstxt_temp

			grobotstxt.Parse(robotstxt, report)
			EXPECT_EQ(sitemap_loc, report.sitemap)
		}()
	})

})

type robotsStatsReporter struct {
	// Line :739
	lastLineSeen      int
	validDirectives   int
	unknownDirectives int
	sitemap           string
}

func (r *robotsStatsReporter) HandleRobotsStart() {
	r.lastLineSeen = 0
	r.validDirectives = 0
	r.unknownDirectives = 0
	r.sitemap = ""
}

func (r *robotsStatsReporter) HandleRobotsEnd() {}

func (r *robotsStatsReporter) HandleUserAgent(lineNum int, value string) {
	r.digest(lineNum)
}

func (r *robotsStatsReporter) HandleAllow(lineNum int, value string) {
	r.digest(lineNum)
}

func (r *robotsStatsReporter) HandleDisallow(lineNum int, value string) {
	r.digest(lineNum)
}

func (r *robotsStatsReporter) HandleSitemap(lineNum int, value string) {
	r.digest(lineNum)
	r.sitemap += value
}

func (r *robotsStatsReporter) HandleUnknownAction(lineNum int, action, value string) {
	r.lastLineSeen = lineNum
	r.unknownDirectives++
}

func (r *robotsStatsReporter) digest(lineNum int) {
	if lineNum < r.lastLineSeen {
		panic("Bad lineNum")
	}
	r.lastLineSeen = lineNum
	r.validDirectives++
}
