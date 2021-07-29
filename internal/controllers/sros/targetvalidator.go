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

/*
import (
	"context"
	"strconv"

	ndrv1 "github.com/netw-device-driver/ndd-core/apis/dvr/v1"
	nddv1 "github.com/netw-device-driver/ndd-runtime/apis/common/v1"
	"github.com/netw-device-driver/ndd-runtime/pkg/logging"
	"github.com/netw-device-driver/ndd-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	// Errors
	errFailedListNetworkNodes             = "failed to list network nodes"
	errNoTargetLabelsPresentOnResource    = "failed to get labels with key target on the resource object"
	errNoTargetLabelsPresentOnNetworkNode = "failed to get labels with key target on the networkNode"
)

type Validator interface {
	// FindHealthy returns the target/targets in healthy state
	FindHealthy(ctx context.Context, targetName string) ([]nddv1.Target, error)

	// FindConfigured returns the target/targets in configured state
	// this is the target state before the ready state
	FindConfigured(ctx context.Context, targetName string) ([]nddv1.Target, error)

	// FindReady returns the target/targets in ready state
	FindReady(ctx context.Context, targetName string) ([]nddv1.Target, error)

	// FindTargetWithStatus returns the target/targets in the respective state
	FindTargetWithStatus(ctx context.Context, targetName string, c nddv1.ConditionKind) ([]nddv1.Target, error)

	// Check if the target got deleted
	IfDeleted(ctx context.Context, targets []nddv1.Target, target string) bool
}

type TargetValidator struct {
	client resource.ClientApplicator
	log    logging.Logger
}

func NewTargetValidator(client resource.ClientApplicator, log logging.Logger) *TargetValidator {
	return &TargetValidator{
		client: client,
		log:    log,
	}
}

func (t *TargetValidator) FindHealthy(ctx context.Context, targetName string) (targets []nddv1.Target, err error) {
	return t.FindTargetWithStatus(ctx, targetName, ndrv1.ConditionKindDeviceDriverHealthy)
}

func (t *TargetValidator) FindConfigured(ctx context.Context, targetName string) (targets []nddv1.Target, err error) {
	return t.FindTargetWithStatus(ctx, targetName, ndrv1.ConditionKindDeviceDriverConfigured)
}

func (t *TargetValidator) FindReady(ctx context.Context, targetName string) (targets []nddv1.Target, err error) {
	return t.FindTargetWithStatus(ctx, targetName, ndrv1.ConditionKindDeviceDriverReady)
}

func (t *TargetValidator) FindTargetWithStatus(ctx context.Context, targetName string, c nddv1.ConditionKind) (targets []nddv1.Target, err error) {
	log := t.log.WithValues("targetName", targetName)
	log.Debug("Find target")

	// get list of network nodes
	nnl := &ndrv1.NetworkNodeList{}
	if err := t.client.List(ctx, nnl); err != nil {
		return nil, errors.Wrap(err, errFailedListNetworkNodes)
	}

	for _, nn := range nnl.Items {
		log.Debug("Find Target", "networkNode", nn)
		// check if the status of the network node is according to the status
		if nn.GetCondition(c).Status == corev1.ConditionTrue {
			// find the applicable target, for all resources there should be an exact match of the network node
			// and the target; there is one exception using
			log.Debug("Find Target", "targetName", targetName, "nnName", nn.Name)
			if targetName == string(nddv1.TargetValueAll) || targetName == nn.Name {
				target := nddv1.Target{
					Name: nn.Name,
					DNS:  ndrv1.PrefixService + "-" + nn.Name + "." + ndrv1.NamespaceLocalK8sDNS + strconv.Itoa(*nn.Spec.GrpcServerPort),
				}
				targets = append(targets, target)
			}
		}
	}
	log.Debug("Find Target", "withCondition", c, "targets", targets)
	return targets, nil
}

func (t *TargetValidator) IfDeleted(ctx context.Context, targets []nddv1.Target, target string) bool {
	deleted := true
	for _, t := range targets {
		if t.Name == target {
			deleted = false
		}
	}
	return deleted
}

*/
