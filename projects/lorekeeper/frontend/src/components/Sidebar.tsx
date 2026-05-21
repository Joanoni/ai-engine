import { type EntityType } from '../types';

interface SidebarProps {
  active: EntityType;
  onSelect: (s: EntityType) => void;
}

const NAV_ITEMS: { key: EntityType; label: string; icon: string }[] = [
  { key: 'characters', label: 'Characters', icon: '👤' },
  { key: 'locations',  label: 'Locations',  icon: '🗺️' },
  { key: 'factions',   label: 'Factions',   icon: '⚑' },
  { key: 'events',     label: 'Events',     icon: '📜' },
  { key: 'graph',      label: 'Graph',      icon: '🕸️' },
];

export default function Sidebar({ active, onSelect }: SidebarProps) {
  return (
    <nav className="sidebar">
      <div className="sidebar-logo">
        <h1>⚔️ Lorekeeper</h1>
        <p>World Builder</p>
      </div>
      <div className="sidebar-nav">
        {NAV_ITEMS.map(({ key, label, icon }) => (
          <button
            key={key}
            className={`sidebar-link${active === key ? ' active' : ''}`}
            onClick={() => onSelect(key)}
          >
            <span className="link-icon">{icon}</span>
            <span className="link-label">{label}</span>
          </button>
        ))}
      </div>
    </nav>
  );
}
