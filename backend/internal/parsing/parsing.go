package parsing

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func NormalizeAndStripURL(inputURL string) (string, error) {
	// parse URL via url.Parse
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("Failed to parse url: '%s': %w", inputURL, err)
	}

	// lowercase url scheme and host
	if parsedURL.Scheme != "" {
		parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	}
	parsedURL.Host = strings.ToLower(parsedURL.Host)

	// canonicalize path if host is present
	if parsedURL.Host != "" && parsedURL.Path == "" {
		parsedURL.Path = "/"
	}

	// remove query
	parsedURL.RawQuery = ""

	// strip fragment
	parsedURL.Fragment = ""

	return parsedURL.String(), nil
}

// Returns true if node is a common block level element
func isBlockElement(nodeName string) bool {
	switch nodeName {
	case "address", "article", "aside", "blockquote", "canvas", "dd", "div", "dl", "dt",
		"fieldset", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6",
		"header", "hr", "li", "main", "nav", "noscript", "ol", "p", "pre", "section",
		"table", "tfoot", "ul", "video", "td", "th", "tr", "caption", "tbody", "thead",
		"colgroup", "col":
		return true
	default:
		return false
	}
}

// Recursively traverses selection, appending text to string builder
func textTraverse(s *goquery.Selection, sb *strings.Builder) {
	s.Contents().Each(func(_ int, child *goquery.Selection) {
		node := child.Get(0)
		if node == nil {
			return
		}

		if node.Type == html.TextNode {
			// raw text, just append
			sb.WriteString(node.Data)
		} else if node.Type == html.ElementNode {
			nodeName := strings.ToLower(node.Data)

			// br and hr treat as spaces
			if nodeName == "br" || nodeName == "hr" {
				sb.WriteString(" ")
				return
			}

			// format pre with .Text()
			if nodeName == "pre" {
				sb.WriteString(child.Text())
				sb.WriteString(" ")
				return
			}

			// recursion
			textTraverse(child, sb)

			// append string to block elements
			if isBlockElement(nodeName) {
				sb.WriteString(" ")
			}
		}
	})
}

func CleanText(doc *goquery.Selection) string {
	doc.Find("script, style, noscript, iframe, nav, footer, header, form, link").Remove()

	var sb strings.Builder

	var contentRoot *goquery.Selection
	if doc.Is("html") {
		body := doc.Find("body")
		if body.Length() > 0 {
			contentRoot = body
		} else {
			// no body element found
			contentRoot = doc
		}
	} else {
		contentRoot = doc
	}

	textTraverse(contentRoot, &sb)

	text := strings.Join(strings.Fields(sb.String()), " ")
	return text
}

// returns words and their positions in the text
func TokenizeText(text string) map[string][]int {
	words := make(map[string][]int)

	allTokens := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	tokenIndex := 0
	for _, raw := range allTokens {
		word := strings.ToLower(raw)
		if word == "" {
			continue
		}

		// only track non-stopwords
		if !Stopwords[word] {
			words[word] = append(words[word], tokenIndex)
		}
		tokenIndex++
	}

	return words
}

func TokenizeQuery(query string) []string {
	var terms []string

	// split by whitespace, lowercase and strip puncuation
	tokens := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	for _, raw := range tokens {
		word := strings.ToLower(raw)
		if Stopwords[word] || word == "" {
			continue
		}
		terms = append(terms, word)
	}

	// if query only consists of stop word(s)
	if len(terms) == 0 {
		for _, raw := range tokens {
			word := strings.ToLower(raw)
			if word == "" {
				continue
			}
			terms = append(terms, word)
		}
	}

	return terms
}

func SplitWords(text string) []string {
	var words []string

	allTokens := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	for _, raw := range allTokens {
		word := strings.ToLower(raw)
		if word == "" {
			continue
		}

		words = append(words, raw)
	}

	return words

}
