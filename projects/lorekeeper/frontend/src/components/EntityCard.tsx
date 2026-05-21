interface Badge {
  label: string;
  muted?: boolean;
}

interface EntityCardProps {
  title: string;
  badges: Badge[];
  description: string;
  onDelete: () => void;
}

export default function EntityCard({ title, badges, description, onDelete }: EntityCardProps) {
  function handleDelete() {
    if (window.confirm('Delete this entry?')) {
      onDelete();
    }
  }

  return (
    <div className="entity-card">
      <button className="delete-btn" onClick={handleDelete} title="Delete">
        🗑
      </button>
      <div className="card-name">{title}</div>
      {badges.length > 0 && (
        <div className="card-badges">
          {badges.map((b, i) => (
            <span key={i} className={`badge${b.muted ? ' muted' : ''}`}>
              {b.label}
            </span>
          ))}
        </div>
      )}
      <p className="card-desc">{description}</p>
    </div>
  );
}
