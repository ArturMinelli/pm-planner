package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"pm-cli/pkg/auth"
)

var (
	dateArg string
)

// listCmd calls the authenticated endpoint to list the day's points
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List time entries for a given date",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dateArg == "" {
			dateArg = time.Now().Format("2006-01-02")
		}

        headers, err := auth.GetAuthHeaders()
		if err != nil {
			return err
		}

		url := fmt.Sprintf("https://api.pontomais.com.br/api/time_card_control/current/work_days/%s", dateArg)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			// Print server response to help debugging
			fmt.Fprintf(os.Stderr, "HTTP %d\n%s\n", resp.StatusCode, string(body))
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}

		// pretty print JSON if possible; otherwise print raw
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


