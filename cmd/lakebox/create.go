package lakebox

import (
	"fmt"

	"github.com/databricks/cli/libs/cmdctx"
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Lakebox environment",
		Long: `Create a new Lakebox environment.

Creates a new personal development environment backed by a microVM.
Blocks until the lakebox is running and prints the lakebox ID.

Examples:
  lakebox create
  lakebox create --name my-project`,
		PreRunE: mustWorkspaceClient,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			w := cmdctx.WorkspaceClient(ctx)
			api := newLakeboxAPI(w)
			stderr := cmd.ErrOrStderr()

			s := spin(stderr, "Provisioning your lakebox…")

			result, err := api.create(ctx, name)
			if err != nil {
				s.fail("Failed to create lakebox")
				return fmt.Errorf("failed to create lakebox: %w", err)
			}

			s.ok(fmt.Sprintf("Lakebox %s is %s", bold(result.SandboxID), status(result.Status)))

			profile := w.Config.Profile
			if profile == "" {
				profile = w.Config.Host
			}

			currentDefault := getDefault(profile)
			shouldSetDefault := currentDefault == ""
			if !shouldSetDefault && currentDefault != "" {
				if _, err := api.get(ctx, currentDefault); err != nil {
					shouldSetDefault = true
				}
			}
			if shouldSetDefault {
				if err := setDefault(profile, result.SandboxID); err != nil {
					warn(stderr, fmt.Sprintf("Could not save default: %v", err))
				} else {
					field(stderr, "default", result.SandboxID)
				}
			}

			blank(stderr)
			fmt.Fprintln(cmd.OutOrStdout(), result.SandboxID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Display label for the lakebox (max 256 bytes)")

	return cmd
}
