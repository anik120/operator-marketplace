package catalogsourceconfig

import (
	"context"

	marketplace "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	"github.com/operator-framework/operator-marketplace/pkg/catalogsource"
	"github.com/operator-framework/operator-marketplace/pkg/phase"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewDeletedReconciler returns a Reconciler that reconciles
// a CatalogSourceConfig that has been marked for deletion.
func NewDeletedReconciler(logger *log.Entry, cache Cache, client client.Client) Reconciler {
	return &deletedReconciler{
		logger: logger,
		cache:  cache,
		client: client,
	}
}

// deletedReconciler is an implementation of Reconciler interface that
// reconciles a CatalogSourceConfig object that has been marked for deletion.
type deletedReconciler struct {
	logger *log.Entry
	cache  Cache
	client client.Client
}

// Reconcile reconciles a CatalogSourceConfig object that is marked for deletion.
// In the generic case, this is called when the CatalogSourceConfig has been marked
// for deletion. It removes all data related to this CatalogSourceConfig from the
// datastore, and it removes the CatalogSourceConfig finalizer from the object so
// that it can be cleaned up by the garbage collector.
//
// in represents the original CatalogSourceConfig object received from the sdk
// and before reconciliation has started.
//
// out represents the CatalogSourceConfig object after reconciliation has completed
// and could be different from the original. The CatalogSourceConfig object received
// (in) should be deep copied into (out) before changes are made.
//
// nextPhase represents the next desired phase for the given CatalogSourceConfig
// object. If nil is returned, it implies that no phase transition is expected.
func (r *deletedReconciler) Reconcile(ctx context.Context, in *marketplace.CatalogSourceConfig) (out *marketplace.CatalogSourceConfig, nextPhase *marketplace.Phase, err error) {
	out = in

	// Evict the catalogsourceconfig data from the cache.
	r.cache.Evict(out)

	// Delete all created resources
	deletionReconciler := catalogsource.NewDeletionReconciler(r.logger, r.client)
	err = deletionReconciler.DeleteCatalogSourceResources(ctx, in.Name, in.Namespace, in.Spec.TargetNamespace)

	if err != nil {
		// Something went wrong before we removed the finalizer, let's retry.
		nextPhase = phase.GetNextWithMessage(in.Status.CurrentPhase.Name, err.Error())
		return
	}

	// Remove the csc finalizer from the object.
	out.RemoveFinalizer()

	// Update the client. Since there is no phase shift, the transitioner
	// will not update it automatically like the normal phases.
	err = r.client.Update(context.TODO(), out)
	if err != nil {
		// An error happened on update. If it was transient, we will retry.
		// If not, and the finalizer was removed, then the delete will clean
		// the object up anyway. Let's set the next phase for a possible retry.
		nextPhase = phase.GetNextWithMessage(in.Status.CurrentPhase.Name, err.Error())
		return
	}

	r.logger.Info("Finalizer removed, now garbage collector will clean it up.")

	return
}
