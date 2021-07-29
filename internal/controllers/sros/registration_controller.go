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
	"fmt"
	"strconv"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	ndrv1 "github.com/netw-device-driver/ndd-core/apis/dvr/v1"
	"github.com/netw-device-driver/ndd-grpc/ndd"
	regclient "github.com/netw-device-driver/ndd-grpc/register/client"
	register "github.com/netw-device-driver/ndd-grpc/register/registerpb"
	srosv1 "github.com/netw-device-driver/ndd-provider-sros/apis/sros/v1"
	"github.com/netw-device-driver/ndd-runtime/pkg/event"
	"github.com/netw-device-driver/ndd-runtime/pkg/logging"
	"github.com/netw-device-driver/ndd-runtime/pkg/meta"
	"github.com/netw-device-driver/ndd-runtime/pkg/reconciler/managed"
	"github.com/netw-device-driver/ndd-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// Finalizer
	RegistrationFinalizer = "Registration.sros.ndd.henderiw.be"

	// Timers
	//reconcileTimeout = 1 * time.Minute
	//shortWait        = 30 * time.Second
	//veryShortWait    = 5 * time.Second

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

	// old erros
	//errGetRegistration             = "cannot get Registration"
	//errAddRegistrationFinalizer    = "cannot add Registration finalizer"
	//errRemoveRegistrationFinalizer = "cannot remove Registration finalizer"
	//errUpdateRegistrationStatus    = "cannot update Registration status"
	//errRegistrationFailed          = "cannot register to the device driver"
	//errDeRegistrationFailed        = "cannot desregister to the device driver"
	//errCacheStatusFailed           = "cannot get cache status from the device driver"

	// Event reasons
	//reasonSync event.Reason = "SyncRegistration"
)

/*
// RegistrationReconcilerOption is used to configure the RegistrationReconciler.
type RegistrationReconcilerOption func(*RegistrationReconciler)

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) RegistrationReconcilerOption {
	return func(r *RegistrationReconciler) {
		r.log = log
	}
}

// WithRecorder specifies how the Reconciler should record Kubernetes events.
func WithRecorder(er event.Recorder) RegistrationReconcilerOption {
	return func(r *RegistrationReconciler) {
		r.record = er
	}
}

// WithValidator specifies how the Reconciler should perform object
// validation.
func WithValidator(v Validator) RegistrationReconcilerOption {
	return func(r *RegistrationReconciler) {
		r.validator = v
	}
}

func WithGrpcApplicator(a gclient.Applicator) RegistrationReconcilerOption {
	return func(r *RegistrationReconciler) {
		r.applicator = a
	}
}

// RegistrationReconciler reconciles a Registration object
type RegistrationReconciler struct {
	client     client.Client
	finalizer  resource.Finalizer
	validator  Validator
	log        logging.Logger
	record     event.Recorder
	applicator gclient.Applicator
}
*/

// SetupRegistration adds a controller that reconciles Registrations.
func SetupRegistration(mgr ctrl.Manager, o controller.Options, l logging.Logger, poll time.Duration, namespace string) error {

	name := managed.ControllerName(srosv1.RegistrationGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(srosv1.RegistrationGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:        mgr.GetClient(),
			usage:       resource.NewNetworkNodeUsageTracker(mgr.GetClient(), &ndrv1.NetworkNodeUsage{}),
			newClientFn: regclient.NewClient},
		),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&srosv1.Registration{}).
		WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
		//Watches(
		//	&source.Kind{Type: &ndrv1.NetworkNode{}},
		//	handler.EnqueueRequestsFromMapFunc(r.NetworkNodeMapFunc),
		//).
		Complete(r)

	/*
		r := NewRegistrationReconciler(mgr,
			WithLogger(l.WithValues("controller", name)),
			WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			WithValidator(NewTargetValidator(resource.ClientApplicator{
				Client:     mgr.GetClient(),
				Applicator: resource.NewAPIPatchingApplicator(mgr.GetClient()),
			}, l)),
			WithGrpcApplicator(gclient.NewClientApplicator(
				gclient.WithLogger(l.WithValues("gclient", name)),
				gclient.WithInsecure(true),
				gclient.WithSkipVerify(true),
			)),
		)

		return ctrl.NewControllerManagedBy(mgr).
			Named(name).
			For(&srosv1.Registration{}).
			WithEventFilter(resource.IgnoreUpdateWithoutGenerationChangePredicate()).
			WithOptions(option).
			Watches(
				&source.Kind{Type: &ndrv1.NetworkNode{}},
				handler.EnqueueRequestsFromMapFunc(r.NetworkNodeMapFunc),
			).
			Complete(r)
	*/
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube        client.Client
	usage       resource.Tracker
	newClientFn func(ctx context.Context, creds ndd.Config) (register.RegistrationClient, error)
}

// Connect produces an ExternalClient by:
// 1. Tracking that the managed resource is using a TargetConfig.
// 2. Getting the managed resource's TargetConfig with connection details
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*srosv1.Registration)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackTCUsage)
	}

	//pc := &v1.TargetConfig{}
	//if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetTargetConfigReference().Name}, pc); err != nil {
	//	return nil, errors.Wrap(err, errGetTC)
	//}

	nn := &ndrv1.NetworkNode{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetNetworkNodeReference().Name}, nn); err != nil {
		return nil, errors.Wrap(err, errGetNetworkNode)
	}

	if nn.GetCondition(ndrv1.ConditionKindDeviceDriverConfigured).Status != corev1.ConditionTrue {
		return nil, errors.New(targetNotConfigured)
	}

	cl, err := c.newClientFn(ctx, ndd.Config{
		SkipVerify: true,
		Insecure:   true,
		Target:     ndrv1.PrefixService + "-" + nn.Name + "." + ndrv1.NamespaceLocalK8sDNS + strconv.Itoa(*nn.Spec.GrpcServerPort),
	})
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{client: cl}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client register.RegistrationClient
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

	// These fmt statements should be removed in the real implementation.
	fmt.Printf("Observing: %+v", o)

	r, err := e.client.Read(ctx, &register.DeviceType{
		DeviceType: string(srosv1.DeviceTypeSROS),
	})
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errRegistrationRead)
	}

	if r.DeviceType == string(srosv1.DeviceTypeSROS) {
		return managed.ExternalObservation{
			ResourceExists:    true,
			ResourceUpToDate:  true,
			ConnectionDetails: managed.ConnectionDetails{},
		}, nil
	} else {
		return managed.ExternalObservation{
			ResourceExists:    false,
			ResourceUpToDate:  false,
			ConnectionDetails: managed.ConnectionDetails{},
		}, nil
	}
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	o, ok := mg.(*srosv1.Registration)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	fmt.Printf("Creating: %+v", o)

	_, err := e.client.Create(ctx, &register.RegistrationInfo{
		DeviceType:             string(srosv1.DeviceTypeSROS),
		MatchString:            srosv1.DeviceMatch,
		Subscriptions:          o.GetSubscriptions(),
		ExcpetionPaths:         o.GetExceptionPaths(),
		ExplicitExceptionPaths: o.GetExplicitExceptionPaths(),
	})
	if err != nil {
		return managed.ExternalCreation{}, errors.New(errRegistrationCreate)
	}

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	o, ok := mg.(*srosv1.Registration)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	fmt.Printf("Updating: %+v", o)

	_, err := e.client.Update(ctx, &register.RegistrationInfo{
		DeviceType:             string(srosv1.DeviceTypeSROS),
		MatchString:            srosv1.DeviceMatch,
		Subscriptions:          o.GetSubscriptions(),
		ExcpetionPaths:         o.GetExceptionPaths(),
		ExplicitExceptionPaths: o.GetExplicitExceptionPaths(),
	})
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errRegistrationUpdate)
	}

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	o, ok := mg.(*srosv1.Registration)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	fmt.Printf("Deleting: %+v", o)

	_, err := e.client.Delete(ctx, &register.DeviceType{
		DeviceType: string(srosv1.DeviceTypeSROS),
	})
	if err != nil {
		return errors.New(errRegistrationDelete)
	}

	return nil
}

/*
// NewRegistrationReconciler creates a new package revision reconciler.
func NewRegistrationReconciler(mgr manager.Manager, opts ...RegistrationReconcilerOption) *RegistrationReconciler {
	r := &RegistrationReconciler{
		client:    mgr.GetClient(),
		finalizer: resource.NewAPIFinalizer(mgr.GetClient(), RegistrationFinalizer),
		log:       logging.NewNopLogger(),
		record:    event.NewNopRecorder(),
	}
	for _, f := range opts {
		f(r)
	}
	return r
}

// NetworkNodeMapFunc is a handler.ToRequestsFunc to be used to enqeue
// request for reconciliation of Registration.
func (r *RegistrationReconciler) NetworkNodeMapFunc(o client.Object) []ctrl.Request {
	log := r.log.WithValues("NetworkNode Object", o)
	result := []ctrl.Request{}

	nn, ok := o.(*ndrv1.NetworkNode)
	if !ok {
		panic(fmt.Sprintf("Expected a NodeTopology but got a %T", o))
	}
	log.WithValues(nn.GetName(), nn.GetNamespace()).Info("NetworkNode MapFunction")

	selectors := []client.ListOption{
		client.InNamespace(nn.Namespace),
		client.MatchingLabels{},
	}
	ss := &srosv1.RegistrationList{}
	if err := r.client.List(context.TODO(), ss, selectors...); err != nil {
		return result
	}

	for _, o := range ss.Items {
		name := client.ObjectKey{
			Namespace: o.GetNamespace(),
			Name:      o.GetName(),
		}
		log.WithValues(o.GetName(), o.GetNamespace()).Info("NetworkNodeMapFunc Registration ReQueue")
		result = append(result, ctrl.Request{NamespacedName: name})
	}
	return result
}

//+kubebuilder:rbac:groups=dvr.ndd.henderiw.be,resources=networknodes,verbs=get;list;watch
//+kubebuilder:rbac:groups=sros.ndd.henderiw.be,resources=registrations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sros.ndd.henderiw.be,resources=registrations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sros.ndd.henderiw.be,resources=registrations/finalizers,verbs=update

func (r *RegistrationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//log := r.log.FromContext(ctx)
	log := r.log.WithValues("Registration", req.NamespacedName)
	log.Debug("reconciling Registration")

	o := &srosv1.Registration{}
	if err := r.client.Get(ctx, req.NamespacedName, o); err != nil {
		// There's no need to requeue if we no longer exist. Otherwise we'll be
		// requeued implicitly because we return an error.
		log.Debug(errGetRegistration, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetRegistration)
	}
	log.Debug("Registration", "Sub", o)

	if meta.WasDeleted(o) {
		// find the available targets and deregister the provider
		targets, err := r.validator.FindConfigured(ctx, o.Spec.TargetConfigReference.Name)
		if err != nil {
			log.Debug(nddv1.ErrFindingTargets, "error", err)
			r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, nddv1.ErrFindingTargets)))
			return reconcile.Result{RequeueAfter: reconcileTimeout}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)
		}
		for _, target := range targets {
			// register the provider to the device driver
			_, err := r.applicator.DeRegister(ctx, target.DNS,
				&netwdevpb.RegistrationRequest{
					DeviceType:  string(srosv1.DeviceTypesros),
					MatchString: srosv1.DeviceMatch,
				})
			if err != nil {
				log.Debug(errDeRegistrationFailed)
				r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, fmt.Sprintf("%s, target, %s", errDeRegistrationFailed, target))))
				return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(err, fmt.Sprintf("%s, target, %s", errDeRegistrationFailed, target))
			}
		}

		// Delete finalizer after all device drivers are deregistered
		if err := r.finalizer.RemoveFinalizer(ctx, o); err != nil {
			log.Debug(errRemoveRegistrationFinalizer, "error", err)
			r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, errRemoveRegistrationFinalizer)))
			return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(err, errRemoveRegistrationFinalizer)
		}
		return reconcile.Result{Requeue: false}, nil
	}

	if err := r.finalizer.AddFinalizer(ctx, o); err != nil {
		log.Debug(errAddRegistrationFinalizer, "error", err)
		r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, errAddRegistrationFinalizer)))
		return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(err, errAddRegistrationFinalizer)
	}

	log = log.WithValues(
		"uid", o.GetUID(),
		"version", o.GetResourceVersion(),
		"name", o.GetName(),
	)

	// find targets the resource should be applied to
	targets, err := r.validator.FindConfigured(ctx, o.Spec.TargetConfigReference.Name)
	if err != nil {
		log.Debug(nddv1.ErrTargetNotFound, "error", err)
		r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, nddv1.ErrTargetNotFound)))
		o.SetConditions(nddv1.TargetNotFound())
		return reconcile.Result{RequeueAfter: reconcileTimeout}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)
	}

	// check if targets got deleted and if so delete them from the status
	for targetName := range o.Status.TargetConditions {
		if r.validator.IfDeleted(ctx, targets, targetName) {
			o.DeleteTargetCondition(targetName)
			log.Debug(nddv1.InfoTargetDeleted, "target", targetName)
			r.record.Event(o, event.Normal(reasonSync, nddv1.InfoTargetDeleted, "target", targetName))
		}
	}

	// if no targets are found we return and update object status with no target found
	if len(targets) == 0 {
		log.Debug(nddv1.ErrTargetNotFound)
		r.record.Event(o, event.Warning(reasonSync, errors.New(nddv1.ErrTargetNotFound)))
		o.SetConditions(nddv1.TargetNotFound())
		return reconcile.Result{RequeueAfter: reconcileTimeout}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)
	}
	// otherwise register the provider to the device driver
	r.record.Event(o, event.Normal(reasonSync, nddv1.InfoTargetFound))
	o.SetConditions(nddv1.TargetFound())

	for _, target := range targets {
		log.Debug("RegistrationTargets", "target", target)
		// initialize the targets ConditionedStatus
		if len(o.Status.TargetConditions) == 0 {
			o.InitializeTargetConditions()
		}

		// set condition if not yet set
		if _, ok := o.Status.TargetConditions[target.Name]; !ok {
			log.Debug("Initialize target condition")
			o.SetTargetConditions(target.Name, nddv1.Unknown())
		}

		// update the device driver with the object info when the condition is not met yet
		log.Debug("targetconditionStatus", "status", o.GetTargetCondition(target.Name, nddv1.ConditionKindConfiguration).Status)
		if o.GetTargetCondition(target.Name, nddv1.ConditionKindConfiguration).Status == corev1.ConditionFalse ||
			o.GetTargetCondition(target.Name, nddv1.ConditionKindConfiguration).Status == corev1.ConditionUnknown {

			// register the provider to the device driver
			rsp, err := r.applicator.Register(ctx, target.DNS,
				&netwdevpb.RegistrationRequest{
					DeviceType:             string(srosv1.DeviceTypesros),
					MatchString:            srosv1.DeviceMatch,
					Subscriptions:          o.GetSubscriptions(),
					ExcpetionPaths:         o.GetExceptionPaths(),
					ExplicitExceptionPaths: o.GetExplicitExceptionPaths(),
				})
			if err != nil {
				log.Debug(errRegistrationFailed)
				r.record.Event(o, event.Warning(reasonSync, errors.Wrap(err, errRegistrationFailed)))
				return reconcile.Result{RequeueAfter: shortWait}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)

			}
			log.Debug("Register", "response", rsp)
			log.Debug("Object status", "status", o.Status)
			o.SetTargetConditions(target.Name, nddv1.ReconcileSuccess())
		}
	}
	// update the status and return
	return reconcile.Result{Requeue: false, RequeueAfter: reconcileTimeout}, errors.Wrap(r.client.Status().Update(ctx, o), errUpdateRegistrationStatus)
}
*/
