import type { GraphData } from '../types';

const BASE = 'http://localhost:3000';

export async function getAll<T>(entity: string): Promise<T[]> {
  const res = await fetch(`${BASE}/${entity}`);
  if (!res.ok) throw new Error(`Failed to fetch ${entity}: ${res.status}`);
  return res.json() as Promise<T[]>;
}

export async function create<T>(entity: string, body: Omit<T, 'id'>): Promise<T> {
  const res = await fetch(`${BASE}/${entity}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`Failed to create ${entity}: ${res.status}`);
  return res.json() as Promise<T>;
}

export async function remove(entity: string, id: string): Promise<void> {
  const res = await fetch(`${BASE}/${entity}/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error(`Failed to delete ${entity}/${id}: ${res.status}`);
}

export async function fetchGraph(): Promise<GraphData> {
  const res = await fetch(`${BASE}/graph`);
  if (!res.ok) throw new Error(`Failed to fetch graph: ${res.status}`);
  return res.json() as Promise<GraphData>;
}
