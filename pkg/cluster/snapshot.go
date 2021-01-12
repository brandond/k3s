package cluster

import (
	"context"
	"errors"

	"github.com/rancher/k3s/pkg/daemons/config"
)

// Snapshot is a proxy method to call the snapshot method on the managedb
// interface for etcd clusters.
func (c *Cluster) Snapshot(ctx context.Context, config *config.Control) error {
	if c.managedDB == nil {
		return errors.New("unable to perform etcd snapshot on non-etcd system")
	}

	return c.managedDB.Snapshot(ctx, config)
}
