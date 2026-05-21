import { useState, useEffect } from 'react';
import { type Event } from '../types';
import { getAll, create, remove } from '../api';
import EntityCard from './EntityCard';

interface FormState {
  title: string;
  date: string;
  description: string;
  participantsRaw: string;
}

const EMPTY: FormState = { title: '', date: '', description: '', participantsRaw: '' };

export default function EventPanel() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(EMPTY);
  const [submitting, setSubmitting] = useState(false);

  async function load() {
    setLoading(true);
    setError(null);
    try {
      const data = await getAll<Event>('events');
      setEvents(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load events');
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
    if (!form.title.trim()) return;
    setSubmitting(true);
    try {
      const participants = form.participantsRaw
        .split(',')
        .map(s => s.trim())
        .filter(Boolean);
      await create<Event>('events', {
        title: form.title,
        date: form.date,
        description: form.description,
        participants,
      });
      setForm(EMPTY);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create event');
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    try {
      await remove('events', id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to delete event');
    }
  }

  return (
    <div>
      {/* Create Form */}
      <div className="create-form-wrapper">
        <div className="create-form-title">✦ New Event</div>
        <form onSubmit={handleSubmit}>
          <div className="form-grid">
            <div className="form-group">
              <label>Title</label>
              <input
                name="title"
                value={form.title}
                onChange={handleChange}
                placeholder="Siege of Eldenmoor"
                required
              />
            </div>
            <div className="form-group">
              <label>Date</label>
              <input
                name="date"
                value={form.date}
                onChange={handleChange}
                placeholder="Year 312, Age of Embers"
              />
            </div>
            <div className="form-group full-width">
              <label>Participants (comma-separated)</label>
              <input
                name="participantsRaw"
                value={form.participantsRaw}
                onChange={handleChange}
                placeholder="Aria Dawnwhisper, Gornak the Unyielding"
              />
            </div>
            <div className="form-group full-width">
              <label>Description</label>
              <textarea
                name="description"
                value={form.description}
                onChange={handleChange}
                placeholder="A pivotal battle that reshaped the political landscape…"
              />
            </div>
          </div>
          <div className="form-actions">
            <button className="form-submit" type="submit" disabled={submitting}>
              {submitting ? 'Adding…' : 'Add Event'}
            </button>
          </div>
        </form>
      </div>

      {/* Error */}
      {error && <div className="error-banner">{error}</div>}

      {/* List */}
      {loading ? (
        <div className="loading-text">Loading events…</div>
      ) : events.length === 0 ? (
        <div className="empty-state">No events yet. Create one above!</div>
      ) : (
        <div className="card-grid">
          {events.map(ev => (
            <EntityCard
              key={ev.id}
              title={ev.title}
              badges={[
                ...(ev.date ? [{ label: ev.date }] : []),
                ...(ev.participants.length > 0
                  ? [{ label: `${ev.participants.length} participant${ev.participants.length !== 1 ? 's' : ''}`, muted: true }]
                  : []),
              ]}
              description={ev.description}
              onDelete={() => handleDelete(ev.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
