package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"go.universe.tf/virtuakube"
)

var newclusterCmd = &cobra.Command{
	Use:   "newcluster",
	Short: "Create a Kubernetes cluster",
	Args:  cobra.NoArgs,
	Run:   withUniverse(&clusterFlags.universe, newcluster),
}

var clusterFlags = struct {
	universe   universeFlags
	name       string
	nodes      int
	image      string
	memory     int
	addons     []string
	networks   []string
	pushimages []string
}{}

func init() {
	rootCmd.AddCommand(newclusterCmd)
	addUniverseFlags(newclusterCmd, &clusterFlags.universe, true, false)
	addVMFlags(newclusterCmd)
	newclusterCmd.Flags().StringVar(&clusterFlags.name, "name", "", "name for the new cluster")
	newclusterCmd.Flags().IntVar(&clusterFlags.nodes, "nodes", 1, "number of nodes in the cluster")
	newclusterCmd.Flags().StringSliceVar(&clusterFlags.addons, "addons", nil, "addons to install")
	newclusterCmd.Flags().StringVar(&clusterFlags.image, "image", "", "base disk image to use")
	newclusterCmd.Flags().IntVar(&clusterFlags.memory, "memory", 1024, "amount of memory to give the VMs in GiB")
	newclusterCmd.Flags().StringSliceVar(&clusterFlags.networks, "networks", []string{}, "networks to attach the VM to")
	newclusterCmd.Flags().StringSliceVar(&clusterFlags.pushimages, "pushimages", []string{}, "docker images to push to cluster nodes")
}

func newcluster(u *virtuakube.Universe) error {
	cfg := &virtuakube.ClusterConfig{
		Name:     clusterFlags.name,
		NumNodes: clusterFlags.nodes,
		VMConfig: &virtuakube.VMConfig{
			Image:     clusterFlags.image,
			MemoryMiB: clusterFlags.memory,
			Networks:  clusterFlags.networks,
		},
	}

	fmt.Printf("Creating cluster %q...\n", clusterFlags.name)

	cluster, err := u.NewCluster(cfg)
	if err != nil {
		return fmt.Errorf("Creating cluster: %v", err)
	}
	if err = cluster.Start(); err != nil {
		return fmt.Errorf("Starting cluster: %v", err)
	}

	if len(clusterFlags.addons) != 0 {
		fmt.Printf("Installing addons %s...\n", strings.Join(clusterFlags.addons, ", "))
		for _, addon := range clusterFlags.addons {
			bs, err := ioutil.ReadFile(addon)
			if err != nil {
				return fmt.Errorf("Reading addon %q: %v", addon, err)
			}

			if err := cluster.ApplyManifest(bs); err != nil {
				return fmt.Errorf("installing addon %q: %v", addon, err)
			}
		}
	}

	if len(clusterFlags.pushimages) != 0 {
		fmt.Printf("Pushing docker images %s...\n", strings.Join(clusterFlags.pushimages, ", "))
		if err := cluster.PushImages(clusterFlags.pushimages...); err != nil {
			return fmt.Errorf("pushing images: %v", err)
		}
	}

	fmt.Printf("Created cluster %q\n", clusterFlags.name)

	return nil
}
