package topology

import (
	clabernetesapisv1alpha1 "github.com/srl-labs/clabernetes/apis/v1alpha1"
	clabernetesutil "github.com/srl-labs/clabernetes/util"
	clabernetesutilcontainerlab "github.com/srl-labs/clabernetes/util/containerlab"
	"gopkg.in/yaml.v3"
)

// ReconcileData is a struct that holds data that is common during a reconciliation process
// regardless of the type of clabernetes topology that is being reconciled.
type ReconcileData struct {
	Kind string

	PreviousHashes clabernetesapisv1alpha1.ReconcileHashes
	ResolvedHashes clabernetesapisv1alpha1.ReconcileHashes

	PreviousConfigs      map[string]*clabernetesutilcontainerlab.Config
	ResolvedConfigs      map[string]*clabernetesutilcontainerlab.Config
	ResolvedConfigsBytes []byte

	PreviousTunnels map[string][]*clabernetesapisv1alpha1.Tunnel
	ResolvedTunnels map[string][]*clabernetesapisv1alpha1.Tunnel

	ResolvedNodeExposedPorts map[string]*clabernetesapisv1alpha1.ExposedPorts

	ShouldUpdateResource bool
	NodesNeedingReboot   clabernetesutil.StringSet
}

// NewReconcileData accepts a TopologyCommonObject and returns a ReconcileData object.
func NewReconcileData(
	owningTopology *clabernetesapisv1alpha1.Topology,
) (*ReconcileData, error) {
	status := owningTopology.Status

	rd := &ReconcileData{
		PreviousHashes: status.ReconcileHashes,
		ResolvedHashes: clabernetesapisv1alpha1.ReconcileHashes{
			FilesFromURL: make(map[string]string),
		},

		PreviousConfigs: make(map[string]*clabernetesutilcontainerlab.Config),
		ResolvedConfigs: make(map[string]*clabernetesutilcontainerlab.Config),

		PreviousTunnels: status.Tunnels,
		ResolvedTunnels: make(map[string][]*clabernetesapisv1alpha1.Tunnel),

		ResolvedNodeExposedPorts: map[string]*clabernetesapisv1alpha1.ExposedPorts{},

		NodesNeedingReboot: clabernetesutil.NewStringSet(),
	}

	if status.Configs != "" {
		err := yaml.Unmarshal([]byte(status.Configs), &rd.PreviousConfigs)
		if err != nil {
			return nil, err
		}
	}

	return rd, nil
}

// SetStatus accepts a topology status and updates it with the ReconcileData information. This is
// called prior to updating a clabernetes topology object so that the hashes and information that
// we set in ReconcileData makes its way to the CR.
func (r *ReconcileData) SetStatus(
	owningTopologyStatus *clabernetesapisv1alpha1.TopologyStatus,
) {
	owningTopologyStatus.Kind = r.Kind
	owningTopologyStatus.Configs = string(r.ResolvedConfigsBytes)
	owningTopologyStatus.Tunnels = r.ResolvedTunnels
	owningTopologyStatus.NodeExposedPorts = r.ResolvedNodeExposedPorts

	owningTopologyStatus.ReconcileHashes = r.ResolvedHashes
}

// ConfigMapHasChanges returns true if the data that gets stored in the topology configmap has
// changed between the last reconcile and the current iteration. This is just a helper to be more
// verbose/clear what we are checking rather than having a giant conditional in the Reconciler.
func (r *ReconcileData) ConfigMapHasChanges() bool {
	if r.PreviousHashes.Config != r.ResolvedHashes.Config {
		return true
	}

	if r.PreviousHashes.Tunnels != r.ResolvedHashes.Tunnels {
		return true
	}

	if r.PreviousHashes.ImagePullSecrets != r.ResolvedHashes.ImagePullSecrets {
		return true
	}

	if r.NodesNeedingReboot.Len() != 0 {
		return true
	}

	return false
}
