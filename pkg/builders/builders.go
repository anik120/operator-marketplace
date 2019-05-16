package builders

// OwnerNameLabel is the label used to mark ownership over a given resources.
// When this label is set, the reconciler should handle these resources when the owner
// is deleted.
const OwnerNameLabel string = "csc-owner-name"

// OwnerNamespaceLabel is the label used to mark ownership over a given resources.
// When this label is set, the reconciler should handle these resources when the owner
// is deleted.
const OwnerNamespaceLabel string = "csc-owner-namespace"
