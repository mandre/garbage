/*
Copyright 2024 The ORC Authors.

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

package port

import (
	"context"
	"errors"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	orcv1alpha1 "github.com/k-orc/openstack-resource-controller/v2/api/v1alpha1"
	"github.com/k-orc/openstack-resource-controller/v2/pkg/predicates"

	"github.com/k-orc/openstack-resource-controller/v2/internal/controllers/generic/interfaces"
	"github.com/k-orc/openstack-resource-controller/v2/internal/controllers/generic/reconciler"
	"github.com/k-orc/openstack-resource-controller/v2/internal/scope"
	"github.com/k-orc/openstack-resource-controller/v2/internal/util/credentials"
	"github.com/k-orc/openstack-resource-controller/v2/internal/util/dependency"
)

// +kubebuilder:rbac:groups=openstack.k-orc.cloud,resources=ports,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=openstack.k-orc.cloud,resources=ports/status,verbs=get;update;patch

const controllerName = "port"

var (
	networkDependency = dependency.NewDeletionGuardDependency[*orcv1alpha1.PortList, *orcv1alpha1.Network](
		"spec.resource.networkRef",
		func(port *orcv1alpha1.Port) []string {
			resource := port.Spec.Resource
			if resource == nil {
				return nil
			}
			return []string{string(resource.NetworkRef)}
		},
		finalizer, externalObjectFieldOwner,
	)

	networkImportDependency = dependency.NewDependency[*orcv1alpha1.PortList, *orcv1alpha1.Network](
		"spec.import.filter.networkRef",
		func(port *orcv1alpha1.Port) []string {
			resource := port.Spec.Import
			if resource == nil || resource.Filter == nil {
				return nil
			}
			return []string{string(resource.Filter.NetworkRef)}
		},
	)

	subnetDependency = dependency.NewDeletionGuardDependency[*orcv1alpha1.PortList, *orcv1alpha1.Subnet](
		"spec.resource.addresses[].subnetRef",
		func(port *orcv1alpha1.Port) []string {
			if port.Spec.Resource == nil {
				return nil
			}
			subnets := make([]string, len(port.Spec.Resource.Addresses))
			for i := range port.Spec.Resource.Addresses {
				subnets[i] = string(port.Spec.Resource.Addresses[i].SubnetRef)
			}
			return subnets
		},
		finalizer, externalObjectFieldOwner,
	)

	securityGroupDependency = dependency.NewDeletionGuardDependency[*orcv1alpha1.PortList, *orcv1alpha1.SecurityGroup](
		"spec.resource.securityGroupRefs",
		func(port *orcv1alpha1.Port) []string {
			if port.Spec.Resource == nil {
				return nil
			}
			securityGroups := make([]string, len(port.Spec.Resource.SecurityGroupRefs))
			for i := range port.Spec.Resource.SecurityGroupRefs {
				securityGroups[i] = string(port.Spec.Resource.SecurityGroupRefs[i])
			}
			return securityGroups
		},
		finalizer, externalObjectFieldOwner,
	)

	projectDependency = dependency.NewDeletionGuardDependency[*orcv1alpha1.PortList, *orcv1alpha1.Project](
		"spec.resource.projectRef",
		func(port *orcv1alpha1.Port) []string {
			resource := port.Spec.Resource
			if resource == nil || resource.ProjectRef == nil {
				return nil
			}
			return []string{string(*resource.ProjectRef)}
		},
		finalizer, externalObjectFieldOwner,
	)

	projectImportDependency = dependency.NewDependency[*orcv1alpha1.PortList, *orcv1alpha1.Project](
		"spec.import.filter.projectRef",
		func(port *orcv1alpha1.Port) []string {
			resource := port.Spec.Import
			if resource == nil || resource.Filter == nil || resource.Filter.ProjectRef == nil {
				return nil
			}
			return []string{string(*resource.Filter.ProjectRef)}
		},
	)
)

type portReconcilerConstructor struct {
	scopeFactory scope.Factory
}

func New(scopeFactory scope.Factory) interfaces.Controller {
	return portReconcilerConstructor{scopeFactory: scopeFactory}
}

func (portReconcilerConstructor) GetName() string {
	return controllerName
}

// SetupWithManager sets up the controller with the Manager.
func (c portReconcilerConstructor) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	log := mgr.GetLogger().WithValues("controller", controllerName)
	k8sClient := mgr.GetClient()

	networkWatchEventHandler, err := networkDependency.WatchEventHandler(log, k8sClient)
	if err != nil {
		return err
	}

	networkImportWatchEventHandler, err := networkImportDependency.WatchEventHandler(log, k8sClient)
	if err != nil {
		return err
	}

	subnetWatchEventHandler, err := subnetDependency.WatchEventHandler(log, k8sClient)
	if err != nil {
		return err
	}

	securityGroupWatchEventHandler, err := securityGroupDependency.WatchEventHandler(log, k8sClient)
	if err != nil {
		return err
	}

	projectWatchEventHandler, err := projectDependency.WatchEventHandler(log, k8sClient)
	if err != nil {
		return err
	}

	projectImportWatchEventHandler, err := projectImportDependency.WatchEventHandler(log, k8sClient)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&orcv1alpha1.Port{}).
		Watches(&orcv1alpha1.Network{}, networkWatchEventHandler,
			builder.WithPredicates(predicates.NewBecameAvailable(log, &orcv1alpha1.Network{})),
		).
		// A second watch is necessary because we need a different handler that omits deletion guards
		Watches(&orcv1alpha1.Network{}, networkImportWatchEventHandler,
			builder.WithPredicates(predicates.NewBecameAvailable(log, &orcv1alpha1.Network{})),
		).
		Watches(&orcv1alpha1.Subnet{}, subnetWatchEventHandler,
			builder.WithPredicates(predicates.NewBecameAvailable(log, &orcv1alpha1.Subnet{})),
		).
		Watches(&orcv1alpha1.SecurityGroup{}, securityGroupWatchEventHandler,
			builder.WithPredicates(predicates.NewBecameAvailable(log, &orcv1alpha1.SecurityGroup{})),
		).
		Watches(&orcv1alpha1.Project{}, projectWatchEventHandler,
			builder.WithPredicates(predicates.NewBecameAvailable(log, &orcv1alpha1.Project{})),
		).
		// A second watch is necessary because we need a different handler that omits deletion guards
		Watches(&orcv1alpha1.Project{}, projectImportWatchEventHandler,
			builder.WithPredicates(predicates.NewBecameAvailable(log, &orcv1alpha1.Project{})),
		)

	if err := errors.Join(
		networkDependency.AddToManager(ctx, mgr),
		networkImportDependency.AddToManager(ctx, mgr),
		subnetDependency.AddToManager(ctx, mgr),
		securityGroupDependency.AddToManager(ctx, mgr),
		projectDependency.AddToManager(ctx, mgr),
		projectImportDependency.AddToManager(ctx, mgr),
		credentialsDependency.AddToManager(ctx, mgr),
		credentials.AddCredentialsWatch(log, k8sClient, builder, credentialsDependency),
	); err != nil {
		return err
	}

	r := reconciler.NewController(controllerName, k8sClient, c.scopeFactory, portHelperFactory{}, portStatusWriter{})
	return builder.Complete(&r)
}
