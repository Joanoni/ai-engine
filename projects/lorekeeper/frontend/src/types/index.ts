export interface Character {
  id: string;
  name: string;
  type: string;
  description: string;
  faction: string;
}

export interface Location {
  id: string;
  name: string;
  region: string;
  description: string;
}

export interface Faction {
  id: string;
  name: string;
  alignment: string;
  description: string;
}

export interface Event {
  id: string;
  title: string;
  date: string;
  description: string;
  participants: string[];
}

export type EntityType = 'characters' | 'locations' | 'factions' | 'events' | 'graph';

export interface GraphNode {
  id: string;
  label: string;
  type: 'character' | 'location' | 'faction' | 'event';
}

export interface GraphEdge {
  source: string;
  target: string;
  label: string;
}

export interface GraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
}
