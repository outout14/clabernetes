package manager

import (
	"context"
	"fmt"
	"time"

	clabernetesapistopologyv1alpha1 "github.com/srl-labs/clabernetes/apis/topology/v1alpha1"
	clabernetesconstants "github.com/srl-labs/clabernetes/constants"
	clabernetescontrollers "github.com/srl-labs/clabernetes/controllers"
	clabernetescontrollerstopologycontainerlab "github.com/srl-labs/clabernetes/controllers/topology/containerlab"
	clabernetescontrollerstopologykne "github.com/srl-labs/clabernetes/controllers/topology/kne"
	clabernetesmanagerelection "github.com/srl-labs/clabernetes/manager/election"
	clabernetesmanagerprestart "github.com/srl-labs/clabernetes/manager/prestart"
	clabernetesutil "github.com/srl-labs/clabernetes/util"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	apimachineryscheme "k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	ctrlruntimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	ctrlruntimemetricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func (c *clabernetes) startLeading() {
	c.leaderElectionIdentity = clabernetesmanagerelection.GenerateLeaderIdentity()
	leaderElectionLockName := fmt.Sprintf("%s-manager", c.appName)

	leaderElectionLock := clabernetesmanagerelection.GetLeaseLock(
		c.kubeClient,
		c.appName,
		c.namespace,
		leaderElectionLockName,
		c.leaderElectionIdentity,
	)

	c.logger.Info("start leader election")
	clabernetesmanagerelection.RunElection(
		c.baseCtx,
		c.leaderElectionIdentity,
		leaderElectionLock,
		clabernetesmanagerelection.Timers{
			Duration:      electionDuration * time.Second,
			RenewDeadline: electionRenew * time.Second,
			RetryPeriod:   electionRetry * time.Second,
		},
		c.start,
		c.stopLeading,
		c.newLeader,
	)
}

func (c *clabernetes) stopLeading() {
	c.logger.Info("stopping clabernetes...")

	c.Exit(clabernetesconstants.ExitCode)
}

func (c *clabernetes) newLeader(newLeaderIdentity string) {
	c.logger.Infof("new leader elected '%s'", newLeaderIdentity)

	if newLeaderIdentity != c.leaderElectionIdentity {
		c.logger.Debug(
			"new leader is not us, nothing else for us to do. setting ready state to true",
		)

		c.ready = true
	} else {
		c.logger.Debug("new leader is us, resetting ready state to false")

		c.ready = false
	}
}

func mustNewManager(scheme *apimachineryruntime.Scheme, appName string) ctrlruntime.Manager {
	mgr, err := ctrlruntime.NewManager(
		ctrlruntime.GetConfigOrDie(),
		ctrlruntime.Options{
			Logger: klog.NewKlogr(),
			Scheme: scheme,
			Metrics: ctrlruntimemetricsserver.Options{
				BindAddress: "0",
			},
			LeaderElection: false,
			NewCache: func(
				config *rest.Config,
				opts ctrlruntimecache.Options,
			) (ctrlruntimecache.Cache, error) {
				opts.DefaultLabelSelector = labels.SelectorFromSet(
					labels.Set{
						// only cache objects with the "clabernetes/app" label, why would we care
						// about anything else (for now -- and we can override it with opts.ByObject
						// anyway?! and... who the hell calls their app "clabernetes" so this should
						// really limit the cache nicely :)
						"clabernetes/app": appName,
					},
				)

				return ctrlruntimecache.New(config, opts)
			},
		},
	)
	if err != nil {
		clabernetesutil.Panic(fmt.Sprintf("unable to start manager, error: %s", err))
	}

	return mgr
}

func (c *clabernetes) start(ctx context.Context) {
	c.leaderCtx = ctx

	c.logger.Info("begin pre-start...")

	clabernetesmanagerprestart.PreStart(c)

	c.logger.Debug("pre-start complete...")

	c.logger.Info("registering apis to scheme...")

	scheme := apimachineryruntime.NewScheme()

	apisToRegisterFuncs := []func() (apimachineryscheme.GroupVersion, []apimachineryruntime.Object){
		clabernetesapistopologyv1alpha1.GetAPIs,
	}

	for _, apiToRegisterFunc := range apisToRegisterFuncs {
		gv, objects := apiToRegisterFunc()

		for _, object := range objects {
			scheme.AddKnownTypes(gv, object)
		}

		metav1.AddToGroupVersion(scheme, gv)
	}

	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		clabernetesutil.Panic(err.Error())
	}

	err = apiextensionsv1.AddToScheme(scheme)
	if err != nil {
		clabernetesutil.Panic(err.Error())
	}

	c.logger.Debug("apis registered...")

	c.mgr = mustNewManager(scheme, c.appName)

	go func() {
		err = c.mgr.Start(c.leaderCtx)
		if err != nil {
			c.logger.Criticalf(
				"encountered error starting controller-runtime manager, err: %s",
				err,
			)

			c.Exit(clabernetesconstants.ExitCodeError)
		}
	}()

	c.logger.Info("begin syncing controller-runtime manager cache...")

	synced := c.mgr.GetCache().WaitForCacheSync(c.leaderCtx)
	if !synced {
		c.logger.Critical("encountered error syncing controller-runtime manager cache")

		c.Exit(clabernetesconstants.ExitCodeError)
	}

	c.logger.Debug("controller-runtime manager cache synced...")

	c.logger.Info("registering controllers...")

	controllersToRegisterFuncs := []clabernetescontrollers.NewController{
		clabernetescontrollerstopologycontainerlab.NewController,
		clabernetescontrollerstopologykne.NewController,
	}

	for _, newF := range controllersToRegisterFuncs {
		ctrl := newF(c.baseCtx, c.appName, c.kubeConfig, c.mgr.GetClient())

		clabernetesutil.MustSetupWithManager(ctrl.SetupWithManager, c.mgr)
	}

	c.logger.Debug("controllers registered...")

	c.ready = true

	c.logger.Debug("startup complete...")

	c.logger.Info("running forever or until interrupt...")

	<-c.leaderCtx.Done()
}
