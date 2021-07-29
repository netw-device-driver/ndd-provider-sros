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

package config

/*
import (
	"github.com/netw-device-driver/ndd-runtime/pkg/event"
	"github.com/netw-device-driver/ndd-runtime/pkg/logging"
	"github.com/netw-device-driver/ndd-runtime/pkg/reconciler/targetconfig"
	"github.com/netw-device-driver/ndd-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1 "github.com/netw-device-driver/ndd-provider-sros/apis/v1"
)


// Setup adds a controller that reconciles TargetConfigs by accounting for
// their current usage.
func Setup(mgr ctrl.Manager, l logging.Logger, o controller.Options) error {
	name := targetconfig.ControllerName(v1.TargetConfigGroupKind)

	of := resource.TargetConfigKinds{
		Config:    v1.TargetConfigGroupVersionKind,
		UsageList: v1.TargetConfigUsageListGroupVersionKind,
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1.TargetConfig{}).
		Watches(&source.Kind{Type: &v1.TargetConfigUsage{}}, &resource.EnqueueRequestForTargetConfig{}).
		Complete(targetconfig.NewReconciler(mgr, of,
			targetconfig.WithLogger(l.WithValues("controller", name)),
			targetconfig.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}
*/
