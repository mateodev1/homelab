import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  createSecret,
  exportSecrets,
  getSecrets,
  revealSecret,
  updateSecret,
} from '../api/secrets';
import type { SecretMetadata } from '../types/secrets';
import { ApiError } from '../types/todo';

const mockFetch = vi.fn();

function makeSecret(overrides: Partial<SecretMetadata> = {}): SecretMetadata {
  return {
    environment_id: 1,
    key: 'DATABASE_URL',
    value: '••••••••',
    created_at: '2026-06-20T10:00:00Z',
    updated_at: '2026-06-20T10:00:00Z',
    ...overrides,
  };
}

describe('secrets API client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
    mockFetch.mockReset();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('lists masked secrets for the selected environment', async () => {
    const secrets = [makeSecret()];
    mockFetch.mockResolvedValueOnce({ ok: true, json: vi.fn().mockResolvedValueOnce(secrets) });

    await expect(getSecrets('test-token', 2, 3, 'production')).resolves.toEqual(secrets);
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/products/2/projects/3/environments/production/secrets',
      {
        headers: { Authorization: 'Bearer test-token' },
      },
    );
  });

  it('creates and updates a secret without exposing the value in metadata', async () => {
    const secret = makeSecret();
    mockFetch
      .mockResolvedValueOnce({ ok: true, json: vi.fn().mockResolvedValueOnce(secret) })
      .mockResolvedValueOnce({ ok: true, json: vi.fn().mockResolvedValueOnce(secret) });

    await expect(
      createSecret('test-token', 2, 3, 'development', 'DATABASE_URL', 'postgres://db'),
    ).resolves.toEqual(secret);
    await expect(
      updateSecret('test-token', 2, 3, 'development', 'DATABASE_URL', 'postgres://new-db'),
    ).resolves.toEqual(secret);

    expect(mockFetch).toHaveBeenNthCalledWith(
      1,
      '/api/products/2/projects/3/environments/development/secrets',
      {
        method: 'POST',
        headers: { Authorization: 'Bearer test-token', 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: 'DATABASE_URL', value: 'postgres://db' }),
      },
    );
    expect(mockFetch).toHaveBeenNthCalledWith(
      2,
      '/api/products/2/projects/3/environments/development/secrets/DATABASE_URL',
      {
        method: 'PUT',
        headers: { Authorization: 'Bearer test-token', 'Content-Type': 'application/json' },
        body: JSON.stringify({ value: 'postgres://new-db' }),
      },
    );
  });

  it('reveals one value and exports the complete environment as dotenv text', async () => {
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: vi.fn().mockResolvedValueOnce({ key: 'API_KEY', value: 'secret' }),
      })
      .mockResolvedValueOnce({
        ok: true,
        text: vi.fn().mockResolvedValueOnce('API_KEY="secret"\n'),
      });

    await expect(revealSecret('test-token', 2, 3, 'staging', 'API_KEY')).resolves.toEqual({
      key: 'API_KEY',
      value: 'secret',
    });
    await expect(exportSecrets('test-token', 2, 3, 'staging')).resolves.toBe('API_KEY="secret"\n');

    expect(mockFetch).toHaveBeenNthCalledWith(
      1,
      '/api/products/2/projects/3/environments/staging/secrets/API_KEY/reveal',
      {
        headers: { Authorization: 'Bearer test-token' },
      },
    );
    expect(mockFetch).toHaveBeenNthCalledWith(
      2,
      '/api/products/2/projects/3/environments/staging/export',
      {
        headers: { Authorization: 'Bearer test-token' },
      },
    );
  });

  it('surfaces export errors without parsing them as JSON', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 403,
      text: vi.fn().mockResolvedValueOnce('forbidden'),
    });

    await expect(exportSecrets('test-token', 2, 3, 'production')).rejects.toEqual(
      new ApiError(403, 'forbidden'),
    );
  });
});
