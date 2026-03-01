package tagger

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TagResult is a single tag hit in the vault.
type TagResult struct {
	Topic string
	File  string
	Tag   string
}

// Search walks topicRoot and returns all TagResults matching tagName.
// Tag format: #tagname (word-boundary match, case-insensitive).
// Only .md files are scanned. Walk is confined to topicRoot — never the vault root.
func Search(topicRoot, tagName string) ([]TagResult, error) {
	pattern := regexp.MustCompile(`(?i)#` + regexp.QuoteMeta(tagName) + `\b`)

	var results []TagResult

	err := filepath.WalkDir(topicRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if pattern.Match(content) {
			rel, _ := filepath.Rel(topicRoot, path)
			parts := strings.SplitN(rel, string(os.PathSeparator), 2)
			topic := parts[0]
			results = append(results, TagResult{
				Topic: topic,
				File:  filepath.Base(path),
				Tag:   tagName,
			})
		}
		return nil
	})

	return results, err
}
