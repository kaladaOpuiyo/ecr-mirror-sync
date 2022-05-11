/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"ecr-mirror-sync/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {
	cmd.Execute()
}
func init() {

	//Set the logn format
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		FieldMap: log.FieldMap{
			log.FieldKeyTime:  "timestamp",
			log.FieldKeyLevel: "level",
			log.FieldKeyMsg:   "message",
		},
	},
	)

}
