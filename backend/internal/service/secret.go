package service

import (
	"context"
	"crypto/cipher"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/store"
)

// maskedValuePlaceholder is returned instead of plaintext for list/get
// responses. It never leaks the real value's length.
const maskedValuePlaceholder = "••••••••"

// SecretMeta is the metadata-only view of a Secret returned by list/get/create/update.
// It never carries the plaintext value.
type SecretMeta struct {
	EnvironmentID int64
	Key           string
	MaskedValue   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SecretService implements all business logic for the secret manager
// hierarchy (Product -> SecretProject -> SecretEnvironment -> Secret). It
// depends exclusively on domain.SecretStore for persistence and a
// cipher.AEAD for encryption at rest — no direct DB or crypto-library
// access outside these seams.
type SecretService struct {
	store domain.SecretStore
	aead  cipher.AEAD
}

// NewSecretService creates a new SecretService. aead is the test-safe
// injection seam: production wires it via store.ResolveEncryptionKey +
// store.NewGCMCipher in cmd/api/main.go, tests build it directly with
// store.NewGCMCipher(fixedKey).
func NewSecretService(secretStore domain.SecretStore, aead cipher.AEAD) *SecretService {
	return &SecretService{store: secretStore, aead: aead}
}

// CreateProduct creates a new Product.
func (s *SecretService) CreateProduct(ctx context.Context, name string, createdAt time.Time) (*domain.Product, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required: %w", ErrValidation)
	}
	product := &domain.Product{Name: name, CreatedAt: createdAt.UTC()}
	if err := s.store.CreateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("service.CreateProduct: %w", err)
	}
	return product, nil
}

// ListProducts returns all Products.
func (s *SecretService) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	products, err := s.store.GetAllProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.ListProducts: %w", err)
	}
	return products, nil
}

// CreateSecretProject creates a new SecretProject under a Product and
// provisions all three default environments (development/staging/production).
func (s *SecretService) CreateSecretProject(ctx context.Context, productID int64, name string, createdAt time.Time) (*domain.SecretProject, []*domain.SecretEnvironment, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil, fmt.Errorf("name is required: %w", ErrValidation)
	}
	if _, err := s.store.GetProductByID(ctx, productID); err != nil {
		return nil, nil, fmt.Errorf("service.CreateSecretProject: %w", err)
	}

	project := &domain.SecretProject{ProductID: productID, Name: name, CreatedAt: createdAt.UTC()}
	envs, err := s.store.CreateSecretProjectWithEnvironments(ctx, project)
	if err != nil {
		return nil, nil, fmt.Errorf("service.CreateSecretProject: %w", err)
	}
	return project, envs, nil
}

// ListSecretProjects returns all SecretProjects for a Product.
func (s *SecretService) ListSecretProjects(ctx context.Context, productID int64) ([]*domain.SecretProject, error) {
	if _, err := s.store.GetProductByID(ctx, productID); err != nil {
		return nil, fmt.Errorf("service.ListSecretProjects: %w", err)
	}
	projects, err := s.store.GetSecretProjectsByProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("service.ListSecretProjects: %w", err)
	}
	return projects, nil
}

// ListEnvironments returns all SecretEnvironments for a SecretProject.
func (s *SecretService) ListEnvironments(ctx context.Context, projectID int64) ([]*domain.SecretEnvironment, error) {
	if _, err := s.store.GetSecretProjectByID(ctx, projectID); err != nil {
		return nil, fmt.Errorf("service.ListEnvironments: %w", err)
	}
	envs, err := s.store.GetEnvironmentsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("service.ListEnvironments: %w", err)
	}
	return envs, nil
}

// resolveEnvironment finds the SecretEnvironment with the given name under a project.
func (s *SecretService) resolveEnvironment(ctx context.Context, projectID int64, envName string) (*domain.SecretEnvironment, error) {
	envs, err := s.store.GetEnvironmentsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("service.resolveEnvironment: %w", err)
	}
	for _, e := range envs {
		if e.Name == envName {
			return e, nil
		}
	}
	return nil, fmt.Errorf("environment %q: %w", envName, domain.ErrNotFound)
}

// CreateSecret encrypts value and stores a new Secret under environmentID.
// Returns metadata only — the plaintext is never returned.
func (s *SecretService) CreateSecret(ctx context.Context, projectID int64, envName, key, value, actor string) (SecretMeta, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return SecretMeta{}, fmt.Errorf("key is required: %w", ErrValidation)
	}
	env, err := s.resolveEnvironment(ctx, projectID, envName)
	if err != nil {
		return SecretMeta{}, err
	}

	encrypted, err := store.EncryptSecretValue(s.aead, value)
	if err != nil {
		return SecretMeta{}, fmt.Errorf("service.CreateSecret encrypt: %w", err)
	}

	now := time.Now().UTC()
	secret := &domain.Secret{EnvironmentID: env.ID, Key: key, ValueEncrypted: encrypted, CreatedAt: now}
	if err := s.store.CreateSecret(ctx, secret); err != nil {
		return SecretMeta{}, fmt.Errorf("service.CreateSecret: %w", err)
	}

	s.audit(ctx, env.ID, key, domain.AuditActionCreate, actor)
	return toSecretMeta(secret), nil
}

// ListSecrets returns masked metadata for every Secret in an environment.
// Values are never decrypted for a list operation.
func (s *SecretService) ListSecrets(ctx context.Context, projectID int64, envName string) ([]SecretMeta, error) {
	env, err := s.resolveEnvironment(ctx, projectID, envName)
	if err != nil {
		return nil, err
	}
	secrets, err := s.store.GetSecretsByEnvironment(ctx, env.ID)
	if err != nil {
		return nil, fmt.Errorf("service.ListSecrets: %w", err)
	}
	out := make([]SecretMeta, len(secrets))
	for i, sec := range secrets {
		out[i] = toSecretMeta(sec)
	}
	return out, nil
}

// GetSecret returns masked metadata for a single Secret.
func (s *SecretService) GetSecret(ctx context.Context, projectID int64, envName, key string) (SecretMeta, error) {
	env, err := s.resolveEnvironment(ctx, projectID, envName)
	if err != nil {
		return SecretMeta{}, err
	}
	secret, err := s.store.GetSecretByKey(ctx, env.ID, key)
	if err != nil {
		return SecretMeta{}, fmt.Errorf("service.GetSecret: %w", err)
	}
	return toSecretMeta(secret), nil
}

// RevealSecret decrypts and returns the plaintext value. Restricted to
// authorized callers by the handler layer (auth middleware). Audited.
func (s *SecretService) RevealSecret(ctx context.Context, projectID int64, envName, key, actor string) (string, error) {
	env, err := s.resolveEnvironment(ctx, projectID, envName)
	if err != nil {
		return "", err
	}
	secret, err := s.store.GetSecretByKey(ctx, env.ID, key)
	if err != nil {
		return "", fmt.Errorf("service.RevealSecret: %w", err)
	}
	plaintext, err := store.DecryptSecretValue(s.aead, secret.ValueEncrypted)
	if err != nil {
		return "", fmt.Errorf("service.RevealSecret decrypt: %w", err)
	}
	s.audit(ctx, env.ID, key, domain.AuditActionReveal, actor)
	return plaintext, nil
}

// UpdateSecret re-encrypts and persists a new value for an existing Secret.
func (s *SecretService) UpdateSecret(ctx context.Context, projectID int64, envName, key, value, actor string) (SecretMeta, error) {
	env, err := s.resolveEnvironment(ctx, projectID, envName)
	if err != nil {
		return SecretMeta{}, err
	}
	if _, err := s.store.GetSecretByKey(ctx, env.ID, key); err != nil {
		return SecretMeta{}, fmt.Errorf("service.UpdateSecret: %w", err)
	}

	encrypted, err := store.EncryptSecretValue(s.aead, value)
	if err != nil {
		return SecretMeta{}, fmt.Errorf("service.UpdateSecret encrypt: %w", err)
	}

	secret := &domain.Secret{EnvironmentID: env.ID, Key: key, ValueEncrypted: encrypted}
	if err := s.store.UpdateSecret(ctx, secret); err != nil {
		return SecretMeta{}, fmt.Errorf("service.UpdateSecret: %w", err)
	}

	updated, err := s.store.GetSecretByKey(ctx, env.ID, key)
	if err != nil {
		return SecretMeta{}, fmt.Errorf("service.UpdateSecret reload: %w", err)
	}

	s.audit(ctx, env.ID, key, domain.AuditActionUpdate, actor)
	return toSecretMeta(updated), nil
}

// DeleteSecret removes a Secret from an environment.
func (s *SecretService) DeleteSecret(ctx context.Context, projectID int64, envName, key, actor string) error {
	env, err := s.resolveEnvironment(ctx, projectID, envName)
	if err != nil {
		return err
	}
	if err := s.store.DeleteSecret(ctx, env.ID, key); err != nil {
		return fmt.Errorf("service.DeleteSecret: %w", err)
	}
	s.audit(ctx, env.ID, key, domain.AuditActionDelete, actor)
	return nil
}

// ExportEnvironment returns all secrets in the environment as a dotenv-formatted
// document, keys sorted deterministically and values escaped/quoted.
func (s *SecretService) ExportEnvironment(ctx context.Context, projectID int64, envName, actor string) (string, error) {
	env, err := s.resolveEnvironment(ctx, projectID, envName)
	if err != nil {
		return "", err
	}
	secrets, err := s.store.GetSecretsByEnvironment(ctx, env.ID)
	if err != nil {
		return "", fmt.Errorf("service.ExportEnvironment: %w", err)
	}

	sort.Slice(secrets, func(i, j int) bool { return secrets[i].Key < secrets[j].Key })

	var b strings.Builder
	for _, sec := range secrets {
		plaintext, err := store.DecryptSecretValue(s.aead, sec.ValueEncrypted)
		if err != nil {
			return "", fmt.Errorf("service.ExportEnvironment decrypt %q: %w", sec.Key, err)
		}
		b.WriteString(sec.Key)
		b.WriteByte('=')
		b.WriteString(dotenvEscape(plaintext))
		b.WriteByte('\n')
	}

	s.audit(ctx, env.ID, "", domain.AuditActionExport, actor)
	return b.String(), nil
}

// ImportEnvironment parses dotenv content and atomically upserts every key in
// the selected environment. Existing keys not present in the document remain.
func (s *SecretService) ImportEnvironment(ctx context.Context, projectID int64, envName, content, actor string) (int, error) {
	env, err := s.resolveEnvironment(ctx, projectID, envName)
	if err != nil {
		return 0, err
	}

	values, err := ParseDotenv(content)
	if err != nil {
		return 0, err
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	secrets := make([]*domain.Secret, 0, len(keys))
	for _, key := range keys {
		encrypted, err := store.EncryptSecretValue(s.aead, values[key])
		if err != nil {
			return 0, fmt.Errorf("service.ImportEnvironment encrypt %q: %w", key, err)
		}
		secrets = append(secrets, &domain.Secret{
			EnvironmentID:  env.ID,
			Key:            key,
			ValueEncrypted: encrypted,
		})
	}

	if err := s.store.UpsertSecrets(ctx, env.ID, secrets); err != nil {
		return 0, fmt.Errorf("service.ImportEnvironment: %w", err)
	}
	for _, key := range keys {
		s.audit(ctx, env.ID, key, domain.AuditActionImport, actor)
	}
	return len(keys), nil
}

// ParseDotenv parses common dotenv forms. Blank lines and comments are
// ignored, export prefixes are accepted, and the last duplicate key wins.
func ParseDotenv(content string) (map[string]string, error) {
	values := make(map[string]string)
	for lineNumber, rawLine := range strings.Split(strings.TrimPrefix(content, "\ufeff"), "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(rawLine, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		separator := strings.IndexByte(line, '=')
		if separator <= 0 {
			return nil, fmt.Errorf("invalid dotenv line %d: expected KEY=value", lineNumber+1)
		}
		key := strings.TrimSpace(line[:separator])
		if key == "" {
			return nil, fmt.Errorf("invalid dotenv line %d: key is empty", lineNumber+1)
		}
		values[key] = parseDotenvValue(strings.TrimSpace(line[separator+1:]))
	}
	return values, nil
}

func parseDotenvValue(value string) string {
	if len(value) < 2 {
		return value
	}
	if value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1]
	}
	if value[0] != '"' || value[len(value)-1] != '"' {
		return value
	}

	value = value[1 : len(value)-1]
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] != '\\' || i+1 >= len(value) {
			b.WriteByte(value[i])
			continue
		}
		i++
		switch value[i] {
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		default:
			b.WriteByte(value[i])
		}
	}
	return b.String()
}

// dotenvEscape quotes and escapes a value for safe inclusion in a dotenv
// file, handling spaces, quotes, backslashes, newlines, '#' and empty values.
func dotenvEscape(value string) string {
	needsQuoting := value == "" ||
		strings.ContainsAny(value, " \t\"'\\\n\r#")
	if !needsQuoting {
		return value
	}

	var b strings.Builder
	b.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// audit best-effort records an audit log entry. Failures are swallowed
// (logged via the returned error being dropped) rather than failing the
// primary operation — audit is a side effect, not the source of truth.
// It never includes the secret value.
func (s *SecretService) audit(ctx context.Context, environmentID int64, key, action, actor string) {
	if actor == "" {
		actor = "unknown"
	}
	_ = s.store.CreateAuditLog(ctx, &domain.SecretAuditLog{
		EnvironmentID: environmentID,
		SecretKey:     key,
		Action:        action,
		Actor:         actor,
		CreatedAt:     time.Now().UTC(),
	})
}

func toSecretMeta(sec *domain.Secret) SecretMeta {
	return SecretMeta{
		EnvironmentID: sec.EnvironmentID,
		Key:           sec.Key,
		MaskedValue:   maskedValuePlaceholder,
		CreatedAt:     sec.CreatedAt,
		UpdatedAt:     sec.UpdatedAt,
	}
}

// ErrSecretNotFound is a convenience alias check for handlers.
var ErrSecretNotFound = errors.New("secret not found")
