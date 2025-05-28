package parsing

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

func CleanText(doc *goquery.Selection) string {
	doc.Find("script, style, noscript, iframe, nav, footer, header, form, link").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})
	text := doc.Find("body").Text()
	content := strings.TrimSpace(text)
	content = strings.Join(strings.Fields(content), " ")
	return content
}
