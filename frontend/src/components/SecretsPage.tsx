import { useAuth0 } from '@auth0/auth0-react';
import { Link } from '@tanstack/react-router';
import {
  ArrowLeft,
  Copy,
  Download,
  Eye,
  EyeOff,
  KeyRound,
  Pencil,
  Plus,
  Trash2,
} from 'lucide-react';
import { type FormEvent, useEffect, useState } from 'react';
import {
  createSecret,
  createSecretProduct,
  createSecretProject,
  deleteSecret,
  exportSecrets,
  getSecretEnvironments,
  getSecretProducts,
  getSecretProjects,
  getSecrets,
  revealSecret,
  updateSecret,
} from '../api/secrets';
import type {
  SecretEnvironment,
  SecretEnvironmentName,
  SecretMetadata,
  SecretProduct,
  SecretProject,
} from '../types/secrets';
import { Button } from './ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Select } from './ui/select';
import { Textarea } from './ui/textarea';

const ENVIRONMENT_LABELS: Record<SecretEnvironmentName, string> = {
  development: 'Development',
  staging: 'Staging',
  production: 'Production',
};

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Something went wrong';
}

export function SecretsPage() {
  const { getAccessTokenSilently } = useAuth0();
  const [products, setProducts] = useState<SecretProduct[]>([]);
  const [projects, setProjects] = useState<SecretProject[]>([]);
  const [environments, setEnvironments] = useState<SecretEnvironment[]>([]);
  const [secrets, setSecrets] = useState<SecretMetadata[]>([]);
  const [productId, setProductId] = useState<number | null>(null);
  const [projectId, setProjectId] = useState<number | null>(null);
  const [environment, setEnvironment] = useState<SecretEnvironmentName>('development');
  const [productName, setProductName] = useState('');
  const [projectName, setProjectName] = useState('');
  const [secretKey, setSecretKey] = useState('');
  const [secretValue, setSecretValue] = useState('');
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editingValue, setEditingValue] = useState('');
  const [revealed, setRevealed] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      try {
        setError(null);
        const token = await getAccessTokenSilently();
        const data = await getSecretProducts(token);
        if (cancelled) return;
        setProducts(data);
        setProductId((current) => current ?? data[0]?.id ?? null);
      } catch (err) {
        if (!cancelled) setError(errorMessage(err));
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [getAccessTokenSilently]);

  useEffect(() => {
    if (productId === null) {
      setProjects([]);
      setProjectId(null);
      return;
    }
    let cancelled = false;
    const load = async () => {
      try {
        setError(null);
        const token = await getAccessTokenSilently();
        const data = await getSecretProjects(token, productId);
        if (cancelled) return;
        setProjects(data);
        setProjectId((current) =>
          data.some((project) => project.id === current) ? current : (data[0]?.id ?? null),
        );
      } catch (err) {
        if (!cancelled) setError(errorMessage(err));
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [getAccessTokenSilently, productId]);

  useEffect(() => {
    if (productId === null || projectId === null) {
      setEnvironments([]);
      setSecrets([]);
      return;
    }
    let cancelled = false;
    const load = async () => {
      try {
        setError(null);
        const token = await getAccessTokenSilently();
        const data = await getSecretEnvironments(token, productId, projectId);
        if (cancelled) return;
        setEnvironments(data);
        setEnvironment((current) =>
          data.some((item) => item.name === current) ? current : (data[0]?.name ?? 'development'),
        );
      } catch (err) {
        if (!cancelled) setError(errorMessage(err));
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [getAccessTokenSilently, productId, projectId]);

  useEffect(() => {
    if (
      productId === null ||
      projectId === null ||
      !environments.some((item) => item.name === environment)
    ) {
      setSecrets([]);
      return;
    }
    let cancelled = false;
    const load = async () => {
      try {
        setError(null);
        const token = await getAccessTokenSilently();
        const data = await getSecrets(token, productId, projectId, environment);
        if (!cancelled) {
          setSecrets(data);
          setRevealed({});
        }
      } catch (err) {
        if (!cancelled) setError(errorMessage(err));
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [environment, environments, getAccessTokenSilently, productId, projectId]);

  const selectedProduct = products.find((product) => product.id === productId);
  const selectedProject = projects.find((project) => project.id === projectId);

  const refreshSecrets = async () => {
    if (productId === null || projectId === null) return;
    const token = await getAccessTokenSilently();
    setSecrets(await getSecrets(token, productId, projectId, environment));
    setRevealed({});
  };

  const handleCreateProduct = async (event: FormEvent) => {
    event.preventDefault();
    const name = productName.trim();
    if (!name) return;
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      const product = await createSecretProduct(token, name);
      setProducts((current) => [...current, product]);
      setProductId(product.id);
      setProductName('');
      setMessage(`Product ${product.name} created`);
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  const handleCreateProject = async (event: FormEvent) => {
    event.preventDefault();
    const name = projectName.trim();
    if (!name || productId === null) return;
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      const project = await createSecretProject(token, productId, name);
      setProjects((current) => [...current, project]);
      setProjectId(project.id);
      setProjectName('');
      setMessage(`Project ${project.name} created with three environments`);
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  const handleCreateSecret = async (event: FormEvent) => {
    event.preventDefault();
    const key = secretKey.trim();
    if (!key || productId === null || projectId === null) return;
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      await createSecret(token, productId, projectId, environment, key, secretValue);
      await refreshSecrets();
      setSecretKey('');
      setSecretValue('');
      setMessage(`${key} added to ${ENVIRONMENT_LABELS[environment]}`);
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  const handleUpdateSecret = async (key: string) => {
    if (productId === null || projectId === null) return;
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      await updateSecret(token, productId, projectId, environment, key, editingValue);
      await refreshSecrets();
      setEditingKey(null);
      setEditingValue('');
      setMessage(`${key} updated`);
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  const handleDeleteSecret = async (key: string) => {
    if (productId === null || projectId === null || !window.confirm(`Delete ${key}?`)) return;
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      await deleteSecret(token, productId, projectId, environment, key);
      await refreshSecrets();
      setMessage(`${key} deleted`);
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  const handleReveal = async (key: string) => {
    if (productId === null || projectId === null) return;
    try {
      const token = await getAccessTokenSilently();
      const result = await revealSecret(token, productId, projectId, environment, key);
      setRevealed((current) => ({ ...current, [key]: result.value }));
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  const handleCopyAll = async () => {
    if (productId === null || projectId === null) return;
    if (
      !window.confirm(
        `Copy all secrets from ${selectedProject?.name ?? 'this project'} / ${ENVIRONMENT_LABELS[environment]}?`,
      )
    ) {
      return;
    }
    try {
      const token = await getAccessTokenSilently();
      const body = await exportSecrets(token, productId, projectId, environment);
      await navigator.clipboard.writeText(body);
      setMessage('All environment secrets copied to the clipboard');
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  const handleDownload = async () => {
    if (productId === null || projectId === null) return;
    if (
      !window.confirm(
        `Download all secrets from ${selectedProject?.name ?? 'this project'} / ${ENVIRONMENT_LABELS[environment]}?`,
      )
    ) {
      return;
    }
    try {
      const token = await getAccessTokenSilently();
      const body = await exportSecrets(token, productId, projectId, environment);
      const url = URL.createObjectURL(new Blob([body], { type: 'text/plain;charset=utf-8' }));
      const link = document.createElement('a');
      link.href = url;
      link.download = '.env';
      link.click();
      URL.revokeObjectURL(url);
      setMessage('Environment downloaded as .env');
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  const handleCopySecret = async (key: string) => {
    if (productId === null || projectId === null) return;
    try {
      const token = await getAccessTokenSilently();
      const result = await revealSecret(token, productId, projectId, environment, key);
      await navigator.clipboard.writeText(result.value);
      setMessage(`${key} copied to the clipboard`);
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="flex h-14 items-center justify-between border-b border-border px-4 md:px-8">
        <div className="flex items-center gap-4">
          <Link
            to="/todos"
            className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft aria-hidden="true" className="size-4" />
            Tasks
          </Link>
          <span className="h-5 w-px bg-border" />
          <div className="flex items-center gap-2 text-sm font-semibold">
            <KeyRound aria-hidden="true" className="size-4" />
            Secrets
          </div>
        </div>
        <span className="text-xs text-muted-foreground">Values are encrypted at rest</span>
      </header>

      <main className="mx-auto grid w-full max-w-7xl gap-4 p-4 md:grid-cols-[18rem_1fr] md:p-8">
        <Card className="h-fit">
          <CardHeader>
            <CardTitle>Products</CardTitle>
            <CardDescription>Organize projects and environments.</CardDescription>
          </CardHeader>
          <CardContent>
            <Select
              aria-label="Product"
              value={productId ?? ''}
              onChange={(event) => setProductId(Number(event.target.value) || null)}
            >
              <option value="">Select a product</option>
              {products.map((product) => (
                <option key={product.id} value={product.id}>
                  {product.name}
                </option>
              ))}
            </Select>
            <form onSubmit={handleCreateProduct} className="flex gap-2">
              <Input
                aria-label="New product name"
                placeholder="New product"
                value={productName}
                onChange={(event) => setProductName(event.target.value)}
              />
              <Button type="submit" size="icon" variant="outline" aria-label="Create product">
                <Plus aria-hidden="true" />
              </Button>
            </form>

            {productId !== null ? (
              <>
                <Label className="pt-3">Projects</Label>
                <Select
                  aria-label="Secret project"
                  value={projectId ?? ''}
                  onChange={(event) => setProjectId(Number(event.target.value) || null)}
                >
                  <option value="">Select a project</option>
                  {projects.map((project) => (
                    <option key={project.id} value={project.id}>
                      {project.name}
                    </option>
                  ))}
                </Select>
                <form onSubmit={handleCreateProject} className="flex gap-2">
                  <Input
                    aria-label="New secret project name"
                    placeholder="New project"
                    value={projectName}
                    onChange={(event) => setProjectName(event.target.value)}
                  />
                  <Button type="submit" size="icon" variant="outline" aria-label="Create project">
                    <Plus aria-hidden="true" />
                  </Button>
                </form>
              </>
            ) : null}
          </CardContent>
        </Card>

        <section className="min-w-0 space-y-4">
          <Card>
            <CardHeader className="gap-3 md:flex-row md:items-end md:justify-between">
              <div>
                <CardTitle>
                  {selectedProject?.name ?? selectedProduct?.name ?? 'Select a project'}
                </CardTitle>
                <CardDescription>
                  {selectedProject
                    ? 'Manage secrets scoped to one environment.'
                    : 'Choose a product and project to begin.'}
                </CardDescription>
              </div>
              {projectId !== null ? (
                <div className="flex flex-wrap items-center gap-2">
                  <Select
                    aria-label="Environment"
                    value={environment}
                    onChange={(event) =>
                      setEnvironment(event.target.value as SecretEnvironmentName)
                    }
                    className="w-40"
                  >
                    {environments.map((item) => (
                      <option key={item.id} value={item.name}>
                        {ENVIRONMENT_LABELS[item.name]}
                      </option>
                    ))}
                  </Select>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => void handleCopyAll()}
                  >
                    <Copy aria-hidden="true" />
                    Copy all
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => void handleDownload()}
                  >
                    <Download aria-hidden="true" />
                    Download .env
                  </Button>
                </div>
              ) : null}
            </CardHeader>
          </Card>

          {error ? (
            <p className="rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
              {error}
            </p>
          ) : null}
          {message ? (
            <p className="rounded-md border border-border bg-muted p-3 text-sm text-muted-foreground">
              {message}
            </p>
          ) : null}

          {projectId !== null ? (
            <>
              <Card>
                <CardHeader>
                  <CardTitle>Add secret</CardTitle>
                  <CardDescription>
                    Add a key/value pair to the complete selected environment.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <form
                    onSubmit={handleCreateSecret}
                    className="grid gap-3 md:grid-cols-[minmax(12rem,0.7fr)_minmax(16rem,1.3fr)_auto] md:items-end"
                  >
                    <div className="grid gap-1.5">
                      <Label htmlFor="secret-key">Key</Label>
                      <Input
                        id="secret-key"
                        placeholder="DATABASE_URL"
                        value={secretKey}
                        onChange={(event) => setSecretKey(event.target.value)}
                      />
                    </div>
                    <div className="grid gap-1.5">
                      <Label htmlFor="secret-value">Value</Label>
                      <Textarea
                        id="secret-value"
                        rows={1}
                        placeholder="Secret value"
                        value={secretValue}
                        onChange={(event) => setSecretValue(event.target.value)}
                      />
                    </div>
                    <Button type="submit">
                      <Plus aria-hidden="true" />
                      Add secret
                    </Button>
                  </form>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>{ENVIRONMENT_LABELS[environment]} secrets</CardTitle>
                  <CardDescription>
                    {secrets.length} secret{secrets.length === 1 ? '' : 's'} in this environment.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {loading ? (
                    <p className="text-sm text-muted-foreground">Loading secrets...</p>
                  ) : null}
                  {!loading && secrets.length === 0 ? (
                    <p className="text-sm text-muted-foreground">No secrets yet.</p>
                  ) : null}
                  <div className="divide-y divide-border rounded-md border border-border">
                    {secrets.map((secret) => (
                      <div
                        key={secret.key}
                        className="grid gap-3 p-3 md:grid-cols-[minmax(10rem,0.7fr)_minmax(0,1fr)_auto] md:items-center"
                      >
                        <code className="break-all text-sm font-medium">{secret.key}</code>
                        {editingKey === secret.key ? (
                          <Input
                            aria-label={`New value for ${secret.key}`}
                            value={editingValue}
                            onChange={(event) => setEditingValue(event.target.value)}
                            placeholder="Enter replacement value"
                          />
                        ) : (
                          <code className="break-all text-sm text-muted-foreground">
                            {revealed[secret.key] ?? secret.value}
                          </code>
                        )}
                        <div className="flex justify-end gap-1">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label={`Copy ${secret.key}`}
                            onClick={() => void handleCopySecret(secret.key)}
                          >
                            <Copy aria-hidden="true" />
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label={
                              revealed[secret.key] ? `Hide ${secret.key}` : `Reveal ${secret.key}`
                            }
                            onClick={() => {
                              if (revealed[secret.key]) {
                                setRevealed((current) => {
                                  const next = { ...current };
                                  delete next[secret.key];
                                  return next;
                                });
                              } else {
                                void handleReveal(secret.key);
                              }
                            }}
                          >
                            {revealed[secret.key] ? (
                              <EyeOff aria-hidden="true" />
                            ) : (
                              <Eye aria-hidden="true" />
                            )}
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label={`Edit ${secret.key}`}
                            onClick={() => {
                              setEditingKey(secret.key);
                              setEditingValue('');
                            }}
                          >
                            <Pencil aria-hidden="true" />
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            aria-label={`Delete ${secret.key}`}
                            onClick={() => void handleDeleteSecret(secret.key)}
                          >
                            <Trash2 aria-hidden="true" />
                          </Button>
                          {editingKey === secret.key ? (
                            <Button
                              type="button"
                              size="sm"
                              onClick={() => void handleUpdateSecret(secret.key)}
                            >
                              Save
                            </Button>
                          ) : null}
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </>
          ) : null}
        </section>
      </main>
    </div>
  );
}
