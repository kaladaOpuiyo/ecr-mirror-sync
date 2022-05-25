/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	mirror "ecr-mirror-sync/pkg/mirror"
	"ecr-mirror-sync/pkg/options"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {

	mirrorFlags, mirrorOpts := options.MirrorFlags(nil, nil, nil, nil)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List ECR repositories and tags marked for mirroring",
		Long:  `List ECR repositories and tags marked for mirroring`,
		Run: func(cmd *cobra.Command, args []string) {

			opts := mirrorOpts

			opts.RenderTable = true
			mirrorRepos := mirror.New(opts)
			start := time.Now()
			mirrorRepos.List()
			elapsed := time.Since(start)
			log.Infof("List took %s", elapsed)

		},
	}

	flags := listCmd.Flags()
	flags.AddFlagSet(&mirrorFlags)

	return listCmd
}
