package catalogsource

import (
	"context"

	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-marketplace/pkg/builders"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

// NewDeletionReconciler returns a new DeletionReconciler.
func NewDeletionReconciler(log *logrus.Entry, client client.Client) DeletionReconciler {
	return DeletionReconciler{
		log:    log,
		client: client,
	}
}

// DeletionReconciler encapsulates the logic used to remove all resources related to a registry.
type DeletionReconciler struct {
	log    *logrus.Entry
	client client.Client
}

// DeleteCatalogSourceResources deletes a CatalogSource and all resources that make up a registry.
func (r *DeletionReconciler) DeleteCatalogSourceResources(ctx context.Context, name, namespace, targetNamespace string) (err error) {
	allErrors := []error{}
	labelMap := map[string]string{
		builders.OwnerNameLabel:      name,
		builders.OwnerNamespaceLabel: namespace,
	}
	labelSelector := labels.SelectorFromSet(labelMap)
	catalogSourceOptions := &client.ListOptions{LabelSelector: labelSelector}
	catalogSourceOptions.InNamespace(targetNamespace)
	namespacedResourceOptions := &client.ListOptions{LabelSelector: labelSelector}
	namespacedResourceOptions.InNamespace(namespace)

	// Delete Catalog Sources
	catalogSources := &olm.CatalogSourceList{}
	err = r.client.List(ctx, catalogSourceOptions, catalogSources)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	for _, catalogSource := range catalogSources.Items {
		r.log.Infof("Removing catalogSource %s from namespace %s", catalogSource.Name, catalogSource.Namespace)
		err := r.client.Delete(ctx, &catalogSource)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// Delete Services
	services := &core.ServiceList{}
	err = r.client.List(ctx, namespacedResourceOptions, services)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	for _, service := range services.Items {
		r.log.Infof("Removing service %s from namespace %s", service.Name, service.Namespace)
		err := r.client.Delete(ctx, &service)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// Delete Deployments
	deployments := &apps.DeploymentList{}
	err = r.client.List(ctx, namespacedResourceOptions, deployments)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	for _, deployment := range deployments.Items {
		r.log.Infof("Removing deployment %s from namespace %s", deployment.Name, deployment.Namespace)
		err := r.client.Delete(ctx, &deployment)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// Delete Role Bindings
	roleBindings := &rbac.RoleBindingList{}
	err = r.client.List(ctx, namespacedResourceOptions, roleBindings)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	for _, roleBinding := range roleBindings.Items {
		r.log.Infof("Removing roleBinding %s from namespace %s", roleBinding.Name, roleBinding.Namespace)
		err := r.client.Delete(ctx, &roleBinding)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// Delete Roles
	roles := &rbac.RoleList{}
	err = r.client.List(ctx, namespacedResourceOptions, roles)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	for _, role := range roles.Items {
		r.log.Infof("Removing role %s from namespace %s", role.Name, role.Namespace)
		err := r.client.Delete(ctx, &role)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// Delete Service Accounts
	serviceAccounts := &core.ServiceAccountList{}
	err = r.client.List(ctx, namespacedResourceOptions, serviceAccounts)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	for _, serviceAccount := range serviceAccounts.Items {
		r.log.Infof("Removing serviceAccount %s from namespace %s", serviceAccount.Name, serviceAccount.Namespace)
		err := r.client.Delete(ctx, &serviceAccount)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	return utilerrors.NewAggregate(allErrors)
}
