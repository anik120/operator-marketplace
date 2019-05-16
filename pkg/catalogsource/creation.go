package catalogsource

import (
	"context"
	"strings"

	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	"github.com/operator-framework/operator-marketplace/pkg/builders"
	wrapper "github.com/operator-framework/operator-marketplace/pkg/client"
	"github.com/operator-framework/operator-marketplace/pkg/datastore"
	"github.com/operator-framework/operator-marketplace/pkg/registry"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// NewResourcesReconciler returns a new ResourcesReconciler.
func NewResourcesReconciler(log *logrus.Entry, reader datastore.Reader, client wrapper.Client) ResourcesReconciler {
	return ResourcesReconciler{
		log:    log,
		reader: reader,
		client: client,
	}
}

// ResourcesReconciler creates the resources required for a grpc CatalogSource
type ResourcesReconciler struct {
	log    *logrus.Entry
	reader datastore.Reader
	client wrapper.Client
}

// ReconcileCatalogSourceResources creates a CatalogSource if one does not already
// exists otherwise it updates an existing one. It then creates/updates all the
// resources it requires.
func (r *ResourcesReconciler) ReconcileCatalogSourceResources(key types.NamespacedName, displayName, publisher, targetNamespace, packages string, cscLabels map[string]string) error {
	// Ensure that the packages in the spec are available in the datastore
	err := r.reader.CheckPackages(v1.GetValidPackageSliceFromString(packages))
	if err != nil {
		return err
	}

	// Ensure that a registry deployment is available
	registry := registry.NewRegistry(r.log, r.client, r.reader, key, packages, registry.ServerImage)
	err = registry.Ensure()
	if err != nil {
		return err
	}

	// Check if the CatalogSource already exists
	catalogSourceGet := new(builders.CatalogSourceBuilder).WithTypeMeta().CatalogSource()
	err = r.client.Get(context.TODO(), key, catalogSourceGet)

	// Update the CatalogSource if it exists else create one.
	if err == nil {
		catalogSourceGet.Spec.Address = registry.GetAddress()
		r.log.Infof("Updating CatalogSource %s", catalogSourceGet.Name)
		err = r.client.Update(context.TODO(), catalogSourceGet)
		if err != nil {
			r.log.Errorf("Failed to update CatalogSource : %v", err)
			return err
		}
		r.log.Infof("Updated CatalogSource %s", catalogSourceGet.Name)
	} else {
		// Create the CatalogSource structure
		catalogSource := newCatalogSource(cscLabels, key, displayName, publisher, targetNamespace, registry.GetAddress())
		r.log.Infof("Creating CatalogSource %s", catalogSource.Name)
		err = r.client.Create(context.TODO(), catalogSource)
		if err != nil && !errors.IsAlreadyExists(err) {
			r.log.Errorf("Failed to create CatalogSource : %v", err)
			return err
		}
		r.log.Infof("Created CatalogSource %s", catalogSource.Name)
	}

	return nil
}

// newCatalogSource returns a CatalogSource object.
func newCatalogSource(cscLabels map[string]string, key types.NamespacedName, displayName, publisher, namespace, address string) *olm.CatalogSource {
	builder := new(builders.CatalogSourceBuilder).
		WithOwnerLabel(key.Name, namespace).
		WithMeta(key.Name, key.Namespace).
		WithSpec(olm.SourceTypeGrpc, address, displayName, publisher)

	// Check if the operatorsource.DatastoreLabel is "true" which indicates that
	// the CatalogSource is the datastore for an OperatorSource. This is a hint
	// for us to set the "olm-visibility" label in the CatalogSource so that it
	// is not visible in the OLM Packages UI. In addition we will set the
	// "openshift-marketplace" label which will be used by the Marketplace UI
	// to filter out global CatalogSources.
	datastoreLabel, found := cscLabels[datastore.DatastoreLabel]
	if found && strings.ToLower(datastoreLabel) == "true" {
		builder.WithOLMLabels(cscLabels)
	}

	return builder.CatalogSource()
}
