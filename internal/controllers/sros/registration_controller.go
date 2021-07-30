/*
Copyright 2021 Wim Henderickx.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sros

import (
	"context"
	"strconv"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/source"

	ndrv1 "github.com/netw-device-driver/ndd-core/apis/dvr/v1"
	"github.com/netw-device-driver/ndd-grpc/ndd"
	regclient "github.com/netw-device-driver/ndd-grpc/register/client"
	register "github.com/netw-device-driver/ndd-grpc/register/registerpb"
	srosv1 "github.com/netw-device-driver/ndd-provider-sros/apis/sros/v1"
	nddv1 "github.com/netw-device-driver/ndd-runtime/apis/common/v1"
	"github.com/netw-device-driver/ndd-runtime/pkg/event"
	"github.com/netw-device-driver/ndd-runtime/pkg/logging"
	"github.com/netw-device-driver/ndd-runtime/pkg/meta"
	"github.com/netw-device-driver/ndd-runtime/pkg/reconciler/managed"
	"github.com/netw-device-driver/ndd-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	// Finalizer
	RegistrationFinalizer = "Registration.sros.ndd.henderiw.be"

	// Errors
	errUnexpectedObject   = "the managed resource is not a Registration resource"
	errTrackTCUsage       = "cannot track TargetConfig usage"
	errGetTC              = "cannot get TargetConfig"
	errGetNetworkNode     = "cannot get NetworkNode"
	errNewClient          = "cannot create new client"
	errKubeUpdateFailed   = "cannot update Registration custom resource"
	targetNotConfigured   = "target is not configured to proceed"
	errRegistrationRead   = "cannot read Registration"
	errRegistrationCreate = "cannot create Registration"
	errRegistrationUpdate = "cannot update Registration"
	errRegistrationDelete = "cannot delete Registration"
	errNoTargetFound      = "target not found"
)

// SetupRegistration adds a controller that reconciles Registrations.
func SetupRegistration(mgr ctrl.Manager, o controller.Options, l logging.Logger, poll time.Duration, namespace string) error {

	name := managed.ControllerName(srosv1.RegistrationGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(srosv1.RegistrationGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			log:         l,
			kube:        mgr.GetClient(),
			usage:       resource.NewNetworkNodeUsageTracker(mgr.GetClient(), &ndrv1.NetworkNodeUsage{}),
			newClientFn: regclient.NewClient},
		),
		managed.WithValidator(&validator{log: l}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&srosv1.Registration{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		Watches(
			&source.Kind{Type: &ndrv1.NetworkNode{}},
			&resource.EnqueueRequestForNetworkNode{}).
		Complete(r)
}

type validator struct {
	log logging.Logger
}

func (v *validator) ValidateLocalleafRef(ctx context.Context, mg resource.Managed) (managed.ValidateLocalleafRefObservation, error) {
	return managed.ValidateLocalleafRefObservation{Success: true}, nil
}

func (v *validator) ValidateExternalleafRef(ctx context.Context, mg resource.Managed, cfg []byte) (managed.ValidateExternalleafRefObservation, error) {
	return managed.ValidateExternalleafRefObservation{Success: true}, nil
}

func (v *validator) ValidateParentDependency(ctx context.Context, mg resource.Managed, cfg []byte) (managed.ValidationParentDependencyObservation, error) {
	return managed.ValidationParentDependencyObservation{Success: true}, nil
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	log         logging.Logger
	kube        client.Client
	usage       resource.Tracker
	newClientFn func(ctx context.Context, cfg ndd.Config) (register.RegistrationClient, error)
}

// Connect produces an ExternalClient by:
// 1. Tracking that the managed resource is using a TargetConfig.
// 2. Getting the managed resource's TargetConfig with connection details
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	log := c.log.WithValues("resosurce", mg.GetName())
	log.Debug("Connect")
	_, ok := mg.(*srosv1.Registration)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackTCUsage)
	}

	selectors := []client.ListOption{}
	nnl := &ndrv1.NetworkNodeList{}
	if err := c.kube.List(ctx, nnl, selectors...); err != nil {
		return nil, errors.Wrap(err, errGetNetworkNode)
	}

	// find all targets that have are in configured status
	var ts []*nddv1.Target
	for _, nn := range nnl.Items {
		log.Debug("Network Node", "Name", nn.GetName(), "Status", nn.GetCondition(ndrv1.ConditionKindDeviceDriverConfigured).Status)
		if nn.GetCondition(ndrv1.ConditionKindDeviceDriverConfigured).Status == corev1.ConditionTrue {
			t := &nddv1.Target{
				Name: nn.GetName(),
				Cfg: ndd.Config{
					SkipVerify: true,
					Insecure:   true,
					Target:     ndrv1.PrefixService + "-" + nn.Name + "." + ndrv1.NamespaceLocalK8sDNS + strconv.Itoa(*nn.Spec.GrpcServerPort),
				},
			}
			ts = append(ts, t)
		}
	}
	log.Debug("Active targets", "targets", ts)

	// when no targets are found we return a not found error
	// this unifies the reconcile code when a dedicate network node is looked up
	if len(ts) == 0 {
		return nil, errors.New(errNoTargetFound)
	}

	// We dont have to update the deletion since the network device driver would have lost
	// its state already

	//get clients for each target
	cls := make([]register.RegistrationClient, 0)
	tns := make([]string, 0)
	for _, t := range ts {
		cl, err := c.newClientFn(ctx, t.Cfg)
		if err != nil {
			return nil, errors.Wrap(err, errNewClient)
		}
		cls = append(cls, cl)
		tns = append(tns, t.Name)
	}

	log.Debug("Connect info", "clients", cls, "targets", tns)

	return &external{clients: cls, targets: tns, log: log}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	clients []register.RegistrationClient
	targets []string
	log     logging.Logger
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	o, ok := mg.(*srosv1.Registration)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(o) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	for _, cl := range e.clients {
		r, err := cl.Get(ctx, &register.DeviceType{
			DeviceType: string(srosv1.DeviceTypeSROS),
		})
		if err != nil {
			// if a single network device driver reports an error this is applicable to all
			// network devices
			return managed.ExternalObservation{}, errors.New(errRegistrationRead)
		}
		// if a network device driver reports a different device type we trigger
		// a recreation of the configuration on all devices by returning
		// Exists = false and
		if r.DeviceType != string(srosv1.DeviceTypeSROS) {
			return managed.ExternalObservation{
				ResourceExists:    false,
				ResourceUpToDate:  false,
				ConnectionDetails: managed.ConnectionDetails{},
			}, nil
		}
	}

	// when all network device driver reports the proper device type
	// we return exists and up to date
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil

}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*srosv1.Registration)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	e.log.Debug("creating", "object", cr)

	for _, cl := range e.clients {
		_, err := cl.Create(ctx, &register.Request{
			DeviceType:             string(srosv1.DeviceTypeSROS),
			MatchString:            srosv1.DeviceMatch,
			Subscriptions:          cr.GetSubscriptions(),
			ExceptionPaths:         cr.GetExceptionPaths(),
			ExplicitExceptionPaths: cr.GetExplicitExceptionPaths(),
		})
		if err != nil {
			return managed.ExternalCreation{}, errors.New(errRegistrationCreate)
		}
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*srosv1.Registration)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	e.log.Debug("creating", "object", cr)

	for _, cl := range e.clients {
		_, err := cl.Update(ctx, &register.Request{
			DeviceType:             string(srosv1.DeviceTypeSROS),
			MatchString:            srosv1.DeviceMatch,
			Subscriptions:          cr.GetSubscriptions(),
			ExceptionPaths:         cr.GetExceptionPaths(),
			ExplicitExceptionPaths: cr.GetExplicitExceptionPaths(),
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.New(errRegistrationUpdate)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*srosv1.Registration)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	e.log.Debug("deleting", "object", cr)

	for _, cl := range e.clients {
		_, err := cl.Delete(ctx, &register.DeviceType{
			DeviceType: string(srosv1.DeviceTypeSROS),
		})
		if err != nil {
			return errors.New(errRegistrationDelete)
		}
	}
	return nil
}

func (e *external) GetTarget() []string {
	return e.targets
}

func (e *external) GetConfig(ctx context.Context) ([]byte, error) {
	return make([]byte, 0), nil
}
