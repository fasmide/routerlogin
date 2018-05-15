package conntrack

// FlowUpdate wraps a flow with meta data about it being an update, delete or a new flow
type FlowUpdate struct {
  // could be NEW, UPDATE and DELETE
  Type string
  // the flow
  Flow Flow
}
