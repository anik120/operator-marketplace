package operatorsource

// DatastoreLabel is the label used in a CatalogSourceConfig to indicate that
// the resulting CatalogSource acts as the datastore for the OperatorSource
// if it is set to "true".
const DatastoreLabel string = "opsrc-datastore"

// OpsrcOwnerNameLabel is the label used to mark ownership over resources
// that are owned by the OperatorSource. When this label is set, the reconciler
// should handle these resources when the OperatorSource is deleted.
const OpsrcOwnerNameLabel string = "opsrc-owner-name"

// OpsrcOwnerNamespaceLabel is the label used to mark ownership over resources
// that are owned by the OperatorSource. When this label is set, the reconciler
// should handle these resources when the OperatorSource is deleted.
const OpsrcOwnerNamespaceLabel string = "opsrc-owner-namespace"
