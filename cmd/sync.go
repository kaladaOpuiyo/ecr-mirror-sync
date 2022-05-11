// /*
// Copyright © 2022 NAME HERE <EMAIL ADDRESS>

// */
package cmd

import (
	mirror "ecr-mirror-sync/pkg/mirror"
	"ecr-mirror-sync/pkg/options"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func syncCmd(global *options.GlobalOptions) *cobra.Command {

	srcFlags, srcOpts := options.ImageFlags(global, "src-", "screds")
	destFlags, destOpts := options.ImageDestFlags(global, "dest-", "dcreds")
	retryFlags, retryOpts := options.RetryFlags()
	mirrorFlags, mirrorOpts := options.MirrorFlags(global, srcOpts, destOpts, retryOpts)

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync all ECR repositories tagged to be mirror with public repositories",
		Long:  `Sync all ECR repositories tagged to be mirror with public repositories`,
		Run: func(cmd *cobra.Command, args []string) {

			log.Info("syncing external repositories...")

			opts := mirrorOpts

			mirrorRepos := mirror.New(opts)
			start := time.Now()
			mirrorRepos.Sync()
			elapsed := time.Since(start)
			log.Infof("Sync Completed.Sync took %s", elapsed)
		},
	}

	flags := syncCmd.Flags()
	flags.AddFlagSet(&destFlags)
	flags.AddFlagSet(&mirrorFlags)
	flags.AddFlagSet(&retryFlags)
	flags.AddFlagSet(&srcFlags)
	return syncCmd
}
