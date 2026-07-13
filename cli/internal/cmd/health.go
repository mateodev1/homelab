package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type healthResponse struct {
	Status string `json:"status"`
	DBOk   bool   `json:"db_ok"`
}

func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "health",
		Short:        "Check backend health (no auth required)",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient(cmd, false)
			if err != nil {
				return err
			}
			raw, _, err := c.Do(cmd.Context(), "GET", "/api/health", nil)
			if err != nil {
				return err
			}
			var h healthResponse
			if err := json.Unmarshal(raw, &h); err != nil {
				return fmt.Errorf("parse health response: %w", err)
			}
			return printJSON(cmd, h)
		},
	}
}