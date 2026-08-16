package domain

import (
	"context"
	"time"
)

// Environment names supported for every SecretProject. Creating a
// SecretProject always provisions all three.
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// ValidEnvironments is the ordered set of environment names created for
// every new SecretProject.
var ValidEnvironments = []string{EnvDevelopment, EnvStaging, EnvProduction}

// Audit action labels recorded in SecretAuditLog.Action.
const (
	AuditActionCreate = "create"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"
	AuditActionReveal = "reveal"
	AuditActionExport = "export"
	AuditActionImport = "import"
)

// Product is the top level of the secret hierarchy:
// Product -> SecretProject -> SecretEnvironment -> Secret.
//
// This is intentionally a distinct table/type from the existing todo
// Project (see project.go) so the two hierarchies never collide.
type Product struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

// SecretProject groups secrets under a Product. It always has exactly one
// SecretEnvironment per name in ValidEnvironments.
type SecretProject struct {
	ID        int64
	ProductID int64
	Name      string
	CreatedAt time.Time
}

// SecretEnvironment is one of development/staging/production scoped to a
// SecretProject.
type SecretEnvironment struct {
	ID        int64
	ProjectID int64
	Name      string
	CreatedAt time.Time
}

// Secret stores an encrypted value under a key, scoped to a
// SecretEnvironment. ValueEncrypted holds ciphertext only — the plaintext
// value is never persisted or logged.
type Secret struct {
	ID             int64
	EnvironmentID  int64
	Key            string
	ValueEncrypted string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SecretAuditLog records a single audited action against a secret. Value is
// deliberately never included.
type SecretAuditLog struct {
	ID            int64
	EnvironmentID int64
	SecretKey     string
	Action        string
	Actor         string
	CreatedAt     time.Time
}

// SecretStore defines the persistence contract for the secret manager
// hierarchy. A single concrete store (e.g. SQLiteStore) implements this
// alongside TodoStore/ProjectStore without name clashes.
type SecretStore interface {
	CreateProduct(ctx context.Context, product *Product) error
	GetAllProducts(ctx context.Context) ([]*Product, error)
	GetProductByID(ctx context.Context, id int64) (*Product, error)

	// CreateSecretProjectWithEnvironments creates a SecretProject and its
	// three default environments atomically, returning the created
	// environments in ValidEnvironments order.
	CreateSecretProjectWithEnvironments(ctx context.Context, project *SecretProject) ([]*SecretEnvironment, error)
	GetSecretProjectsByProduct(ctx context.Context, productID int64) ([]*SecretProject, error)
	GetSecretProjectByID(ctx context.Context, id int64) (*SecretProject, error)

	GetEnvironmentsByProject(ctx context.Context, projectID int64) ([]*SecretEnvironment, error)
	GetEnvironmentByID(ctx context.Context, id int64) (*SecretEnvironment, error)

	CreateSecret(ctx context.Context, secret *Secret) error
	GetSecretsByEnvironment(ctx context.Context, environmentID int64) ([]*Secret, error)
	GetSecretByKey(ctx context.Context, environmentID int64, key string) (*Secret, error)
	UpdateSecret(ctx context.Context, secret *Secret) error
	UpsertSecrets(ctx context.Context, environmentID int64, secrets []*Secret) error
	DeleteSecret(ctx context.Context, environmentID int64, key string) error

	CreateAuditLog(ctx context.Context, log *SecretAuditLog) error
	GetAuditLogs(ctx context.Context, environmentID int64) ([]*SecretAuditLog, error)
}
