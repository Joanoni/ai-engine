export interface Board {
  id: string;
  name: string;
  description: string;
  createdAt: string;
}

export interface Reference {
  id: string;
  boardId: string;
  type: 'image' | 'color' | 'note';
  content: string;
  label: string;
  createdAt: string;
}
