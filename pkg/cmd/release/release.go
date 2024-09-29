package rel

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/templates"
	cmdutil "kusionstack.io/kusion/pkg/cmd/util"
	"kusionstack.io/kusion/pkg/util/i18n"
)

var relLong = i18n.T(`
		Commands for observing and operating Kusion release files. 
		
		These commands help you observe and operate the Kusion release files of a Project in a Workspace. `)

// NewCmdRel returns an initialized Command instance for 'release' sub command.
func NewCmdRel(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "release",
		DisableFlagsInUseLine: true,
		Short:                 "Observe and operate Kusion release files",
		Long:                  templates.LongDesc(relLong),
		Run:                   cmdutil.DefaultSubCommandRun(streams.ErrOut),
	}

	cmd.AddCommand(NewCmdUnlock(streams), NewCmdList(streams), NewCmdShow(streams))

	return cmd
}
