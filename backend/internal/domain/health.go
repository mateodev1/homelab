package domain

// HealthStatus represents the current health of the API and its dependencies.
type HealthStatus struct {
	Status string
	DBOk   bool
}
