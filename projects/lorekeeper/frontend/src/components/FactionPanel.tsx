import { useState, useEffect } from 'react';
import { type Faction } from '../types';
import { getAll, create, remove } from '../api';
import EntityCard from './EntityCard';

interface FormState {
  name: string;
  alignment: string;
  description: string;
}

const EMPTY: FormState = { name: '', alignment: '', description: '' };

export default function FactionPanel() {
  const [factions, setFactions] = useState<Faction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(EMPTY);
  const [submitting, setSubmitting] = useState(false);

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const data = await getAll<Faction>('factions');
      setFactions(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load factions');
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
      await create<Faction>('factions', form);
      setForm(EMPTY);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create faction');
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    try {
      await remove('factions', id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to delete faction');
    }
  }

  return (
    <div>
      {/* Create Form */}
      <div className="create-form-wrapper">
        <div className="create-form-title">✦ New Faction</div>
        <form onSubmit={handleSubmit}>
          <div className="form-grid">
            <div className="form-group">
              <label>Name</label>
              <input
                name="name"
                value={form.name}
                onChange={handleChange}
                placeholder="Verdant Accord"
                required
              />
            </div>
            <div className="form-group">
              <label>Alignment</label>
              <input
                name="alignment"
                value={form.alignment}
                onChange={handleChange}
                placeholder="Neutral Good"
              />
            </div>
            <div className="form-group full-width">
              <label>Description</label>
              <textarea
                name="description"
                value={form.description}
                onChange={handleChange}
                placeholder="A coalition of forest guardians who protect the ancient groves…"
              />
            </div>
          </div>
          <div className="form-actions">
            <button className="form-submit" type="submit" disabled={submitting}>
              {submitting ? 'Adding…' : 'Add Faction'}
            </button>
          </div>
        </form>
      </div>

      {/* Error */}
      {error && <div className="error-banner">{error}</div>}

      {/* List */}
      {loading ? (
        <div className="loading-text">Loading factions…</div>
      ) : factions.length === 0 ? (
        <div className="empty-state">No factions yet. Create one above!</div>
      ) : (
        <div className="card-grid">
          {factions.map(f => (
            <EntityCard
              key={f.id}
              title={f.name}
              badges={f.alignment ? [{ label: f.alignment }] : []}
              description={f.description}
              onDelete={() => handleDelete(f.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
