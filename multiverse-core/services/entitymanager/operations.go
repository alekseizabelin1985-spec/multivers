// services/entitymanager/manager.go
package entitymanager

type OperationType string

const (
	OpSet             OperationType = "set"
	OpAddToSlice      OperationType = "add_to_slice"
	OpRemoveFromSlice OperationType = "remove_from_slice"
	OpRemove          OperationType = "remove"
)

type Operation struct {
	Op    OperationType `json:"op"`
	Path  string        `json:"path"`
	Value interface{}   `json:"value,omitempty"`
}

type StateChange struct {
	EntityID   string      `json:"entity_id"`
	Operations []Operation `json:"operations"`
}
