// /*
// Copyright © 2022 NAME HERE <EMAIL ADDRESS>

// */
package cmd

import (
	mirror "ecr-mirror-sync/pkg/mirror"
	"time"

	log "github.com/sirupsen/logrus"

	"ecr-mirror-sync/pkg/options"

	"github.com/spf13/cobra"
)

var (
	ecrRespository   string
	upstreamImageTag string
)

func copyCmd() *cobra.Command {

	globalFlags, globalOpts := options.GlobalFlags()
	srcFlags, srcOpts := options.ImageFlags(globalOpts, "src-", "screds")
	destFlags, destOpts := options.ImageDestFlags(globalOpts, "dest-", "dcreds")
	retryFlags, retryOpts := options.RetryFlags()
	mirrorFlags, mirrorOpts := options.MirrorFlags(globalOpts, srcOpts, destOpts, retryOpts)

	copyCmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy image:tag from public source to ECR",
		Long:  `Copy image:tag from public source to ECR`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("copy requested")

			opts := mirrorOpts

			copy := mirror.New(opts)
			start := time.Now()
			copy.Copy(upstreamImageTag, ecrRespository)
			elapsed := time.Since(start)
			log.Infof("Copy took %s", elapsed)
		},
	}

	flags := copyCmd.Flags()
	flags.AddFlagSet(&globalFlags)
	flags.AddFlagSet(&destFlags)
	flags.AddFlagSet(&mirrorFlags)
	flags.AddFlagSet(&retryFlags)
	flags.AddFlagSet(&srcFlags)
	flags.StringVarP(&ecrRespository, "dest", "d", "", "ecr destingation repository")
	flags.StringVarP(&upstreamImageTag, "src", "s", "", "public Docker hub image:tag source")
	return copyCmd
}
