/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"ecr-mirror-sync/pkg/options"
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/reexec"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	if reexec.Init() {
		return
	}
	rootCmd, _ := coreOptions()
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

//  coreOptions returns a cobra.Command, and the underlying globalOptions object, to be run or tested.
func coreOptions() (*cobra.Command, *options.GlobalOptions) {
	globalOpts := options.GlobalOptions{}
	rootCmd := &cobra.Command{
		Use:               "ecr-mirror-sync",
		Long:              "Tool used to Sync Public Images with ECR Repositories",
		RunE:              requireSubcommand,
		SilenceUsage:      true,
		SilenceErrors:     true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		TraverseChildren:  true,
	}
	rootCmd.PersistentFlags().BoolVar(&globalOpts.InsecurePolicy, "insecure-policy", false, "run the tool without any policy check")
	rootCmd.PersistentFlags().StringVar(&globalOpts.OverrideArch, "override-arch", "amd64", "use `ARCH` instead of the architecture of the machine for choosing images")
	rootCmd.PersistentFlags().StringVar(&globalOpts.OverrideOS, "override-os", "linux", "use `OS` instead of the running OS for choosing images")
	rootCmd.PersistentFlags().StringVar(&globalOpts.OverrideVariant, "override-variant", "", "use `VARIANT` instead of the running architecture variant for choosing images")
	rootCmd.PersistentFlags().StringVar(&globalOpts.PolicyPath, "policy", "", "Path to a trust policy file")

	rootCmd.AddCommand(
		listCmd(),
		copyCmd(&globalOpts),
		syncCmd(&globalOpts),
	)
	return rootCmd, &globalOpts
}

// requireSubcommand returns an error if no sub command is provided
// This was copied from podman: `github.com/containers/podman/cmd/podman/validate/args.go
// Some small style changes to match skopeo were applied, but try to apply any
// bugfixes there first.
func requireSubcommand(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		suggestions := cmd.SuggestionsFor(args[0])
		if len(suggestions) == 0 {
			return fmt.Errorf("unrecognized command `%[1]s %[2]s`\nTry '%[1]s --help' for more information", cmd.CommandPath(), args[0])
		}
		return fmt.Errorf("unrecognized command `%[1]s %[2]s`\n\nDid you mean this?\n\t%[3]s\n\nTry '%[1]s --help' for more information", cmd.CommandPath(), args[0], strings.Join(suggestions, "\n\t"))
	}
	return fmt.Errorf("missing command '%[1]s COMMAND'\nTry '%[1]s --help' for more information", cmd.CommandPath())
}
