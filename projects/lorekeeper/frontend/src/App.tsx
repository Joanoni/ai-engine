import { useState } from 'react';
import Sidebar from './components/Sidebar';
import Navbar from './components/Navbar';
import CharacterPanel from './components/CharacterPanel';
import LocationPanel from './components/LocationPanel';
import FactionPanel from './components/FactionPanel';
import EventPanel from './components/EventPanel';
import GraphPanel from './components/GraphPanel';
import { type EntityType } from './types';
import './styles/global.css';
import './styles/sidebar.css';
import './styles/card.css';
import './styles/form.css';

export default function App() {
  const [active, setActive] = useState<EntityType>('characters');

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <Sidebar active={active} onSelect={setActive} />
      <div style={{ marginLeft: 240, flex: 1, display: 'flex', flexDirection: 'column' }}>
        <Navbar activeSection={active} />
        <main style={{ padding: '1.5rem', marginTop: 60 }}>
          {active === 'characters' && <CharacterPanel />}
          {active === 'locations'  && <LocationPanel />}
          {active === 'factions'   && <FactionPanel />}
          {active === 'events'     && <EventPanel />}
          {active === 'graph'      && <GraphPanel />}
        </main>
      </div>
    </div>
  );
}
