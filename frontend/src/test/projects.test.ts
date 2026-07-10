import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createProject, deleteProject, getProjects, updateProject } from '../api/projects';
import type { Project } from '../types/project';
import { ApiError } from '../types/todo';

const mockFetch = vi.fn();

function makeProject(overrides: Partial<Project> = {}): Project {
  return {
    id: 1,
    name: 'Homelab',
    color: 'default',
    created_at: '2026-06-20T10:00:00Z',
    ...overrides,
  };
}

describe('projects API client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
    mockFetch.mockReset();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('getProjects returns projects on success', async () => {
    const projects: Project[] = [makeProject()];

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(projects),
    });

    await expect(getProjects('test-token')).resolves.toEqual(projects);
    expect(mockFetch).toHaveBeenCalledWith('/api/projects', {
      headers: { Authorization: 'Bearer test-token' },
      signal: undefined,
    });
  });

  it('getProjects throws ApiError on non-2xx response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: vi.fn().mockResolvedValueOnce('boom'),
    });

    await expect(getProjects('test-token')).rejects.toEqual(new ApiError(500, 'boom'));
  });

  it('createProject sends payload, token and returns project', async () => {
    const created = makeProject({ id: 2, name: 'Created' });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(created),
    });

    await expect(createProject('test-token', { name: 'Created' })).resolves.toEqual(created);
    expect(mockFetch).toHaveBeenCalledWith('/api/projects', {
      method: 'POST',
      headers: { Authorization: 'Bearer test-token', 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'Created' }),
    });
  });

  it('updateProject sends payload, token and returns updated project', async () => {
    const updated = makeProject({ id: 4, name: 'Renamed', color: 'blue' });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: vi.fn().mockResolvedValueOnce(updated),
    });

    await expect(
      updateProject('test-token', 4, { name: 'Renamed', color: 'blue' }),
    ).resolves.toEqual(updated);
    expect(mockFetch).toHaveBeenCalledWith('/api/projects/4', {
      method: 'PUT',
      headers: { Authorization: 'Bearer test-token', 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'Renamed', color: 'blue' }),
    });
  });

  it('deleteProject returns void for 204 and does not parse json', async () => {
    const jsonSpy = vi.fn();

    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: jsonSpy,
    });

    await expect(deleteProject('test-token', 5)).resolves.toBeUndefined();
    expect(mockFetch).toHaveBeenCalledWith('/api/projects/5', {
      method: 'DELETE',
      headers: { Authorization: 'Bearer test-token' },
    });
    expect(jsonSpy).not.toHaveBeenCalled();
  });

  it('deleteProject throws ApiError on non-2xx response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: vi.fn().mockResolvedValueOnce('delete failed'),
    });

    await expect(deleteProject('test-token', 5)).rejects.toEqual(
      new ApiError(500, 'delete failed'),
    );
  });
});
