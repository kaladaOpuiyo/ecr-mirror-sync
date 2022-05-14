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

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List ECR repositories and tags marked for mirroring",
		Long:  `List ECR repositories and tags marked for mirroring`,
		Run: func(cmd *cobra.Command, args []string) {

			opts := options.MirrorOptions{}

			opts.RenderTable = true
			mirrorRepos := mirror.New(&opts)
			start := time.Now()
			mirrorRepos.List()
			elapsed := time.Since(start)
			log.Infof("List took %s", elapsed)

		},
	}
	return listCmd
}
