package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"pm-cli/pkg/api"
)

var (
	dateArg string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List time entries for a given date",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dateArg == "" {
			dateArg = time.Now().Format("2006-01-02")
		}

		client := api.New()
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		url := fmt.Sprintf("%s/time_card_control/current/work_days/%s", api.BaseURL, dateArg)
		req, err := client.NewAuthenticatedRequest(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := client.HTTP.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Fprintf(os.Stderr, "HTTP %d\n%s\n", resp.StatusCode, string(body))
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}

		var pretty any
		if json.Unmarshal(body, &pretty) == nil {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(pretty)
		}
		fmt.Println(string(body))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&dateArg, "date", "", "Date in YYYY-MM-DD format (default: today)")
}


