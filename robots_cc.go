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
	"io"
	"strings"
	"unicode"
)

// AllowFrequentTypos to allow for typos such as DISALOW in robots.txt.
var AllowFrequentTypos = true

// A RobotsMatchStrategy defines a strategy for matching individual lines in a
// robots.txt file. Each Match* method should return a match priority, which is
// interpreted as:
//
// match priority < 0:
//    No match.
//
// match priority == 0:
//    Match, but treat it as if matched an empty pattern.
//
// match priority > 0:
//    Match.
type RobotsMatchStrategy interface {
	MatchAllow(path, pattern string) int
	MatchDisallow(path, pattern string) int
	Matches(path, pattern string) bool
}

// Implements robots.txt pattern matching.
//
// Returns true if URI path matches the specified pattern. Pattern is anchored
// at the beginning of path. '$' is special only at the end of pattern.
//
// Since 'path' and 'pattern' are both externally determined (by the webmaster),
// we make sure to have acceptable worst-case performance.
func RobotsMatchStrategy_Matches(path, pattern string) bool {
	// Line :69
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

// GetPathParamsQuery is not in anonymous namespace to allow testing.
//
// Extracts path (with params) and query part from URL. Removes scheme,
// authority, and fragment. Result always starts with "/".
// Returns "/" if the url doesn't have a path or is not valid.
func GetPathParamsQuery(uri string) string {
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

// MaybeEscapePattern is not in anonymous namespace to allow testing.
//
// Canonicalize the allowed/disallowed paths. For example:
//     /SanJoséSellers ==> /Sanjos%C3%A9Sellers
//     %aa ==> %AA
// When the function returns, (*dst) either points to src, or is newly
// allocated.
// Returns true if dst was newly allocated.
func MaybeEscapePattern(src string) string {
	// Line :156

	needCapitalise := false
	numToEscape := 0

	s := func(i int) byte {
		if i < len(src) {
			return src[i]
		}
		return 0
	}

	// First, scan the buffer to see if changes are needed. Most don't.
	for i := 0; i < len(src); i++ {
		// (a) % escape sequence.
		if src[i] == '%' &&
			isHexDigit(s(i+1)) && isHexDigit(s(i+2)) {
			if isLower(s(i+1)) || isLower(s(i+2)) {
				needCapitalise = true
			}
		} else if src[i] >= 0x80 {
			// (b) needs escaping.
			numToEscape++
		}
		// (c) Already escaped and escape-characters normalized (eg. %2f -> %2F).
	}
	// Return if no changes needed.
	if numToEscape == 0 && !needCapitalise {
		return src
	}

	by := make([]byte, 0, numToEscape*2+len(src))
	dst := bytes.NewBuffer(by)
	for i := 0; i < len(src); i++ {
		// (a) Normalize %-escaped sequence (eg. %2f -> %2F).
		if src[i] == '%' &&
			isHexDigit(s(i+1)) && isHexDigit(s(i+2)) {
			dst.WriteByte('%')
			i++
			dst.WriteByte(toUpper(src[i]))
			i++
			dst.WriteByte(toUpper(src[i]))
		} else if src[i] >= 0x80 {
			// (b) %-escape octets whose highest bit is set. These are outside the
			// ASCII range.
			dst.WriteByte('%')
			dst.WriteByte(hexDigits[(src[i]>>4)&0xf])
			dst.WriteByte(hexDigits[src[i]&0xf])
		} else {
			// (c) Normal character, no modification needed.
			dst.WriteByte(src[i])
		}
	}
	return string(dst.Bytes())
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

type KeyType int

const (
	// Unrecognized field; Zero value so that additions to the
	// enumeration below do not change the serialization,
	// and to provide useful default.
	Unknown KeyType = iota // Note in the C++ this is 128, not 0.

	// Generic highlevel fields.

	UserAgent
	Sitemap

	// Fields within a user-agent.

	Allow
	Disallow
)

//

// A robots.txt has lines of key/value pairs. A ParsedRobotsKey represents
// a key. This class can parse a text-representation (including common typos)
// and represent them as an enumeration which allows for faster processing
// afterwards.
// For unparsable keys, the original string representation is kept.
type ParsedRobotsKey struct {
	typ     KeyType
	keyText string
}

// Parse given key text. Does not copy the text, so the text_key must stay
// valid for the object's life-time or the next Parse() call.
func (k *ParsedRobotsKey) Parse(key string) {
	// Line :659
	k.keyText = ""
	if k.KeyIsUserAgent(key) {
		k.typ = UserAgent
	} else if k.KeyIsAllow(key) {
		k.typ = Allow
	} else if k.KeyIsDisallow(key) {
		k.typ = Disallow
	} else if k.KeyIsSitemap(key) {
		k.typ = Sitemap
	} else {
		k.typ = Unknown
		k.keyText = key
	}
}

// Returns the type of key.
func (k *ParsedRobotsKey) Type() KeyType {
	return k.typ
}

// If this is an unknown key, get the text.
func (k *ParsedRobotsKey) GetUnknownText() string {
	// Line :675
	if !(k.typ == Unknown && k.keyText != "") {
		panic("bad call to GetUnknownText") // TODO Remove this panic.
	}
	return k.keyText
}

func (k *ParsedRobotsKey) KeyIsUserAgent(key string) bool {
	// Line :680
	return startsWithIgnoreCase(key, "user-agent") ||
		(AllowFrequentTypos && (startsWithIgnoreCase(key, "useragent") ||
			startsWithIgnoreCase(key, "user agent")))
}

func (k *ParsedRobotsKey) KeyIsAllow(key string) bool {
	// Line :687
	return startsWithIgnoreCase(key, "allow")
}

func (k *ParsedRobotsKey) KeyIsDisallow(key string) bool {
	// Line :691
	return startsWithIgnoreCase(key, "disallow") ||
		(AllowFrequentTypos && (startsWithIgnoreCase(key, "dissallow") ||
			startsWithIgnoreCase(key, "dissalow") ||
			startsWithIgnoreCase(key, "disalow") ||
			startsWithIgnoreCase(key, "diasllow") ||
			startsWithIgnoreCase(key, "disallaw")))
}

func (k *ParsedRobotsKey) KeyIsSitemap(key string) bool {
	// Line :701
	return startsWithIgnoreCase(key, "sitemap") ||
		startsWithIgnoreCase(key, "site-map")
}

func startsWithIgnoreCase(x, y string) bool {
	return strings.HasPrefix(strings.ToLower(x), strings.ToLower(y))
}

//

func EmitKeyValueToHandler(line int, key *ParsedRobotsKey, value string, handler RobotsParseHandler) {
	// Line :262
	switch key.Type() {
	case UserAgent:
		handler.HandleUserAgent(line, value)
	case Allow:
		handler.HandleAllow(line, value)
	case Disallow:
		handler.HandleDisallow(line, value)
	case Sitemap:
		handler.HandleSitemap(line, value)
	case Unknown:
		handler.HandleUnknownAction(line, key.GetUnknownText(), value)
	}
}

//

type Key = ParsedRobotsKey

type RobotsTxtParser struct {
	// Line :278
	robotsBody string // TODO Should be []byte
	handler    RobotsParseHandler
}

func NewRobotsTxtParser(robotsBody string, handler RobotsParseHandler) *RobotsTxtParser {
	// Line :282
	p := RobotsTxtParser{
		robotsBody: robotsBody,
		handler:    handler,
	}
	return &p
}

func (p *RobotsTxtParser) NeedEscapeValueForKey(key *Key) bool {
	// Line :300
	switch key.Type() {
	case UserAgent, Sitemap:
		return false
	default:
		return true
	}
}

func (p *RobotsTxtParser) GetKeyAndValueFrom(line string) (string, string, bool) {
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

	if len(key) > 0 {
		value := line[sep+1:]            // Value starts after the separator.
		value = strings.TrimSpace(value) // Get rid of any leading whitespace.
		return key, value, true
	}

	return key, "", false
}

func (p *RobotsTxtParser) ParseAndEmitLine(currentLine int, line string /* TODO []byte here? */) {
	// Line :362
	stringKey, value, ok := p.GetKeyAndValueFrom(line)
	if !ok {
		return
	}

	key := &Key{}
	key.Parse(stringKey)
	if p.NeedEscapeValueForKey(key) {
		escapedValue := MaybeEscapePattern(value)
		EmitKeyValueToHandler(currentLine, key, escapedValue, p.handler)
	} else {
		EmitKeyValueToHandler(currentLine, key, value, p.handler)
	}
}

func (p *RobotsTxtParser) Parse() {
	// TODO see line :381
	// UTF-8 byte order marks.
	utfBOM := []byte{0xEF, 0xBB, 0xBF}

	// Certain browsers limit the URL length to 2083 bytes. In a robots.txt, it's
	// fairly safe to assume any valid line isn't going to be more than many times
	// that max url length of 2KB. We want some padding for
	// UTF-8 encoding/nulls/etc. but a much smaller bound would be okay as well.
	// If so, we can ignore the chars on a line past that.
	const maxLineLen = 2083 * 8

	var err error
	var b byte

	p.handler.HandleRobotsStart()

	r := strings.NewReader(p.robotsBody)
	// Skip BOM if present - including partial BOMs.
	for i := 0; i < len(utfBOM); i++ {
		b, err = r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err) // TODO Cleanup this panic.
		}
		if b != utfBOM[i] {
			r.UnreadByte()
			break
		}
	}

	lineNum := 0
	lastWasCarriageReturn := false

	// TODO I believe this can all be done without a buffer, by moving a slice over p.robotsBody.

	var lineBuffer []byte
	// lineBuffer := make([]byte, 0, maxLineLen)
	for {
		b, err = r.ReadByte()
		if err != nil {
			break
		}
		if b != 0x0A && b != 0x0D { // Non-line-ending char case.
			// Put in next spot on current line, as long as there's room.
			if len(lineBuffer) < maxLineLen-1 {
				lineBuffer = append(lineBuffer, b)
			}
		} else { // Line-ending character char case.
			// Only emit an empty line if this was not due to the second character
			// of the DOS line-ending \r\n .
			isCRLFContinuation := len(lineBuffer) == 0 && lastWasCarriageReturn && b == 0x0A
			if !isCRLFContinuation {
				lineNum++
				p.ParseAndEmitLine(lineNum, string(lineBuffer))
			}
			lineBuffer = lineBuffer[:0]
			lastWasCarriageReturn = b == 0x0D
		}
	}
	lineNum++
	p.ParseAndEmitLine(lineNum, string(lineBuffer))
	p.handler.HandleRobotsEnd()
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

//

var _ RobotsMatchStrategy = LongestMatchRobotsMatchStrategy{}

// Implements the default robots.txt matching strategy. The maximum number of
// characters matched by a pattern is returned as its match priority.
type LongestMatchRobotsMatchStrategy struct{}

func (s LongestMatchRobotsMatchStrategy) MatchAllow(path, pattern string) int {
	// Line :640
	if s.Matches(path, pattern) {
		return len(pattern)
	}
	return -1
}

func (s LongestMatchRobotsMatchStrategy) MatchDisallow(path, pattern string) int {
	// Line :645
	if s.Matches(path, pattern) {
		return len(pattern)
	}
	return -1
}

func (s LongestMatchRobotsMatchStrategy) Matches(path, pattern string) bool {
	return RobotsMatchStrategy_Matches(path, pattern)
}

//

// Parses body of a robots.txt and emits parse callbacks. This will accept
// typical typos found in robots.txt, such as 'disalow'.
//
// Note, this function will accept all kind of input but will skip
// everything that does not look like a robots directive.
func ParseRobotsTxt(robotsBody string, parseCallback RobotsParseHandler) {
	// Line :454
	parser := NewRobotsTxtParser(robotsBody, parseCallback)
	parser.Parse()
}

//

// Create a RobotsMatcher with the default matching strategy. The default
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
		allowMatch:    NewMatchHierarchy(),
		disallowMatch: NewMatchHierarchy(),
		matchStrategy: LongestMatchRobotsMatchStrategy{},
	}
	return &m
}

func (m *RobotsMatcher) InitUserAgentsAndPath(userAgents []string, path string) {
	// Line :478
	m.path = path
	if path[0] != '/' {
		panic("Path must begin with '/'") // TODO Cleanup this panic.
	}
	m.userAgents = userAgents
}

func (m *RobotsMatcher) AllowedByRobots(robotsBody string, userAgents []string, uri string) bool {
	// Line :487
	// The url is not normalized (escaped, percent encoded) here because the user
	// is asked to provide it in escaped form already.
	path := GetPathParamsQuery(uri)
	m.InitUserAgentsAndPath(userAgents, path)
	ParseRobotsTxt(robotsBody, m)
	return !m.disallow()
}

func (m *RobotsMatcher) OneAgentAllowedByRobots(robotsBody string, userAgent string, uri string) bool {
	// Line :498
	return m.AllowedByRobots(robotsBody, []string{userAgent}, uri)
}

func (m *RobotsMatcher) disallow() bool {
	// Line :506
	if m.allowMatch.specific.priority > 0 || m.disallowMatch.specific.priority > 0 {
		return m.disallowMatch.specific.priority > m.allowMatch.specific.priority
	}

	if m.everSeenSpecificAgent {
		// Matching group for user-agent but either without disallow or empty one,
		// i.e. priority == 0.
		return false
	}

	if m.disallowMatch.global.priority > 0 || m.allowMatch.global.priority > 0 {
		return m.disallowMatch.global.priority > m.allowMatch.global.priority
	}
	return false
}

func (m *RobotsMatcher) disallowIgnoreGlobal() bool {
	// Line :523
	if m.allowMatch.specific.priority > 0 || m.disallowMatch.specific.priority > 0 {
		return m.disallowMatch.specific.priority > m.allowMatch.specific.priority
	}
	return false
}

func (m *RobotsMatcher) matchingLine() int {
	// Line :530
	if m.everSeenSpecificAgent {
		return Match_HigherPriorityMatch(m.disallowMatch.specific, m.allowMatch.specific).line
	}
	return Match_HigherPriorityMatch(m.disallowMatch.global, m.allowMatch.global).line
}

func (m *RobotsMatcher) HandleRobotsStart() {
	// Line :538
	// This is a new robots.txt file, so we need to reset all the instance member
	// variables. We do it in the same order the instance member variables are
	// declared, so it's easier to keep track of which ones we have (or maybe
	// haven't!) done.
	m.allowMatch.Clear()
	m.disallowMatch.Clear()

	m.seenGlobalAgent = false
	m.seenSpecificAgent = false
	m.everSeenSpecificAgent = false
	m.seenSeparator = false
}

func /*(m *RobotsMatcher)*/ ExtractUserAgent(userAgent string) string {
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

func /*(m *RobotsMatcher)*/ IsValidUserAgentToObey(userAgent string) bool {
	// Line :562
	return len(userAgent) > 0 && ExtractUserAgent(userAgent) == userAgent
}

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
		userAgent = ExtractUserAgent(userAgent)
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

func (m *RobotsMatcher) HandleAllow(lineNum int, value string) {
	// Line :589
	if !m.seenAnyAgent() {
		return
	}
	m.seenSeparator = true
	priority := m.matchStrategy.MatchAllow(m.path, value)
	if priority >= 0 {
		if m.seenSpecificAgent {
			if m.allowMatch.specific.priority < priority {
				m.allowMatch.specific.Set(priority, lineNum)
			}
		} else {
			if !m.seenGlobalAgent {
				panic("Not seen global agent") // TODO Cleanup this panic.
			}
			if m.allowMatch.global.priority < priority {
				m.allowMatch.global.Set(priority, lineNum)
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

func (m *RobotsMatcher) HandleDisallow(lineNum int, value string) {
	// Line :622
	if !m.seenAnyAgent() {
		return
	}
	m.seenSeparator = true
	priority := m.matchStrategy.MatchDisallow(m.path, value)
	if priority >= 0 {
		if m.seenSpecificAgent {
			if m.disallowMatch.specific.priority < priority {
				m.disallowMatch.specific.Set(priority, lineNum)
			}
		} else {
			if !m.seenGlobalAgent {
				panic("Not seen global agent") // TODO Cleanup this panic.
			}
			if m.disallowMatch.global.priority < priority {
				m.disallowMatch.global.Set(priority, lineNum)
			}
		}
	}
}

func (m *RobotsMatcher) HandleRobotsEnd() {}

func (m *RobotsMatcher) HandleSitemap(lineNum int, value string) {}

func (m *RobotsMatcher) HandleUnknownAction(lineNum int, action, value string) {}
