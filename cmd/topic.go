package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ReggieAlbiosA/vb/internal/index"
	"github.com/ReggieAlbiosA/vb/internal/resolver"
	"github.com/spf13/cobra"
)

var (
	flagTopicIn   string
	flagTopicTree bool
)

var topicCmd = &cobra.Command{
	Use:   "topic",
	Short: "Manage topics in the active vault",
	Long: `Create and list topics in the active vault.

Topics are directories containing lens files (.md). Use --in to nest
topics inside existing ones, and .. to address nested topics when querying.

Examples:
  vb topic create disk                     # flat topic
  vb topic create fs --in partition        # nested under partition
  vb topic create mnt --in partition..fs   # deep nesting
  vb topic list                            # flat list
  vb topic list --tree                     # indented tree`,
}

var topicCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new topic with lens scaffolding",
	Long: `Create a new topic directory with 6 empty lens files:
  WHY.md, IMPORTANCE.md, CLI_TOOLS.md, ARCH.md, GOTCHAS.md, REFS.md

Use --in to nest the topic under an existing parent topic.
Parent directories are created automatically if they don't exist.

Examples:
  vb topic create disk                     # TopicRoot/disk/
  vb topic create fs --in partition        # TopicRoot/partition/fs/
  vb topic create mnt --in partition..fs   # TopicRoot/partition/fs/mnt/`,
	Args: cobra.ExactArgs(1),
	RunE: runTopicCreate,
}

var topicListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all topics in the active vault",
	Long: `List all indexed topics. By default shows a flat sorted list.
Use --tree to display as an indented tree.

Examples:
  vb topic list          # flat list
  vb topic list --tree   # tree view`,
	Args: cobra.NoArgs,
	RunE: runTopicList,
}

func init() {
	topicCreateCmd.Flags().StringVarP(&flagTopicIn, "in", "i", "", "parent topic (leaf name or full..path)")
	topicListCmd.Flags().BoolVarP(&flagTopicTree, "tree", "t", false, "display as indented tree")

	topicCmd.AddCommand(topicCreateCmd)
	topicCmd.AddCommand(topicListCmd)
}

// scaffoldFiles are the 6 built-in lens files created for every new topic.
var scaffoldFiles = func() []string {
	files := make([]string, 0, len(resolver.LensToFile))
	for _, f := range resolver.LensToFile {
		files = append(files, f)
	}
	sort.Strings(files)
	return files
}()

func runTopicCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	ctx, err := resolveVault()
	if err != nil {
		return err
	}

	var targetDir string
	if flagTopicIn != "" {
		// Resolve the parent topic via the index.
		schema, loadErr := index.Load(ctx.VaultRoot)
		if loadErr != nil {
			return loadErr
		}
		parentDir, resolveErr := resolver.ResolveTopic(flagTopicIn, schema, ctx.TopicRoot)
		if resolveErr != nil {
			return resolveErr
		}
		targetDir = filepath.Join(parentDir, name)
	} else {
		targetDir = filepath.Join(ctx.TopicRoot, name)
	}

	// Guard: don't overwrite an existing topic.
	if hasMDFiles(targetDir) {
		return fmt.Errorf("topic %q already exists at %s", name, targetDir)
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating topic directory: %w", err)
	}

	// Scaffold 6 empty lens files.
	for _, f := range scaffoldFiles {
		fp := filepath.Join(targetDir, f)
		if err := os.WriteFile(fp, []byte(""), 0o644); err != nil {
			return fmt.Errorf("creating %s: %w", f, err)
		}
	}

	// Auto-reindex so the new topic is immediately queryable.
	if _, err := index.Build(ctx.VaultRoot, ctx.TopicRoot); err != nil {
		return fmt.Errorf("reindexing after creation: %w", err)
	}

	rel, _ := filepath.Rel(ctx.TopicRoot, targetDir)
	fmt.Fprintf(cmd.OutOrStdout(), "created topic %q at %s\n", name, filepath.ToSlash(rel))
	return nil
}

func runTopicList(cmd *cobra.Command, args []string) error {
	ctx, err := resolveVault()
	if err != nil {
		return err
	}

	schema, err := index.Load(ctx.VaultRoot)
	if err != nil {
		return err
	}

	if len(schema.Topics) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no topics found — create one with: vb topic create <name>")
		return nil
	}

	if flagTopicTree {
		return printTree(cmd, schema)
	}
	return printFlat(cmd, schema)
}

// printFlat prints a sorted list of topics with their relative paths.
func printFlat(cmd *cobra.Command, schema index.Schema) error {
	// Deduplicate: collect unique rel paths, pick the best display key.
	type entry struct {
		display string
		rel     string
	}
	seen := make(map[string]entry)
	for key, rel := range schema.Topics {
		prev, exists := seen[rel]
		if !exists {
			seen[rel] = entry{display: key, rel: rel}
			continue
		}
		// Prefer the ..-joined key over the leaf-only key for nested topics.
		if strings.Contains(key, "..") && !strings.Contains(prev.display, "..") {
			seen[rel] = entry{display: key, rel: rel}
		}
	}

	// Sort by display key.
	entries := make([]entry, 0, len(seen))
	for _, e := range seen {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].display < entries[j].display })

	w := cmd.OutOrStdout()
	for _, e := range entries {
		fmt.Fprintf(w, "  %-25s %s\n", e.display, e.rel)
	}
	return nil
}

// printTree prints an indented tree of topics.
func printTree(cmd *cobra.Command, schema index.Schema) error {
	// Collect unique relative paths (values), deduplicated.
	pathSet := make(map[string]bool)
	for _, rel := range schema.Topics {
		pathSet[rel] = true
	}
	paths := make([]string, 0, len(pathSet))
	for p := range pathSet {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	w := cmd.OutOrStdout()
	for _, p := range paths {
		parts := strings.Split(p, "/")
		depth := len(parts) - 1
		indent := strings.Repeat("  ", depth)
		fmt.Fprintf(w, "%s%s\n", indent, parts[len(parts)-1])
	}
	return nil
}

// hasMDFiles reports whether dir directly contains at least one .md file.
func hasMDFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(filepath.Ext(e.Name())) == ".md" {
			return true
		}
	}
	return false
}
