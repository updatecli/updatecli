package cmd

import (
	"github.com/spf13/cobra"
	"github.com/updatecli/updatecli/pkg/core/compose"
)

var (
	// composeCmdClean represents the compose clean flag to enable or disable the clean stage
	composeCmdClean bool
	// composeCmdDisablePrepare represents the compose disable prepare flag to enable or disable the prepare stage
	composeCmdDisablePrepare bool
	// composeCmdDisableTemplating represents the compose disable templating flag to enable or disable the templating stage
	composeCmdDisableTemplating bool
	// composeCmdFile represents the compose filename
	composeCmdFile string
	// composeDefaultCmdFile represents the default compose filename
	composeDefaultCmdFile = compose.GetDefaultComposeFilename()
	// composeCmd represents the compose command
	composeCmd = &cobra.Command{
		Use:   "compose",
		Short: "compose executes specific Updatecli compose tasks such as diff or apply",
	}
	// policyFolder represents the policy folder
	policyFolder string
)
