package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
)

var rootCmd = &cobra.Command{
	Use:  "buildah-rmi",
	Long: "Remove oldest images from storage",
	RunE: func(cmd *cobra.Command, args []string) error {

	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		unshare.MaybeReexecUsingUserNamespace(false)
	},
}

func main() {
	options, err := storage.DefaultStoreOptions(unshare.IsRootless(), unshare.GetRootlessUID())
	if err != nil {
		fmt.Println(err)
	}

	store, err := storage.GetStore(options)
	if err != nil {
		fmt.Printf("storage.GetStore: %s", err)
	}
	if store != nil {
		is.Transport.SetStore(store)
	}

	images, err := store.Images()
	if err != nil {
		fmt.Printf("store.Images: %s", err)
	}

	for _, image := range images {
		// fmt.Printf("%+v\n", image)
		fmt.Printf("name: %+v, created: %+v\n", image.Names, image.Created)
	}
}
