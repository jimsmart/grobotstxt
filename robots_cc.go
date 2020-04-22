// Copyright 2020 Jim Smart
// Copyright 1999 Google LLC
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
// -----------------------------------------------------------------------------
// File: robots.cc
// -----------------------------------------------------------------------------
//
// Implements expired internet draft
//   http://www.robotstxt.org/norobots-rfc.txt
// with Google-specific optimizations detailed at
//   https://developers.google.com/search/reference/robots_txt

//

// Converted 2020-04-21, from https://github.com/google/robotstxt/blob/master/robots.cc

package grobotstxt

import (
	"bytes"
	"strings"
	"unicode"
)

// AllowFrequentTypos enables the parsing of common typos in robots.txt, such as DISALOW.
var AllowFrequentTypos = true

// A MatchStrategy defines a strategy for matching individual lines in a
// robots.txt file.
//
// Each Match* method should return a match priority, which is
// interpreted as:
//
//  match priority < 0:  No match.
//
//  match priority == 0: Match, but treat it as if matched an empty pattern.
//
//  match priority > 0:  Match.
//
type MatchStrategy interface {
	MatchAllow(path, pattern string) int
	MatchDisallow(path, pattern string) int
}

// Matches implements robots.txt pattern matching.
//
// Returns true if URI path matches the specified pattern. Pattern is anchored
// at the beginning of path. '$' is special only at the end of pattern.
//
// Since both path and pattern are externally determined (by the webmaster),
// we make sure to have acceptable worst-case performance.
func Matches(path, pattern string) bool {
	// Line :69
	// This method originally belonged to abstract base class RobotsMatchStrategy.
	pathlen := len(path)
	pos := make([]int, pathlen+1)
	var numpos int

	// The pos[] array holds a sorted list of indexes of 'path', with length
	// 'numpos'.  At the start and end of each iteration of the main loop below,
	// the pos[] array will hold a list of the prefixes of the 'path' which can
	// match the current prefix of 'pattern'. If this list is ever empty,
	// return false. If we reach the end of 'pattern' with at least one element
	// in pos[], return true.

	pos[0] = 0
	numpos = 1

	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '$' && (i+1 == len(pattern)) {
			return pos[numpos-1] == pathlen
		}
		if pattern[i] == '*' {
			numpos = pathlen - pos[0] + 1
			for j := 1; j < numpos; j++ {
				pos[j] = pos[j-1] + 1
			}
		} else {
			// Includes '$' when not at end of pattern.
			newnumpos := 0
			for j := 0; j < numpos; j++ {
				if pos[j] < pathlen && path[pos[j]] == pattern[i] {
					pos[newnumpos] = pos[j] + 1
					newnumpos++
				}
			}
			numpos = newnumpos
			if numpos == 0 {
				return false
			}
		}
	}

	return true
}

// getPathParamsQuery extracts path (with params) and query part from the given URI.
// Removes scheme, authority, and fragment. Result always starts with "/".
// Returns "/" if the URI doesn't have a path or is not valid.
func getPathParamsQuery(uri string) string {
	// Line :117

	// path := ""

	// Initial two slashes are ignored.
	searchStart := 0
	if len(uri) >= 2 && uri[0] == '/' && uri[1] == '/' {
		searchStart = 2
	}

	earlyPath := findFirstOf(uri, "/?;", searchStart)
	protocolEnd := find(uri, "://", searchStart)
	if earlyPath != -1 && earlyPath < protocolEnd {
		// If path, param or query starts before ://, :// doesn't indicate protocol.
		protocolEnd = -1
	}
	if protocolEnd == -1 {
		protocolEnd = searchStart
	} else {
		protocolEnd += 3
	}

	pathStart := findFirstOf(uri, "/?;", protocolEnd)
	if pathStart != -1 {
		hashPos := findByte(uri, '#', searchStart)
		if hashPos != -1 && hashPos < pathStart {
			return "/"
		}
		pathEnd := hashPos
		if hashPos == -1 {
			pathEnd = len(uri)
		}
		if uri[pathStart] != '/' {
			return "/" + uri[pathStart:pathEnd]
		}
		return uri[pathStart:pathEnd]
	}

	return "/"
}

func findFirstOf(s, match string, i int) int {
	j := strings.IndexAny(s[i:], match)
	if j != -1 {
		j += i
	}
	return j
}

func find(s, match string, i int) int {
	j := strings.Index(s[i:], match)
	if j != -1 {
		j += i
	}
	return j
}

func findByte(s string, b byte, i int) int {
	j := strings.IndexByte(s[i:], b)
	if j != -1 {
		j += i
	}
	return j
}

// escapePattern is used to canonicalize the allowed/disallowed path patterns.
// UTF-8 multibyte sequences (and other out-of-range ASCII values) are percent-encoded,
// and any existing percent-encoded values have their hex values normalised to uppercase.
//
// For example:
//     /SanJoséSellers ==> /Sanjos%C3%A9Sellers
//     %aa ==> %AA
// If the given path pattern is already adequately escaped,
// the original string is returned unchanged.
func escapePattern(path string) string {
	// Line :156

	needCapitalise := false
	numToEscape := 0

	s := func(i int) byte {
		if i < len(path) {
			return path[i]
		}
		return 0
	}

	// First, scan the buffer to see if changes are needed. Most don't.
	for i := 0; i < len(path); i++ {
		// (a) % escape sequence.
		if path[i] == '%' &&
			isHexDigit(s(i+1)) && isHexDigit(s(i+2)) {
			if isLower(s(i+1)) || isLower(s(i+2)) {
				needCapitalise = true
			}
		} else if path[i] >= 0x80 {
			// (b) needs escaping.
			numToEscape++
		}
		// (c) Already escaped and escape-characters normalized (eg. %2f -> %2F).
	}
	// Return if no changes needed.
	if numToEscape == 0 && !needCapitalise {
		return path
	}

	by := make([]byte, 0, numToEscape*2+len(path))
	out := bytes.NewBuffer(by)
	for i := 0; i < len(path); i++ {
		// (a) Normalize %-escaped sequence (eg. %2f -> %2F).
		if path[i] == '%' &&
			isHexDigit(s(i+1)) && isHexDigit(s(i+2)) {
			out.WriteByte('%')
			i++
			out.WriteByte(toUpper(path[i]))
			i++
			out.WriteByte(toUpper(path[i]))
		} else if path[i] >= 0x80 {
			// (b) %-escape octets whose highest bit is set. These are outside the
			// ASCII range.
			out.WriteByte('%')
			out.WriteByte(hexDigits[(path[i]>>4)&0xf])
			out.WriteByte(hexDigits[path[i]&0xf])
		} else {
			// (c) Normal character, no modification needed.
			out.WriteByte(path[i])
		}
	}
	return string(out.Bytes())
}

const hexDigits = "0123456789ABCDEF"

func isHexDigit(c byte) bool {
	// const hexDigits = "01234567890abcdefABCDEF"
	return '0' <= c && c <= '9' ||
		'a' <= c && c <= 'f' ||
		'A' <= c && c <= 'F'
}

func isLower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func toUpper(c byte) byte {
	return c & (0xFF - 0x20)
}

//

// keyType denotes the type of key in a robots.txt key/value pair.
type keyType int

const (
	unknownKey keyType = iota // unknownKey for unrecognised keys.

	// Generic highlevel fields.
	userAgentKey // userAgentKey for "User-Agent:" keys.
	sitemapKey   // sitemapKey for "Sitemap:" keys.

	// Fields within a user-agent group/section.
	allowKey    // allowKey for "Allow:" keys.
	disallowKey // disallowKey for "Disallow:" keys.
)

//

// A robots.txt has lines of key/value pairs. A ParsedRobotsKey represents
// a key. This class can parse a text-representation (including common typos)
// and represent them as an enumeration which allows for faster processing
// afterwards.
// For unparsable keys, the original string representation is kept.

type parsedKey struct {
	typ keyType
	key string
}

// parseKey parses given key text, returning a suitably initialised parsedKey.
func parseKey(key string) parsedKey {

	// Line :659
	k := parsedKey{}
	if keyIsUserAgent(key) {
		k.typ = userAgentKey
	} else if keyIsAllow(key) {
		k.typ = allowKey
	} else if keyIsDisallow(key) {
		k.typ = disallowKey
	} else if keyIsSitemap(key) {
		k.typ = sitemapKey
	} else {
		k.typ = unknownKey
		k.key = key
	}
	return k
}

// Type returns the type of key.
func (k parsedKey) Type() keyType {
	return k.typ
}

// UnknownKey returns the text of the key for Unknown key types.
// For all other key types it returns an empty string.
func (k parsedKey) UnknownKey() string {
	// Line :675
	return k.key
}

func keyIsUserAgent(key string) bool {
	// Line :680
	return startsWithIgnoreCase(key, "user-agent") ||
		(AllowFrequentTypos && (startsWithIgnoreCase(key, "useragent") ||
			startsWithIgnoreCase(key, "user agent")))
}

func keyIsAllow(key string) bool {
	// Line :687
	return startsWithIgnoreCase(key, "allow")
}

func keyIsDisallow(key string) bool {
	// Line :691
	return startsWithIgnoreCase(key, "disallow") ||
		(AllowFrequentTypos && (startsWithIgnoreCase(key, "dissallow") ||
			startsWithIgnoreCase(key, "dissalow") ||
			startsWithIgnoreCase(key, "disalow") ||
			startsWithIgnoreCase(key, "diasllow") ||
			startsWithIgnoreCase(key, "disallaw")))
}

func keyIsSitemap(key string) bool {
	// Line :701
	return startsWithIgnoreCase(key, "sitemap") ||
		startsWithIgnoreCase(key, "site-map")
}

func startsWithIgnoreCase(x, y string) bool {
	return strings.HasPrefix(strings.ToLower(x), strings.ToLower(y))
}

//

func emitKeyValueToHandler(line int, key parsedKey, value string, handler ParseHandler) {
	// Line :262
	switch key.Type() {
	case userAgentKey:
		handler.HandleUserAgent(line, value)
	case allowKey:
		handler.HandleAllow(line, value)
	case disallowKey:
		handler.HandleDisallow(line, value)
	case sitemapKey:
		handler.HandleSitemap(line, value)
	case unknownKey:
		handler.HandleUnknownAction(line, key.UnknownKey(), value)
	}
}

//

type Parser struct {
	// Line :278
	robotsBody string // TODO Should be []byte
	handler    ParseHandler
}

func NewParser(robotsBody string, handler ParseHandler) *Parser {
	// Line :282
	p := Parser{
		robotsBody: robotsBody,
		handler:    handler,
	}
	return &p
}

func (p *Parser) needEscapeValueForKey(key parsedKey) bool {
	// Line :300
	switch key.Type() {
	case userAgentKey, sitemapKey:
		return false
	default:
		return true
	}
}

// parseKeyAndValue attempts to parse a line of robots.txt into a key/value pair.
//
// On success, the parsed key and value, and true, are returned. If parsing is
// unsuccessful, parseKeyAndValue returns two empty strings and false.
func (p *Parser) parseKeyAndValue(line string) (string, string, bool) {
	// Line :317
	// Remove comments from the current robots.txt line.
	comment := strings.IndexByte(line, '#')
	if comment != -1 {
		line = line[:comment]
	}
	line = strings.TrimSpace(line)

	// Rules must match the following pattern:
	//   <key>[ \t]*:[ \t]*<value>
	sep := strings.IndexByte(line, ':')
	if sep == -1 {
		// Google-specific optimization: some people forget the colon, so we need to
		// accept whitespace in its stead.
		white := " \t"
		sep = strings.IndexAny(line, white)
		if sep != -1 {
			val := strings.TrimSpace(line[sep:])
			if len(val) == 0 { // since we dropped trailing whitespace above.
				panic("Syntax error") // TODO Cleanup panics.
			}
			if strings.IndexAny(val, white) != -1 {
				// We only accept whitespace as a separator if there are exactly two
				// sequences of non-whitespace characters.  If we get here, there were
				// more than 2 such sequences since we stripped trailing whitespace
				// above.
				return "", "", false
			}
		}
	}

	if sep == -1 {
		return "", "", false // Couldn't find a separator.
	}

	key := line[:sep]            // Key starts at beginning of line, and stops at the separator.
	key = strings.TrimSpace(key) // Get rid of any trailing whitespace.

	if len(key) == 0 {
		return "", "", false
	}

	value := line[sep+1:]            // Value starts after the separator.
	value = strings.TrimSpace(value) // Get rid of any leading whitespace.
	return key, value, true
}

func (p *Parser) parseAndEmitLine(currentLine int, line string) {
	// Line :362
	stringKey, value, ok := p.parseKeyAndValue(line)
	if !ok {
		return
	}

	key := parseKey(stringKey)
	if p.needEscapeValueForKey(key) {
		value = escapePattern(value)
	}
	emitKeyValueToHandler(currentLine, key, value, p.handler)
}

// Parse body of this Parser's robots.txt and emit parse callbacks. This will accept
// typical typos found in robots.txt, such as 'disalow'.
//
// Note, this function will accept all kind of input but will skip
// everything that does not look like a robots directive.
func (p *Parser) Parse() {
	// TODO see line :381
	// UTF-8 byte order marks.
	utfBOM := []byte{0xEF, 0xBB, 0xBF}

	// Certain browsers limit the URL length to 2083 bytes. In a robots.txt, it's
	// fairly safe to assume any valid line isn't going to be more than many times
	// that max url length of 2KB. We want some padding for
	// UTF-8 encoding/nulls/etc. but a much smaller bound would be okay as well.
	// If so, we can ignore the chars on a line past that.
	const maxLineLen = 2083 * 8

	var b byte

	p.handler.HandleRobotsStart()

	length := len(p.robotsBody)
	cur := 0
	// Skip BOM if present - including partial BOMs.
	for i := 0; i < len(utfBOM); i++ {
		if cur == length {
			break
		}
		b = p.robotsBody[cur]
		if b != utfBOM[i] {
			break
		}
		cur++
	}

	lineNum := 0
	lastWasCarriageReturn := false
	start := cur
	end := cur
	for {
		if cur == length {
			break
		}
		b = p.robotsBody[cur]
		cur++
		if b != 0x0A && b != 0x0D { // Non-line-ending char case.
			// Put in next spot on current line, as long as there's room.
			if end-start < maxLineLen-1 {
				end++
			}
		} else { // Line-ending character char case.
			// Only emit an empty line if this was not due to the second character
			// of the DOS line-ending \r\n .
			isCRLFContinuation := end-start == 0 && lastWasCarriageReturn && b == 0x0A
			if !isCRLFContinuation {
				lineNum++
				p.parseAndEmitLine(lineNum, p.robotsBody[start:end])
			}
			start = cur
			end = cur
			lastWasCarriageReturn = b == 0x0D
		}
	}
	lineNum++
	p.parseAndEmitLine(lineNum, p.robotsBody[start:end])
	p.handler.HandleRobotsEnd()
}

//

var _ MatchStrategy = LongestMatchStrategy{}

// LongestMatchStrategy implements the default robots.txt matching strategy.
//
// The maximum number of characters matched by a pattern is returned as its match priority.
type LongestMatchStrategy struct{}

func (s LongestMatchStrategy) MatchAllow(path, pattern string) int {
	// Line :640
	if Matches(path, pattern) {
		return len(pattern)
	}
	return -1
}

func (s LongestMatchStrategy) MatchDisallow(path, pattern string) int {
	// Line :645
	if Matches(path, pattern) {
		return len(pattern)
	}
	return -1
}

//

// Parse uses the given robots.txt body and ParseHandler
// to create a Parser, and calls its Parse method.
func Parse(robotsBody string, handler ParseHandler) {
	// Line :454
	parser := NewParser(robotsBody, handler)
	parser.Parse()
}

//

// NewRobotsMatcher creates a RobotsMatcher with the default matching strategy. The default
// matching strategy is longest-match as opposed to the former internet draft
// that provisioned first-match strategy. Analysis shows that longest-match,
// while more restrictive for crawlers, is what webmasters assume when writing
// directives. For example, in case of conflicting matches (both Allow and
// Disallow), the longest match is the one the user wants. For example, in
// case of a robots.txt file that has the following rules
//   Allow: /
//   Disallow: /cgi-bin
// it's pretty obvious what the webmaster wants: they want to allow crawl of
// every URI except /cgi-bin. However, according to the expired internet
// standard, crawlers should be allowed to crawl everything with such a rule.
func NewRobotsMatcher() *RobotsMatcher {
	// Line :460
	m := RobotsMatcher{
		allow:         newMatchHierarchy(),
		disallow:      newMatchHierarchy(),
		MatchStrategy: LongestMatchStrategy{},
	}
	return &m
}

// init Initialises next path and user-agents to check. Path must contain only the
// path, params, and query (if any) of the url and must start with a '/'.
func (m *RobotsMatcher) init(userAgents []string, path string) {
	// Line :478
	m.path = path
	if path[0] != '/' {
		panic("Path must begin with '/'") // TODO Cleanup this panic.
	}
	m.userAgents = userAgents
}

// AgentsAllowed parses the given robots.txt content, matching it against
// the given userAgents and URI, and returns true if access is allowed.
func (m *RobotsMatcher) AgentsAllowed(robotsBody string, userAgents []string, uri string) bool {
	// Line :487
	// The url is not normalized (escaped, percent encoded) here because the user
	// is asked to provide it in escaped form already.
	path := getPathParamsQuery(uri)
	m.init(userAgents, path)
	Parse(robotsBody, m)
	return !m.disallowed()
}

// AgentsAllowed parses the given robots.txt content, matching it against
// the given userAgents and URI, and returns true if access is allowed.
func AgentsAllowed(robotsBody string, userAgents []string, uri string) bool {
	return NewRobotsMatcher().AgentsAllowed(robotsBody, userAgents, uri)
}

// AgentAllowed parses the given robots.txt content, matching it against
// the given userAgent and URI, and returns true if access is allowed.
func (m *RobotsMatcher) AgentAllowed(robotsBody string, userAgent string, uri string) bool {
	// Line :498
	return m.AgentsAllowed(robotsBody, []string{userAgent}, uri)
}

// AgentAllowed parses the given robots.txt content, matching it against
// the given userAgent and URI, and returns true if access is allowed.
func AgentAllowed(robotsBody string, userAgent string, uri string) bool {
	return NewRobotsMatcher().AgentAllowed(robotsBody, userAgent, uri)
}

func (m *RobotsMatcher) disallowed() bool {
	// Line :506
	if m.allow.specific.priority > 0 || m.disallow.specific.priority > 0 {
		return m.disallow.specific.priority > m.allow.specific.priority
	}

	if m.everSeenSpecificAgent {
		// Matching group for user-agent but either without disallow or empty one,
		// i.e. priority == 0.
		return false
	}

	if m.disallow.global.priority > 0 || m.allow.global.priority > 0 {
		return m.disallow.global.priority > m.allow.global.priority
	}
	return false
}

func (m *RobotsMatcher) disallowedIgnoreGlobal() bool {
	// Line :523
	if m.allow.specific.priority > 0 || m.disallow.specific.priority > 0 {
		return m.disallow.specific.priority > m.allow.specific.priority
	}
	return false
}

func (m *RobotsMatcher) matchingLine() int {
	// Line :530
	if m.everSeenSpecificAgent {
		return higherPriorityMatch(m.disallow.specific, m.allow.specific).line
	}
	return higherPriorityMatch(m.disallow.global, m.allow.global).line
}

// HandleRobotsStart is called at the start of parsing a robots.txt file,
// and resets all instance member variables.
func (m *RobotsMatcher) HandleRobotsStart() {
	// Line :538
	// We init in the same order the instance member variables are declared, so
	// it's easier to keep track of which ones we have (or maybe haven't!) done.
	m.allow.Clear()
	m.disallow.Clear()

	m.seenGlobalAgent = false
	m.seenSpecificAgent = false
	m.everSeenSpecificAgent = false
	m.seenSeparator = false
}

// extractUserAgent extracts the matchable part of a user agent string,
// essentially stopping at the first invalid character.
// Example: 'Googlebot/2.1' becomes 'Googlebot'
func (m *RobotsMatcher) extractUserAgent(userAgent string) string {
	// Line :552
	// Allowed characters in user-agent are [a-zA-Z_-].
	i := 0
	for ; i < len(userAgent); i++ {
		c := userAgent[i]
		if !(asciiIsAlpha(c) || c == '-' || c == '_') {
			break
		}
	}
	return userAgent[:i]
}

func asciiIsAlpha(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// isValidUserAgentToObey verifies that the given user agent is valid to be matched against
// robots.txt. Valid user agent strings only contain the characters
// [a-zA-Z_-].
func (m *RobotsMatcher) isValidUserAgentToObey(userAgent string) bool {
	// Line :562
	return len(userAgent) > 0 && m.extractUserAgent(userAgent) == userAgent
}

// HandleUserAgent is called for every "User-Agent:" line in robots.txt.
func (m *RobotsMatcher) HandleUserAgent(lineNum int, userAgent string) {
	// Line :567
	if m.seenSeparator {
		m.seenSpecificAgent = false
		m.seenGlobalAgent = false
		m.seenSeparator = false
	}

	// Google-specific optimization: a '*' followed by space and more characters
	// in a user-agent record is still regarded a global rule.
	if len(userAgent) >= 1 && userAgent[0] == '*' &&
		(len(userAgent) == 1 || isSpace(userAgent[1])) {
		m.seenGlobalAgent = true
	} else {
		userAgent = m.extractUserAgent(userAgent)
		for _, agent := range m.userAgents {
			if equalsIgnoreCase(userAgent, agent) {
				m.everSeenSpecificAgent = true
				m.seenSpecificAgent = true
				break
			}
		}
	}
}

func isSpace(c byte) bool {
	return unicode.IsSpace(rune(c))
	// return c == ' ' || c == '\t'
}

func equalsIgnoreCase(a, b string) bool {
	return strings.EqualFold(a, b)
	// return strings.ToLower(a) == strings.ToLower(b)
}

// HandleAllow is called for every "Allow:" line in robots.txt.
func (m *RobotsMatcher) HandleAllow(lineNum int, value string) {
	// Line :589
	if !m.seenAnyAgent() {
		return
	}
	m.seenSeparator = true
	priority := m.MatchStrategy.MatchAllow(m.path, value)
	if priority >= 0 {
		if m.seenSpecificAgent {
			if m.allow.specific.priority < priority {
				m.allow.specific.Set(priority, lineNum)
			}
		} else {
			if !m.seenGlobalAgent {
				panic("Not seen global agent") // TODO Cleanup this panic.
			}
			if m.allow.global.priority < priority {
				m.allow.global.Set(priority, lineNum)
			}
		}
	} else {
		// Google-specific optimization: 'index.htm' and 'index.html' are normalized
		// to '/'.
		slashPos := strings.LastIndexByte(value, '/')

		if slashPos != -1 && strings.HasPrefix(value[slashPos:], "/index.htm") {
			newPattern := value[:slashPos+1] + "$"
			m.HandleAllow(lineNum, newPattern)
		}
	}
}

// HandleDisallow is called for every "Disallow:" line in robots.txt.
func (m *RobotsMatcher) HandleDisallow(lineNum int, value string) {
	// Line :622
	if !m.seenAnyAgent() {
		return
	}
	m.seenSeparator = true
	priority := m.MatchStrategy.MatchDisallow(m.path, value)
	if priority >= 0 {
		if m.seenSpecificAgent {
			if m.disallow.specific.priority < priority {
				m.disallow.specific.Set(priority, lineNum)
			}
		} else {
			if !m.seenGlobalAgent {
				panic("Not seen global agent") // TODO Cleanup this panic.
			}
			if m.disallow.global.priority < priority {
				m.disallow.global.Set(priority, lineNum)
			}
		}
	}
}

// HandleRobotsEnd is called at the end of parsing the robots.txt file.
//
// For RobotsMatcher, this does nothing.
func (m *RobotsMatcher) HandleRobotsEnd() {}

// HandleSitemap is called for every "Sitemap:" line in robots.txt.
//
// For RobotsMatcher, this does nothing.
func (m *RobotsMatcher) HandleSitemap(lineNum int, value string) {}

// HandleUnknownAction is called for every unrecognised line in robots.txt.
//
// For RobotsMatcher, this does nothing.
func (m *RobotsMatcher) HandleUnknownAction(lineNum int, action, value string) {}
