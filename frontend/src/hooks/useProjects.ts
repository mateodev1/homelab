import { useAuth0 } from '@auth0/auth0-react';
import { useEffect, useState } from 'react';
import { createProject, deleteProject, getProjects, updateProject } from '../api/projects';
import type { Project } from '../types/project';

interface UseProjectsReturn {
  projects: Project[];
  loading: boolean;
  error: string | null;
  addProject: (name: string, color?: string) => Promise<void>;
  editProject: (id: number, changes: Partial<Pick<Project, 'name' | 'color'>>) => Promise<void>;
  removeProject: (id: number) => Promise<void>;
}

function toMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unknown error';
}

export function useProjects(): UseProjectsReturn {
  const { getAccessTokenSilently } = useAuth0();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    const loadProjects = async () => {
      try {
        setError(null);
        const token = await getAccessTokenSilently();
        const data = await getProjects(token, controller.signal);
        setProjects(data);
      } catch (err) {
        if (controller.signal.aborted) {
          return;
        }
        setError(toMessage(err));
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    };

    void loadProjects();

    return () => {
      controller.abort();
    };
  }, [getAccessTokenSilently]);

  const addProject = async (name: string, color = 'default') => {
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      const created = await createProject(token, { name, color });
      setProjects((current) => [...current, created]);
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const editProject = async (id: number, changes: Partial<Pick<Project, 'name' | 'color'>>) => {
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      const updated = await updateProject(token, id, changes);
      setProjects((current) => current.map((project) => (project.id === id ? updated : project)));
    } catch (err) {
      setError(toMessage(err));
    }
  };

  const removeProject = async (id: number) => {
    try {
      setError(null);
      const token = await getAccessTokenSilently();
      await deleteProject(token, id);
      setProjects((current) => current.filter((project) => project.id !== id));
    } catch (err) {
      setError(toMessage(err));
    }
  };

  return { projects, loading, error, addProject, editProject, removeProject };
}
