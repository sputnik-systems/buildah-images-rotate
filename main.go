package main

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
)

type flags struct {
	Names  []string
	Older  string
	DryRun bool
}

var (
	cliFlags flags
)

var rootCmd = &cobra.Command{
	Use:  "buildah-rmi",
	Long: "Remove oldest images from storage",
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeImages()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		unshare.MaybeReexecUsingUserNamespace(false)
	},
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&cliFlags.Names, "name", []string{}, "image name, that should be rotated")
	rootCmd.PersistentFlags().StringVar(&cliFlags.Older, "older-than", "744h", "images older than specified date will be deleteted")
	rootCmd.PersistentFlags().BoolVar(&cliFlags.DryRun, "dry-run", false, "dry-run")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("an error has occured: %s\n", err)
	}
}

func removeImages() error {
	d, err := time.ParseDuration(cliFlags.Older)
	if err != nil {
		return err
	}

	options, err := storage.DefaultStoreOptions(unshare.IsRootless(), unshare.GetRootlessUID())
	if err != nil {
		return err
	}

	store, err := storage.GetStore(options)
	if err != nil {
		return err
	}
	if store != nil {
		is.Transport.SetStore(store)
	}

	images, err := store.Images()
	if err != nil {
		return err
	}

	olderThan := time.Now().Add(-d)

	var imagesToDelete []storage.Image
	for _, image := range images {
		if image.Created.Before(olderThan) {
			// dangling images
			if len(image.Names) == 0 {
				imagesToDelete = append(imagesToDelete, image)
			}

			for _, exName := range cliFlags.Names {
				for _, name := range image.Names {
					if strings.Contains(name, exName) {
						imagesToDelete = append(imagesToDelete, image)
					}
				}
			}
		}
	}

	for _, image := range imagesToDelete {
		layers, err := store.DeleteImage(image.ID, !cliFlags.DryRun)
		if err != nil {
			log.Infof("image with ID %s not removed. Error: %s", image.ID, err)
		}

		if len(image.Names) > 0 {
			log.Infof("image removed: %s, image names: %v, created: %s\n", image.ID, image.Names, image.Created)
		} else {
			log.Infof("image removed: %s, created: %s\n", image.ID, image.Created)
		}

		if len(layers) > 0 {
			log.Infof("layers removed: %v\n", layers)
		}
	}

	return nil
}
