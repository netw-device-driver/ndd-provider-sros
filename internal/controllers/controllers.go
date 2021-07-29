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

package controllers

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/netw-device-driver/ndd-provider-sros/internal/controllers/sros"
	"github.com/netw-device-driver/ndd-runtime/pkg/logging"
)

// Setup package controllers.
func Setup(mgr ctrl.Manager, option controller.Options, l logging.Logger, poll time.Duration, namespace string) error {
	for _, setup := range []func(ctrl.Manager, controller.Options, logging.Logger, time.Duration, string) error{
		sros.SetupRegistration,
	} {
		if err := setup(mgr, option, l, poll, namespace); err != nil {
			return err
		}
	}
	return nil
	//return config.Setup(mgr, l, option)
}
