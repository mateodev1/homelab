export interface Project {
  id: number;
  name: string;
  color: string;
  created_at: string;
}

export interface CreateProjectPayload {
  name: string;
  color?: string;
}

export interface UpdateProjectPayload {
  name?: string;
  color?: string;
}
