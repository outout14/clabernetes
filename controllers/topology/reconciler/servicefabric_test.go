package reconciler_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	clabernetesconstants "github.com/srl-labs/clabernetes/constants"
	clabernetesutil "github.com/srl-labs/clabernetes/util"
	clabernetesutilcontainerlab "github.com/srl-labs/clabernetes/util/containerlab"

	clabernetesapistopology "github.com/srl-labs/clabernetes/apis/topology"
	clabernetesapistopologyv1alpha1 "github.com/srl-labs/clabernetes/apis/topology/v1alpha1"
	clabernetesconfig "github.com/srl-labs/clabernetes/config"
	clabernetescontrollerstopologyreconciler "github.com/srl-labs/clabernetes/controllers/topology/reconciler"
	claberneteslogging "github.com/srl-labs/clabernetes/logging"
	clabernetestesthelper "github.com/srl-labs/clabernetes/testhelper"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const renderServiceFabricTestName = "servicefabric/render-service"

func TestResolveServiceFabric(t *testing.T) {
	cases := []struct {
		name               string
		ownedServices      *k8scorev1.ServiceList
		clabernetesConfigs map[string]*clabernetesutilcontainerlab.Config
		expectedCurrent    []string
		expectedMissing    []string
		expectedExtra      []*k8scorev1.Service
	}{
		{
			name:               "simple",
			ownedServices:      &k8scorev1.ServiceList{},
			clabernetesConfigs: nil,
			expectedCurrent:    nil,
			expectedMissing:    nil,
			expectedExtra:      []*k8scorev1.Service{},
		},
		{
			name:          "missing-nodes",
			ownedServices: &k8scorev1.ServiceList{},
			clabernetesConfigs: map[string]*clabernetesutilcontainerlab.Config{
				"node1": nil,
				"node2": nil,
			},
			expectedCurrent: nil,
			expectedMissing: []string{"node1", "node2"},
			expectedExtra:   []*k8scorev1.Service{},
		},
		{
			name: "extra-nodes",
			ownedServices: &k8scorev1.ServiceList{
				Items: []k8scorev1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "resolve-servicefabric-test",
							Namespace: "clabernetes",
							Labels: map[string]string{
								clabernetesconstants.LabelTopologyServiceType: clabernetesconstants.TopologyServiceTypeFabric, //nolint:lll
								clabernetesconstants.LabelTopologyNode:        "node2",
							},
						},
					},
				},
			},
			clabernetesConfigs: map[string]*clabernetesutilcontainerlab.Config{
				"node1": nil,
			},
			expectedCurrent: nil,
			expectedMissing: nil,
			expectedExtra: []*k8scorev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "resolve-servicefabric-test",
						Namespace: "clabernetes",
						Labels: map[string]string{
							clabernetesconstants.LabelTopologyServiceType: clabernetesconstants.TopologyServiceTypeFabric, //nolint:lll
							clabernetesconstants.LabelTopologyNode:        "node2",
						},
					},
				},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Logf("%s: starting", testCase.name)

				reconciler := clabernetescontrollerstopologyreconciler.NewServiceFabricReconciler(
					&claberneteslogging.FakeInstance{},
					clabernetesapistopology.Containerlab,
					clabernetesconfig.GetFakeManager,
				)

				got, err := reconciler.Resolve(
					testCase.ownedServices,
					testCase.clabernetesConfigs,
					nil,
				)
				if err != nil {
					t.Fatal(err)
				}

				var gotCurrent []string

				for current := range got.Current {
					gotCurrent = append(gotCurrent, current)
				}

				if !clabernetesutil.StringSliceContainsAll(gotCurrent, testCase.expectedCurrent) {
					clabernetestesthelper.FailOutput(t, gotCurrent, testCase.expectedCurrent)
				}

				if !clabernetesutil.StringSliceContainsAll(got.Missing, testCase.expectedMissing) {
					clabernetestesthelper.FailOutput(t, got.Missing, testCase.expectedMissing)
				}

				if !reflect.DeepEqual(got.Extra, testCase.expectedExtra) {
					clabernetestesthelper.FailOutput(t, got.Extra, testCase.expectedExtra)
				}
			})
	}
}

func TestRenderServiceFabric(t *testing.T) {
	cases := []struct {
		name                 string
		owningTopologyObject clabernetesapistopologyv1alpha1.TopologyCommonObject
		nodeName             string
	}{
		{
			name: "simple",
			owningTopologyObject: &clabernetesapistopologyv1alpha1.Containerlab{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "render-service-fabric-test",
					Namespace: "clabernetes",
				},
				Spec: clabernetesapistopologyv1alpha1.ContainerlabSpec{
					TopologyCommonSpec: clabernetesapistopologyv1alpha1.TopologyCommonSpec{},
					Config: `---
    name: test
    topology:
      nodes:
        srl1:
          kind: srl
          image: ghcr.io/nokia/srlinux
`,
				},
			},
			nodeName: "srl1",
		},
	}

	for _, testCase := range cases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Logf("%s: starting", testCase.name)

				reconciler := clabernetescontrollerstopologyreconciler.NewServiceFabricReconciler(
					&claberneteslogging.FakeInstance{},
					clabernetesapistopology.Containerlab,
					clabernetesconfig.GetFakeManager,
				)

				got := reconciler.Render(
					testCase.owningTopologyObject,
					testCase.nodeName,
				)

				if *clabernetestesthelper.Update {
					clabernetestesthelper.WriteTestFixtureJSON(
						t,
						fmt.Sprintf(
							"golden/%s/%s.json",
							renderServiceFabricTestName,
							testCase.name,
						),
						got,
					)
				}

				var want k8scorev1.Service

				err := json.Unmarshal(
					clabernetestesthelper.ReadTestFixtureFile(
						t,
						fmt.Sprintf(
							"golden/%s/%s.json",
							renderServiceFabricTestName,
							testCase.name,
						),
					),
					&want,
				)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(got.Annotations, want.Annotations) {
					clabernetestesthelper.FailOutput(t, got.Annotations, want.Annotations)
				}
				if !reflect.DeepEqual(got.Labels, want.Labels) {
					clabernetestesthelper.FailOutput(t, got.Labels, want.Labels)
				}
				if !reflect.DeepEqual(got.Spec, want.Spec) {
					clabernetestesthelper.FailOutput(t, got.Spec, want.Spec)
				}
			})
	}
}