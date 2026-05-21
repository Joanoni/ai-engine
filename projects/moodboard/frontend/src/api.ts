import { Board, Reference } from './types';

const BASE = 'http://localhost:3000';

export async function getBoards(): Promise<Board[]> {
  const res = await fetch(`${BASE}/boards`);
  if (!res.ok) throw new Error('Failed to fetch boards');
  return res.json();
}

export async function createBoard(name: string, description: string): Promise<Board> {
  const res = await fetch(`${BASE}/boards`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, description }),
  });
  if (!res.ok) throw new Error('Failed to create board');
  return res.json();
}

export async function deleteBoard(id: string): Promise<void> {
  const res = await fetch(`${BASE}/boards/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error('Failed to delete board');
}

export async function getReferences(boardId: string): Promise<Reference[]> {
  const res = await fetch(`${BASE}/boards/${boardId}/references`);
  if (!res.ok) throw new Error('Failed to fetch references');
  return res.json();
}

export async function createReference(
  boardId: string,
  type: Reference['type'],
  content: string,
  label: string,
): Promise<Reference> {
  const res = await fetch(`${BASE}/boards/${boardId}/references`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ type, content, label }),
  });
  if (!res.ok) throw new Error('Failed to create reference');
  return res.json();
}

export async function deleteReference(id: string): Promise<void> {
  const res = await fetch(`${BASE}/references/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error('Failed to delete reference');
}
