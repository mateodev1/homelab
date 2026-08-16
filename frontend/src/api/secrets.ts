import type {
  SecretEnvironment,
  SecretEnvironmentName,
  SecretMetadata,
  SecretProduct,
  SecretProject,
  SecretProjectCreated,
  SecretReveal,
} from '../types/secrets';
import { ApiError } from '../types/todo';

async function parseJSON<T>(response: Response): Promise<T> {
  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }
  return response.json() as Promise<T>;
}

function authorizedHeaders(token: string, extra?: Record<string, string>): Record<string, string> {
  return { Authorization: `Bearer ${token}`, ...extra };
}

function productProjectsPath(productId: number): string {
  return `/api/products/${productId}/projects`;
}

function environmentsPath(productId: number, projectId: number): string {
  return `${productProjectsPath(productId)}/${projectId}/environments`;
}

function secretsPath(
  productId: number,
  projectId: number,
  environment: SecretEnvironmentName,
): string {
  return `${environmentsPath(productId, projectId)}/${encodeURIComponent(environment)}/secrets`;
}

export async function getSecretProducts(token: string): Promise<SecretProduct[]> {
  const response = await fetch('/api/products', { headers: authorizedHeaders(token) });
  return parseJSON<SecretProduct[]>(response);
}

export async function createSecretProduct(token: string, name: string): Promise<SecretProduct> {
  const response = await fetch('/api/products', {
    method: 'POST',
    headers: authorizedHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify({ name }),
  });
  return parseJSON<SecretProduct>(response);
}

export async function getSecretProjects(
  token: string,
  productId: number,
): Promise<SecretProject[]> {
  const response = await fetch(productProjectsPath(productId), {
    headers: authorizedHeaders(token),
  });
  return parseJSON<SecretProject[]>(response);
}

export async function createSecretProject(
  token: string,
  productId: number,
  name: string,
): Promise<SecretProjectCreated> {
  const response = await fetch(productProjectsPath(productId), {
    method: 'POST',
    headers: authorizedHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify({ name }),
  });
  return parseJSON<SecretProjectCreated>(response);
}

export async function getSecretEnvironments(
  token: string,
  productId: number,
  projectId: number,
): Promise<SecretEnvironment[]> {
  const response = await fetch(environmentsPath(productId, projectId), {
    headers: authorizedHeaders(token),
  });
  return parseJSON<SecretEnvironment[]>(response);
}

export async function getSecrets(
  token: string,
  productId: number,
  projectId: number,
  environment: SecretEnvironmentName,
): Promise<SecretMetadata[]> {
  const response = await fetch(secretsPath(productId, projectId, environment), {
    headers: authorizedHeaders(token),
  });
  return parseJSON<SecretMetadata[]>(response);
}

export async function createSecret(
  token: string,
  productId: number,
  projectId: number,
  environment: SecretEnvironmentName,
  key: string,
  value: string,
): Promise<SecretMetadata> {
  const response = await fetch(secretsPath(productId, projectId, environment), {
    method: 'POST',
    headers: authorizedHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify({ key, value }),
  });
  return parseJSON<SecretMetadata>(response);
}

export async function updateSecret(
  token: string,
  productId: number,
  projectId: number,
  environment: SecretEnvironmentName,
  key: string,
  value: string,
): Promise<SecretMetadata> {
  const response = await fetch(
    `${secretsPath(productId, projectId, environment)}/${encodeURIComponent(key)}`,
    {
      method: 'PUT',
      headers: authorizedHeaders(token, { 'Content-Type': 'application/json' }),
      body: JSON.stringify({ value }),
    },
  );
  return parseJSON<SecretMetadata>(response);
}

export async function deleteSecret(
  token: string,
  productId: number,
  projectId: number,
  environment: SecretEnvironmentName,
  key: string,
): Promise<void> {
  const response = await fetch(
    `${secretsPath(productId, projectId, environment)}/${encodeURIComponent(key)}`,
    {
      method: 'DELETE',
      headers: authorizedHeaders(token),
    },
  );
  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }
}

export async function revealSecret(
  token: string,
  productId: number,
  projectId: number,
  environment: SecretEnvironmentName,
  key: string,
): Promise<SecretReveal> {
  const response = await fetch(
    `${secretsPath(productId, projectId, environment)}/${encodeURIComponent(key)}/reveal`,
    { headers: authorizedHeaders(token) },
  );
  return parseJSON<SecretReveal>(response);
}

export async function exportSecrets(
  token: string,
  productId: number,
  projectId: number,
  environment: SecretEnvironmentName,
): Promise<string> {
  const response = await fetch(
    `${environmentsPath(productId, projectId)}/${encodeURIComponent(environment)}/export`,
    { headers: authorizedHeaders(token) },
  );
  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }
  return response.text();
}
