package domain

// CodeAgentResult is a plain value object describing an external result, not an entity —
// same rationale as Sandbox: a port in application/services referencing an infrastructure
// package would point the dependency arrow outward.
type CodeAgentResult struct {
	Summary  string            `json:"summary"`
	Commands []string          `json:"commands"`
	Files    map[string]string `json:"files"`
}
