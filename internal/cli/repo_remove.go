package cli

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cryskram/relith/internal/db"
)

var repoRemoveCmd = &cobra.Command{
	Use:   "remove <id-or-name>",
	Short: "Remove a repository and all its data",
	Long:  `Remove a repository from the index by ID or name. All documents, chunks, symbols, and references for this repository are deleted.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := openDB()
		if err != nil {
			return err
		}
		defer app.close()

		q := db.New(app.db)

		arg := args[0]

		// Try by ID first
		if id, parseErr := strconv.ParseInt(arg, 10, 64); parseErr == nil {
			repo, err := q.GetRepo(context.Background(), id)
			if err == nil {
				if err := q.DeleteRepo(context.Background(), repo.ID); err != nil {
					return fmt.Errorf("delete repo: %w", err)
				}
				fmt.Printf("Removed repository: id=%d  name=%s  path=%s\n", repo.ID, repo.Name, repo.Path)
				vacuumIfNeeded(app.db)
				return nil
			}
		}

		// Try by name
		repos, err := q.ListRepos(context.Background())
		if err != nil {
			return fmt.Errorf("list repos: %w", err)
		}

		for _, r := range repos {
			if r.Name == arg {
				if err := q.DeleteRepo(context.Background(), r.ID); err != nil {
					return fmt.Errorf("delete repo: %w", err)
				}
				fmt.Printf("Removed repository: id=%d  name=%s  path=%s\n", r.ID, r.Name, r.Path)
				vacuumIfNeeded(app.db)
				return nil
			}
		}

		return fmt.Errorf("repository not found: %s", arg)
	},
}

func vacuumIfNeeded(database *sql.DB) {
	var pageCount, freelistCount int
	ctx := context.Background()
	database.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pageCount)
	database.QueryRowContext(ctx, "PRAGMA freelist_count").Scan(&freelistCount)
	if pageCount == 0 {
		return
	}
	freeRatio := float64(freelistCount) / float64(pageCount)
	if freeRatio > 0.2 {
		fmt.Printf("Free pages: %.0f%% of database. Running VACUUM to reclaim space...\n", freeRatio*100)
		if _, err := database.ExecContext(ctx, "VACUUM"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: VACUUM failed: %v\n", err)
			return
		}
		fmt.Println("VACUUM complete.")
	}
}

func init() {
	repoCmd.AddCommand(repoRemoveCmd)
}
