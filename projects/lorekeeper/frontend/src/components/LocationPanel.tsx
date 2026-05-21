import { useState, useEffect } from 'react';
import { type Location } from '../types';
import { getAll, create, remove } from '../api';
import EntityCard from './EntityCard';

interface FormState {
  name: string;
  region: string;
  description: string;
}

const EMPTY: FormState = { name: '', region: '', description: '' };

export default function LocationPanel() {
  const [locations, setLocations] = useState<Location[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(EMPTY);
  const [submitting, setSubmitting] = useState(false);

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const data = await getAll<Location>('locations');
      setLocations(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load locations');
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
      await create<Location>('locations', form);
      setForm(EMPTY);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create location');
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    try {
      await remove('locations', id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to delete location');
    }
  }

  return (
    <div>
      {/* Create Form */}
      <div className="create-form-wrapper">
        <div className="create-form-title">✦ New Location</div>
        <form onSubmit={handleSubmit}>
          <div className="form-grid">
            <div className="form-group">
              <label>Name</label>
              <input
                name="name"
                value={form.name}
                onChange={handleChange}
                placeholder="Eldenmoor"
                required
              />
            </div>
            <div className="form-group">
              <label>Region</label>
              <input
                name="region"
                value={form.region}
                onChange={handleChange}
                placeholder="Northern Reaches"
              />
            </div>
            <div className="form-group full-width">
              <label>Description</label>
              <textarea
                name="description"
                value={form.description}
                onChange={handleChange}
                placeholder="A fog-shrouded settlement at the edge of the ancient wood…"
              />
            </div>
          </div>
          <div className="form-actions">
            <button className="form-submit" type="submit" disabled={submitting}>
              {submitting ? 'Adding…' : 'Add Location'}
            </button>
          </div>
        </form>
      </div>

      {/* Error */}
      {error && <div className="error-banner">{error}</div>}

      {/* List */}
      {loading ? (
        <div className="loading-text">Loading locations…</div>
      ) : locations.length === 0 ? (
        <div className="empty-state">No locations yet. Create one above!</div>
      ) : (
        <div className="card-grid">
          {locations.map(l => (
            <EntityCard
              key={l.id}
              title={l.name}
              badges={l.region ? [{ label: l.region }] : []}
              description={l.description}
              onDelete={() => handleDelete(l.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
