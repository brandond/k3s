package etcd

import (
	"context"
	"os"
	"time"

	controllerv1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	nodeID       = "etcd.k3s.cattle.io/node-name"
	nodeAddress  = "etcd.k3s.cattle.io/node-address"
	master       = "node-role.kubernetes.io/master"
	controlPlane = "node-role.kubernetes.io/control-plane"
	etcdRole     = "node-role.kubernetes.io/etcd"
)

type NodeControllerGetter func() controllerv1.NodeController

type handler struct {
	etcd           *ETCD
	nodeController controllerv1.NodeController
	ctx            context.Context
}

// Register adds callbacks for node events.
func Register(ctx context.Context, etcd *ETCD, nodes controllerv1.NodeController) {
	h := &handler{
		etcd:           etcd,
		nodeController: nodes,
		ctx:            ctx,
	}
	nodes.OnChange(ctx, "managed-etcd-controller", h.sync)
}

// sync is the callback for node change events. If the changed node
// is the local node, it calls handlSelf to ensure annotations and labels.
func (h *handler) sync(key string, node *v1.Node) (*v1.Node, error) {
	if node == nil {
		return nil, nil
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		logrus.Debug("waiting for node to be assigned for etcd controller")
		h.nodeController.EnqueueAfter(key, 5*time.Second)
		return node, nil
	}

	if key == nodeName {
		return h.handleSelf(node)
	}

	return node, nil
}

// handleSelf is called if the changed node is the local node.
// It ensures that the etcd annotations and role labels are present.
func (h *handler) handleSelf(node *v1.Node) (*v1.Node, error) {
	if node.Annotations[nodeID] == h.etcd.name &&
		node.Annotations[nodeAddress] == h.etcd.address &&
		node.Labels[etcdRole] == "true" &&
		node.Labels[controlPlane] == "true" {
		return node, nil
	}

	node = node.DeepCopy()
	if node.Annotations == nil {
		node.Annotations = map[string]string{}
	}
	node.Annotations[nodeID] = h.etcd.name
	node.Annotations[nodeAddress] = h.etcd.address
	node.Labels[etcdRole] = "true"
	node.Labels[master] = "true"
	node.Labels[controlPlane] = "true"

	return h.nodeController.Update(node)
}
