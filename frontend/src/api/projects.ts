import type { CreateProjectPayload, Project, UpdateProjectPayload } from '../types/project';
import { ApiError } from '../types/todo';

async function parseResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.json() as Promise<T>;
}

function authorizedHeaders(token: string, extra?: Record<string, string>): Record<string, string> {
  return { Authorization: `Bearer ${token}`, ...extra };
}

export async function getProjects(token: string, signal?: AbortSignal): Promise<Project[]> {
  const response = await fetch('/api/projects', { headers: authorizedHeaders(token), signal });
  return parseResponse<Project[]>(response);
}

export async function createProject(
  token: string,
  payload: CreateProjectPayload,
): Promise<Project> {
  const response = await fetch('/api/projects', {
    method: 'POST',
    headers: authorizedHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify(payload),
  });

  return parseResponse<Project>(response);
}

export async function updateProject(
  token: string,
  id: number,
  payload: UpdateProjectPayload,
): Promise<Project> {
  const response = await fetch(`/api/projects/${id}`, {
    method: 'PUT',
    headers: authorizedHeaders(token, { 'Content-Type': 'application/json' }),
    body: JSON.stringify(payload),
  });

  return parseResponse<Project>(response);
}

export async function deleteProject(token: string, id: number): Promise<void> {
  const response = await fetch(`/api/projects/${id}`, {
    method: 'DELETE',
    headers: authorizedHeaders(token),
  });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }
}
