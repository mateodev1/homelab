export type SecretEnvironmentName = 'development' | 'staging' | 'production';

export interface SecretProduct {
  id: number;
  name: string;
  created_at: string;
}

export interface SecretProject {
  id: number;
  product_id: number;
  name: string;
  created_at: string;
}

export interface SecretEnvironment {
  id: number;
  project_id: number;
  name: SecretEnvironmentName;
  created_at: string;
}

export interface SecretMetadata {
  environment_id: number;
  key: string;
  value: string;
  created_at: string;
  updated_at: string;
}

export interface SecretProjectCreated extends SecretProject {
  environments: SecretEnvironment[];
}

export interface SecretReveal {
  key: string;
  value: string;
}
