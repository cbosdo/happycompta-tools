// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// Helper function to create a node from an HTML string snippet for testing
func parseSnippet(t *testing.T, htmlStr string) *html.Node {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("Failed to parse snippet: %v", err)
	}
	return doc
}

func TestGetAttr(t *testing.T) {
	node := parseSnippet(t, `<div id="test-id" class="test-class">Content</div>`)
	div := findNodeWithTagName(node, "div")

	tests := []struct {
		key      string
		expected string
	}{
		{"id", "test-id"},
		{"class", "test-class"},
		{"non-existent", ""},
	}

	for _, test := range tests {
		result := getAttr(div, test.key)
		if result != test.expected {
			t.Errorf("getAttr(%s) returned %s, expected %s", test.key, result, test.expected)
		}
	}
}

func TestFindNodeWithTagName(t *testing.T) {
	htmlStr := `<body><div><p>Text</p></div><a>Link</a></body>`
	doc := parseSnippet(t, htmlStr)

	tests := []struct {
		tagName  string
		expected bool
	}{
		{"p", true},
		{"a", true},
		{"div", true},
		{"span", false},
	}

	for _, test := range tests {
		result := findNodeWithTagName(doc, test.tagName)
		if (result != nil) != test.expected {
			t.Errorf("findNodeWithTagName(%s) result = %v, expected node presence: %t", test.tagName, result, test.expected)
		}
		if result != nil && result.Data != test.tagName {
			t.Errorf("findNodeWithTagName(%s) returned node with wrong tag: %s", test.tagName, result.Data)
		}
	}
}

func TestFindNodeWithAttr(t *testing.T) {
	htmlStr := `<body><p data-id="123">Text</p><div><a href="/link">Link</a></div></body>`
	doc := parseSnippet(t, htmlStr)

	tests := []struct {
		attrKey  string
		expected string
	}{
		{"data-id", "123"},
		{"href", "/link"},
		{"class", ""},
	}

	for _, test := range tests {
		resultNode := findNodeWithAttr(doc, test.attrKey)
		resultAttr := ""
		if resultNode != nil {
			resultAttr = getAttr(resultNode, test.attrKey)
		}

		if resultAttr != test.expected {
			t.Errorf("findNodeWithAttr(%s) returned attr %s, expected %s", test.attrKey, resultAttr, test.expected)
		}
	}
}

func TestFindNodeWithKeyValueAttr(t *testing.T) {
	htmlStr := `<body><div class="target">Found</div><p class="other">Skip</p></body>`
	doc := parseSnippet(t, htmlStr)

	tests := []struct {
		key      string
		value    string
		expected bool
	}{
		{"class", "target", true},
		{"class", "other", true},
		{"id", "target", false},
		{"class", "missing", false},
	}

	for _, test := range tests {
		result := findNodeWithKeyValueAttr(doc, test.key, test.value)
		if (result != nil) != test.expected {
			t.Errorf("findNodeWithKeyValueAttr(%s, %s) result = %v, expected node presence: %t", test.key, test.value, result, test.expected)
		}
		if result != nil && getAttr(result, test.key) != test.value {
			t.Errorf("findNodeWithKeyValueAttr(%s, %s) returned node with wrong value: %s", test.key, test.value, getAttr(result, test.key))
		}
	}
}

func TestExtractTextContent(t *testing.T) {
	htmlStr := `<td>  Text Before <strong>Nested Text</strong> Text After   </td>`
	td := parseSnippet(t, htmlStr)

	// find the <td> node

	expected := "Text Before Nested Text Text After"
	result := extractTextContent(td)

	if result != expected {
		t.Errorf("extractTextContent failed. Got: '%s', Expected: '%s'", result, expected)
	}

	// Test case with just whitespace and newlines
	htmlStrWhitespace := `<td>
		 
	 </td>`
	tdWhitespace := parseSnippet(t, htmlStrWhitespace)
	expectedWhitespace := ""
	resultWhitespace := extractTextContent(tdWhitespace)

	if resultWhitespace != expectedWhitespace {
		t.Errorf("extractTextContent with whitespace failed. Got: '%s', Expected: '%s'", resultWhitespace, expectedWhitespace)
	}
}
