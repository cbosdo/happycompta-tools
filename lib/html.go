// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"strings"

	"golang.org/x/net/html"
)

// getAttr is a helper function to find an attribute value by key.
func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// findNodeWithTagName recursively traverses the node and its descendants
// starting from n until it finds an ElementNode with the specified tag name.
func findNodeWithTagName(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNodeWithTagName(c, tagName); found != nil {
			return found
		}
	}
	return nil
}

// findNodeWithAttr recursively traverses the node's children and siblings
// starting from n until it finds a node possessing the specified attribute key.
func findNodeWithAttr(n *html.Node, attrKey string) *html.Node {
	// Check the current node
	if getAttr(n, attrKey) != "" {
		return n
	}

	// Recursively search children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNodeWithAttr(c, attrKey); found != nil {
			return found
		}
	}
	return nil
}

// findNodeWithKeyValueAttr recursively traverses the node's children and siblings
// starting from n until it finds an ElementNode possessing the specified attribute key
// with the specified attribute value.
func findNodeWithKeyValueAttr(n *html.Node, key, value string) *html.Node {
	// Check the current node
	if n.Type == html.ElementNode && getAttr(n, key) == value {
		return n
	}

	// Recursively search children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNodeWithKeyValueAttr(c, key, value); found != nil {
			return found
		}
	}
	return nil
}

// extractTextContent recursively extracts and concatenates all text content from a node and its descendants.
// It trims leading/trailing whitespace from the resulting string.
func extractTextContent(node *html.Node) string {
	var builder strings.Builder
	var traverseText func(*html.Node)
	traverseText = func(n *html.Node) {
		if n.Type == html.TextNode {
			builder.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverseText(c)
		}
	}
	traverseText(node)
	return strings.TrimSpace(builder.String())
}

// findClassText gets the text of a node with the given class name.
func findClassText(node *html.Node, className string) string {
	found := findNodeWithKeyValueAttr(node, "class", className)
	if found != nil {
		return extractTextContent(found)
	}
	return ""
}
