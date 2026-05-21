import { useState, useEffect } from 'react';
import { type Character } from '../types';
import { getAll, create, remove } from '../api';
import EntityCard from './EntityCard';

interface FormState {
  name: string;
  type: string;
  description: string;
  faction: string;
}

const EMPTY: FormState = { name: '', type: '', description: '', faction: '' };

export default function CharacterPanel() {
  const [characters, setCharacters] = useState<Character[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(EMPTY);
  const [submitting, setSubmitting] = useState(false);

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const data = await getAll<Character>('characters');
      setCharacters(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load characters');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { void load(); }, []);

  function handleChange(e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) {
    setForm(prev => ({ ...prev, [e.target.name]: e.target.value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!form.name.trim()) return;
    setSubmitting(true);
    try {
      await create<Character>('characters', form);
      setForm(EMPTY);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create character');
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    try {
      await remove('characters', id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to delete character');
    }
  }

  return (
    <div>
      {/* Create Form */}
      <div className="create-form-wrapper">
        <div className="create-form-title">✦ New Character</div>
        <form onSubmit={handleSubmit}>
          <div className="form-grid">
            <div className="form-group">
              <label>Name</label>
              <input
                name="name"
                value={form.name}
                onChange={handleChange}
                placeholder="Aria Dawnwhisper"
                required
              />
            </div>
            <div className="form-group">
              <label>Type</label>
              <input
                name="type"
                value={form.type}
                onChange={handleChange}
                placeholder="Elf Ranger"
              />
            </div>
            <div className="form-group">
              <label>Faction</label>
              <input
                name="faction"
                value={form.faction}
                onChange={handleChange}
                placeholder="Verdant Accord"
              />
            </div>
            <div className="form-group full-width">
              <label>Description</label>
              <textarea
                name="description"
                value={form.description}
                onChange={handleChange}
                placeholder="A seasoned scout who patrols the ancient forests…"
              />
            </div>
          </div>
          <div className="form-actions">
            <button className="form-submit" type="submit" disabled={submitting}>
              {submitting ? 'Adding…' : 'Add Character'}
            </button>
          </div>
        </form>
      </div>

      {/* Error */}
      {error && <div className="error-banner">{error}</div>}

      {/* List */}
      {loading ? (
        <div className="loading-text">Loading characters…</div>
      ) : characters.length === 0 ? (
        <div className="empty-state">No characters yet. Create one above!</div>
      ) : (
        <div className="card-grid">
          {characters.map(c => (
            <EntityCard
              key={c.id}
              title={c.name}
              badges={[
                ...(c.type ? [{ label: c.type }] : []),
                ...(c.faction ? [{ label: c.faction, muted: true }] : []),
              ]}
              description={c.description}
              onDelete={() => handleDelete(c.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
