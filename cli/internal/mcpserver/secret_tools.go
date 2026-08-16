package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/mateo/homelab/cli/internal/dto"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type secretProductListInput struct{}

type secretProductListOutput struct {
	Products []dto.Product `json:"products" jsonschema:"list of products"`
}

type secretProductCreateInput struct {
	Name string `json:"name" jsonschema:"required product name"`
}

type secretProjectListInput struct {
	ProductID int64 `json:"product_id" jsonschema:"product id"`
}

type secretProjectListOutput struct {
	Projects []dto.SecretProject `json:"projects" jsonschema:"list of secret projects"`
}

type secretProjectCreateInput struct {
	ProductID int64  `json:"product_id" jsonschema:"product id"`
	Name      string `json:"name" jsonschema:"required secret project name"`
}

type secretEnvironmentListInput struct {
	ProductID int64 `json:"product_id" jsonschema:"product id"`
	ProjectID int64 `json:"project_id" jsonschema:"secret project id"`
}

type secretEnvironmentListOutput struct {
	Environments []dto.Environment `json:"environments" jsonschema:"list of environments"`
}

type secretListInput struct {
	ProductID   int64  `json:"product_id" jsonschema:"product id"`
	ProjectID   int64  `json:"project_id" jsonschema:"secret project id"`
	Environment string `json:"environment" jsonschema:"environment name: development, staging, or production"`
}

type secretListOutput struct {
	Secrets []dto.SecretMetadata `json:"secrets" jsonschema:"list of secret metadata (values masked)"`
}

type secretCreateInput struct {
	ProductID   int64  `json:"product_id" jsonschema:"product id"`
	ProjectID   int64  `json:"project_id" jsonschema:"secret project id"`
	Environment string `json:"environment" jsonschema:"environment name: development, staging, or production"`
	Key         string `json:"key" jsonschema:"required secret key"`
	Value       string `json:"value,omitempty" jsonschema:"secret value"`
}

type secretKeyInput struct {
	ProductID   int64  `json:"product_id" jsonschema:"product id"`
	ProjectID   int64  `json:"project_id" jsonschema:"secret project id"`
	Environment string `json:"environment" jsonschema:"environment name: development, staging, or production"`
	Key         string `json:"key" jsonschema:"required secret key"`
}

type secretUpdateInput struct {
	ProductID   int64  `json:"product_id" jsonschema:"product id"`
	ProjectID   int64  `json:"project_id" jsonschema:"secret project id"`
	Environment string `json:"environment" jsonschema:"environment name: development, staging, or production"`
	Key         string `json:"key" jsonschema:"required secret key"`
	Value       string `json:"value" jsonschema:"required new secret value"`
}

type secretExportInput struct {
	ProductID   int64  `json:"product_id" jsonschema:"product id"`
	ProjectID   int64  `json:"project_id" jsonschema:"secret project id"`
	Environment string `json:"environment" jsonschema:"environment name: development, staging, or production"`
}

type secretExportOutput struct {
	Dotenv string `json:"dotenv" jsonschema:"plaintext dotenv export of every secret in the environment, sorted by key"`
}

func (s *Server) secretProductList(ctx context.Context, _ *mcp.CallToolRequest, _ secretProductListInput) (*mcp.CallToolResult, secretProductListOutput, error) {
	raw, _, err := s.c.Do(ctx, "GET", productsPath(), nil)
	if err != nil {
		return nil, secretProductListOutput{}, err
	}
	products, err := decodeJSON[[]dto.Product](raw)
	if err != nil {
		return nil, secretProductListOutput{}, err
	}
	if products == nil {
		products = []dto.Product{}
	}
	return nil, secretProductListOutput{Products: products}, nil
}

func (s *Server) secretProductCreate(ctx context.Context, _ *mcp.CallToolRequest, in secretProductCreateInput) (*mcp.CallToolResult, dto.Product, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, dto.Product{}, fmt.Errorf("name is required")
	}
	body, err := encodeJSON(dto.CreateProduct{Name: in.Name})
	if err != nil {
		return nil, dto.Product{}, err
	}
	raw, _, err := s.c.Do(ctx, "POST", productsPath(), body)
	if err != nil {
		return nil, dto.Product{}, err
	}
	p, err := decodeJSON[dto.Product](raw)
	if err != nil {
		return nil, dto.Product{}, err
	}
	return nil, p, nil
}

func (s *Server) secretProjectList(ctx context.Context, _ *mcp.CallToolRequest, in secretProjectListInput) (*mcp.CallToolResult, secretProjectListOutput, error) {
	if in.ProductID <= 0 {
		return nil, secretProjectListOutput{}, errPositiveID("product_id")
	}
	raw, _, err := s.c.Do(ctx, "GET", secretProjectsPath(in.ProductID), nil)
	if err != nil {
		return nil, secretProjectListOutput{}, err
	}
	projects, err := decodeJSON[[]dto.SecretProject](raw)
	if err != nil {
		return nil, secretProjectListOutput{}, err
	}
	if projects == nil {
		projects = []dto.SecretProject{}
	}
	return nil, secretProjectListOutput{Projects: projects}, nil
}

func (s *Server) secretProjectCreate(ctx context.Context, _ *mcp.CallToolRequest, in secretProjectCreateInput) (*mcp.CallToolResult, dto.CreateSecretProjectResponse, error) {
	if in.ProductID <= 0 {
		return nil, dto.CreateSecretProjectResponse{}, errPositiveID("product_id")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, dto.CreateSecretProjectResponse{}, fmt.Errorf("name is required")
	}
	body, err := encodeJSON(dto.CreateSecretProject{Name: in.Name})
	if err != nil {
		return nil, dto.CreateSecretProjectResponse{}, err
	}
	raw, _, err := s.c.Do(ctx, "POST", secretProjectsPath(in.ProductID), body)
	if err != nil {
		return nil, dto.CreateSecretProjectResponse{}, err
	}
	p, err := decodeJSON[dto.CreateSecretProjectResponse](raw)
	if err != nil {
		return nil, dto.CreateSecretProjectResponse{}, err
	}
	return nil, p, nil
}

func (s *Server) secretEnvironmentList(ctx context.Context, _ *mcp.CallToolRequest, in secretEnvironmentListInput) (*mcp.CallToolResult, secretEnvironmentListOutput, error) {
	if in.ProductID <= 0 {
		return nil, secretEnvironmentListOutput{}, errPositiveID("product_id")
	}
	if in.ProjectID <= 0 {
		return nil, secretEnvironmentListOutput{}, errPositiveID("project_id")
	}
	raw, _, err := s.c.Do(ctx, "GET", secretEnvironmentsPath(in.ProductID, in.ProjectID), nil)
	if err != nil {
		return nil, secretEnvironmentListOutput{}, err
	}
	envs, err := decodeJSON[[]dto.Environment](raw)
	if err != nil {
		return nil, secretEnvironmentListOutput{}, err
	}
	if envs == nil {
		envs = []dto.Environment{}
	}
	return nil, secretEnvironmentListOutput{Environments: envs}, nil
}

func (s *Server) secretList(ctx context.Context, _ *mcp.CallToolRequest, in secretListInput) (*mcp.CallToolResult, secretListOutput, error) {
	if err := validateSecretScope(in.ProductID, in.ProjectID, in.Environment); err != nil {
		return nil, secretListOutput{}, err
	}
	raw, _, err := s.c.Do(ctx, "GET", secretsPath(in.ProductID, in.ProjectID, in.Environment), nil)
	if err != nil {
		return nil, secretListOutput{}, err
	}
	secrets, err := decodeJSON[[]dto.SecretMetadata](raw)
	if err != nil {
		return nil, secretListOutput{}, err
	}
	if secrets == nil {
		secrets = []dto.SecretMetadata{}
	}
	return nil, secretListOutput{Secrets: secrets}, nil
}

func (s *Server) secretCreate(ctx context.Context, _ *mcp.CallToolRequest, in secretCreateInput) (*mcp.CallToolResult, dto.SecretMetadata, error) {
	if err := validateSecretScope(in.ProductID, in.ProjectID, in.Environment); err != nil {
		return nil, dto.SecretMetadata{}, err
	}
	if strings.TrimSpace(in.Key) == "" {
		return nil, dto.SecretMetadata{}, fmt.Errorf("key is required")
	}
	body, err := encodeJSON(dto.CreateSecret{Key: in.Key, Value: in.Value})
	if err != nil {
		return nil, dto.SecretMetadata{}, err
	}
	raw, _, err := s.c.Do(ctx, "POST", secretsPath(in.ProductID, in.ProjectID, in.Environment), body)
	if err != nil {
		return nil, dto.SecretMetadata{}, err
	}
	meta, err := decodeJSON[dto.SecretMetadata](raw)
	if err != nil {
		return nil, dto.SecretMetadata{}, err
	}
	return nil, meta, nil
}

// secretReveal intentionally returns the plaintext value: reveal is an
// explicit, deliberate MCP tool action, not incidental exposure.
func (s *Server) secretReveal(ctx context.Context, _ *mcp.CallToolRequest, in secretKeyInput) (*mcp.CallToolResult, dto.SecretRevealResponse, error) {
	if err := validateSecretScope(in.ProductID, in.ProjectID, in.Environment); err != nil {
		return nil, dto.SecretRevealResponse{}, err
	}
	if strings.TrimSpace(in.Key) == "" {
		return nil, dto.SecretRevealResponse{}, fmt.Errorf("key is required")
	}
	raw, _, err := s.c.Do(ctx, "GET", secretRevealPath(in.ProductID, in.ProjectID, in.Environment, in.Key), nil)
	if err != nil {
		return nil, dto.SecretRevealResponse{}, err
	}
	resp, err := decodeJSON[dto.SecretRevealResponse](raw)
	if err != nil {
		return nil, dto.SecretRevealResponse{}, err
	}
	return nil, resp, nil
}

func (s *Server) secretUpdate(ctx context.Context, _ *mcp.CallToolRequest, in secretUpdateInput) (*mcp.CallToolResult, dto.SecretMetadata, error) {
	if err := validateSecretScope(in.ProductID, in.ProjectID, in.Environment); err != nil {
		return nil, dto.SecretMetadata{}, err
	}
	if strings.TrimSpace(in.Key) == "" {
		return nil, dto.SecretMetadata{}, fmt.Errorf("key is required")
	}
	body, err := encodeJSON(dto.UpdateSecret{Value: in.Value})
	if err != nil {
		return nil, dto.SecretMetadata{}, err
	}
	raw, _, err := s.c.Do(ctx, "PUT", secretKeyPath(in.ProductID, in.ProjectID, in.Environment, in.Key), body)
	if err != nil {
		return nil, dto.SecretMetadata{}, err
	}
	meta, err := decodeJSON[dto.SecretMetadata](raw)
	if err != nil {
		return nil, dto.SecretMetadata{}, err
	}
	return nil, meta, nil
}

func (s *Server) secretDelete(ctx context.Context, _ *mcp.CallToolRequest, in secretKeyInput) (*mcp.CallToolResult, deleteKeyOutput, error) {
	if err := validateSecretScope(in.ProductID, in.ProjectID, in.Environment); err != nil {
		return nil, deleteKeyOutput{}, err
	}
	if strings.TrimSpace(in.Key) == "" {
		return nil, deleteKeyOutput{}, fmt.Errorf("key is required")
	}
	if _, _, err := s.c.Do(ctx, "DELETE", secretKeyPath(in.ProductID, in.ProjectID, in.Environment, in.Key), nil); err != nil {
		return nil, deleteKeyOutput{}, err
	}
	return nil, deleteKeyOutput{Deleted: true, Key: in.Key}, nil
}

// secretExport intentionally returns plaintext dotenv content: export is an
// explicit, deliberate MCP tool action, not incidental exposure.
func (s *Server) secretExport(ctx context.Context, _ *mcp.CallToolRequest, in secretExportInput) (*mcp.CallToolResult, secretExportOutput, error) {
	if err := validateSecretScope(in.ProductID, in.ProjectID, in.Environment); err != nil {
		return nil, secretExportOutput{}, err
	}
	raw, _, err := s.c.Do(ctx, "GET", secretExportPath(in.ProductID, in.ProjectID, in.Environment), nil)
	if err != nil {
		return nil, secretExportOutput{}, err
	}
	return nil, secretExportOutput{Dotenv: string(raw)}, nil
}

// deleteKeyOutput mirrors deleteOutput but reports the deleted key (secrets
// have no numeric id on the delete path — the key is the identifier).
type deleteKeyOutput struct {
	Deleted bool   `json:"deleted" jsonschema:"true when the resource was deleted"`
	Key     string `json:"key" jsonschema:"key of the deleted secret"`
}
