package dto

// Product mirrors the backend wire shape for a product (the top-level
// container for secret projects).
type Product struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// CreateProduct is the request body for POST /api/products.
type CreateProduct struct {
	Name string `json:"name"`
}

// SecretProject mirrors the backend wire shape for a secret project nested
// under a product.
type SecretProject struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// CreateSecretProject is the request body for
// POST /api/products/{productID}/projects.
type CreateSecretProject struct {
	Name string `json:"name"`
}

// CreateSecretProjectResponse is the response body for
// POST /api/products/{productID}/projects. The backend auto-provisions the
// standard environments (development/staging/production) alongside the
// project and returns them so callers don't need a follow-up list call.
type CreateSecretProjectResponse struct {
	ID           int64         `json:"id"`
	ProductID    int64         `json:"product_id"`
	Name         string        `json:"name"`
	CreatedAt    string        `json:"created_at"`
	Environments []Environment `json:"environments"`
}

// Environment mirrors the backend wire shape for an environment nested under
// a secret project. Name is one of development/staging/production.
type Environment struct {
	ID        int64  `json:"id"`
	ProjectID int64  `json:"project_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// SecretMetadata mirrors the backend wire shape for a secret's metadata.
// Value is masked by the backend on list/get responses; only the reveal
// endpoint returns the plaintext value.
type SecretMetadata struct {
	EnvironmentID int64  `json:"environment_id"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// CreateSecret is the request body for
// POST /api/products/{productID}/projects/{projectID}/environments/{envName}/secrets.
type CreateSecret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// UpdateSecret is the request body for
// PUT .../secrets/{key}.
type UpdateSecret struct {
	Value string `json:"value"`
}

// SecretRevealResponse is the response body for GET .../secrets/{key}/reveal.
// It intentionally carries the plaintext value — reveal is an explicit,
// deliberate action by the caller.
type SecretRevealResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
