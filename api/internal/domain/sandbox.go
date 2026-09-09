package domain

// Sandbox is a plain value object describing an external result, not an entity. Public
// fields with json tags on purpose: application ports reference this type, and it may cross a
// Temporal activity boundary later (Temporal's JSON converter would serialise unexported fields
// to {}).
type Sandbox struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
