package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/cryskram/relith/internal/search"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search indexed code",
	Long: `Search across all indexed repositories using SQLite FTS5 full-text search.

Simple term:  relith search "auth middleware"
Phrase:       relith search "\"jwt token\""
Prefix:       relith search "auth*"
Advanced:     relith search "auth AND (middleware OR interceptor)"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := openDB()
		if err != nil {
			return err
		}
		defer app.close()

		limitStr, _ := cmd.Flags().GetString("limit")
		limit := app.cfg.Search.MaxResults
		if limitStr != "" {
			parsed, err := strconv.Atoi(limitStr)
			if err == nil && parsed > 0 {
				limit = parsed
			}
		}

		searcher := search.New(app.db, app.logger, app.cfg.Search)
		results, err := searcher.Search(context.Background(), args[0], limit)
		if err != nil {
			return fmt.Errorf("search: %w", err)
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		fmt.Printf("%d result(s) for %q:\n\n", len(results), args[0])

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for i, r := range results {
			fmt.Fprintf(w, "%d.\t%s\t[%s]\n", i+1, r.Path, r.RepoName)
			fmt.Fprintf(w, "\t%s\n", truncateForDisplay(r.Content, 120))
			fmt.Fprintln(w, "\t")
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().String("limit", "", "Max results (default: from config)")
}

func truncateForDisplay(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
