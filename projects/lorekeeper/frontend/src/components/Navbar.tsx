interface NavbarProps {
  activeSection: string;
}

const sectionLabels: Record<string, string> = {
  characters: 'Characters',
  locations: 'Locations',
  factions: 'Factions',
  events: 'Events',
};

export default function Navbar({ activeSection }: NavbarProps) {
  return (
    <header
      style={{
        position: 'fixed',
        top: 0,
        left: 240,
        right: 0,
        height: 60,
        background: 'var(--bg-secondary)',
        borderBottom: '1px solid var(--border)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 1.5rem',
        zIndex: 99,
      }}
    >
      <span style={{ fontWeight: 700, fontSize: '1.1rem', color: 'var(--text-primary)' }}>
        <span style={{ color: 'var(--accent)', marginRight: '0.4rem' }}>⚔</span>
        Lorekeeper
      </span>
      <span
        style={{
          fontSize: '0.85rem',
          color: 'var(--text-muted)',
          textTransform: 'uppercase',
          letterSpacing: '0.08em',
        }}
      >
        {sectionLabels[activeSection] ?? activeSection}
      </span>
    </header>
  );
}
