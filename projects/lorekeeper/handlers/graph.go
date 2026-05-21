package handlers

import (
	"net/http"
	"strings"

	"lorekeeper/models"
	"lorekeeper/store"
)

// graphNode is a vertex in the relationship graph.
type graphNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// graphEdge is a directed relationship between two nodes.
type graphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

// graphResponse is the top-level payload returned by GET /graph.
type graphResponse struct {
	Nodes []graphNode `json:"nodes"`
	Edges []graphEdge `json:"edges"`
}

// RegisterGraph registers the GET /graph route on mux.
func RegisterGraph(mux *http.ServeMux) {
	mux.HandleFunc("/graph", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleGraph(w, r)
	})
}

func handleGraph(w http.ResponseWriter, _ *http.Request) {
	// ------------------------------------------------------------------ load
	characters, err := store.Load[models.Character]("characters.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	locations, err := store.Load[models.Location]("locations.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	factions, err := store.Load[models.Faction]("factions.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	events, err := store.Load[models.Event]("events.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ------------------------------------------------------------------ nodes
	nodes := []graphNode{}

	for _, c := range characters {
		nodes = append(nodes, graphNode{ID: c.ID, Label: c.Name, Type: "character"})
	}
	for _, l := range locations {
		nodes = append(nodes, graphNode{ID: l.ID, Label: l.Name, Type: "location"})
	}
	for _, f := range factions {
		nodes = append(nodes, graphNode{ID: f.ID, Label: f.Name, Type: "faction"})
	}
	for _, e := range events {
		nodes = append(nodes, graphNode{ID: e.ID, Label: e.Title, Type: "event"})
	}

	// ------------------------------------------------------------------ edges
	edges := []graphEdge{}

	// Rule 1 — Character → Faction ("member of")
	for _, c := range characters {
		if c.Faction == "" {
			continue
		}
		for _, f := range factions {
			if strings.EqualFold(c.Faction, f.Name) {
				edges = append(edges, graphEdge{
					Source: c.ID,
					Target: f.ID,
					Label:  "member of",
				})
				break // a character belongs to at most one faction
			}
		}
	}

	// Rule 2 — Event → Character ("participant")
	for _, e := range events {
		for _, participant := range e.Participants {
			for _, c := range characters {
				if strings.EqualFold(participant, c.Name) {
					edges = append(edges, graphEdge{
						Source: e.ID,
						Target: c.ID,
						Label:  "participant",
					})
					break // matched; move to next participant
				}
			}
		}
	}

	// Rule 3 — Event → Location ("located at")
	for _, e := range events {
		titleLower := strings.ToLower(e.Title)
		descLower := strings.ToLower(e.Description)
		for _, l := range locations {
			locLower := strings.ToLower(l.Name)
			if strings.Contains(titleLower, locLower) || strings.Contains(descLower, locLower) {
				edges = append(edges, graphEdge{
					Source: e.ID,
					Target: l.ID,
					Label:  "located at",
				})
			}
		}
	}

	jsonResponse(w, http.StatusOK, graphResponse{Nodes: nodes, Edges: edges})
}
