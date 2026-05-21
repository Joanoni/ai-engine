package main

import (
	"fmt"
	"log"
	"net/http"

	"lorekeeper/handlers"
	"lorekeeper/models"
	"lorekeeper/store"
)

// corsMiddleware adds CORS headers to every response and handles pre-flight requests.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Seed data files with sample content when they are empty.
	seed()

	mux := http.NewServeMux()

	// Register all entity routes.
	handlers.RegisterCharacters(mux)
	handlers.RegisterLocations(mux)
	handlers.RegisterFactions(mux)
	handlers.RegisterEvents(mux)
	handlers.RegisterGraph(mux)

	addr := ":3000"
	fmt.Printf("Lorekeeper API listening on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, corsMiddleware(mux)))
}

// seed populates each data file with starter records if the file is empty.
func seed() {
	seedCharacters()
	seedLocations()
	seedFactions()
	seedEvents()
}

func seedCharacters() {
	existing, _ := store.Load[models.Character]("characters.json")
	if len(existing) > 0 {
		return
	}
	characters := []models.Character{
		{
			ID:          models.NewUUID(),
			Name:        "Aria Dawnwhisper",
			Type:        "Elf Ranger",
			Description: "A keen-eyed elven ranger who guards the forest borders of the Northern Highlands.",
			Faction:     "Verdant Accord",
		},
		{
			ID:          models.NewUUID(),
			Name:        "Thorne Ashveil",
			Type:        "Human Rogue",
			Description: "A silver-tongued rogue whose loyalties shift like shadows — rumoured to work for multiple patrons.",
			Faction:     "Shadow Compact",
		},
		{
			ID:          models.NewUUID(),
			Name:        "Gornak the Unyielding",
			Type:        "Orc Warrior",
			Description: "A battle-hardened orc champion who has never lost a single duel in recorded history.",
			Faction:     "Iron Pact",
		},
	}
	if err := store.Save("characters.json", characters); err != nil {
		log.Printf("seed characters: %v", err)
	}
}

func seedLocations() {
	existing, _ := store.Load[models.Location]("locations.json")
	if len(existing) > 0 {
		return
	}
	locations := []models.Location{
		{
			ID:          models.NewUUID(),
			Name:        "Eldenmoor",
			Region:      "Northern Highlands",
			Description: "A mist-shrouded fortress town perched on the cliffs above the Elden River.",
		},
		{
			ID:          models.NewUUID(),
			Name:        "The Sunken Archives",
			Region:      "Deepwater Trench",
			Description: "Ancient library ruins submerged beneath the ocean, accessible only to those who can breathe water.",
		},
		{
			ID:          models.NewUUID(),
			Name:        "Ashveil Crossing",
			Region:      "Midland Plains",
			Description: "A busy crossroads town known for its black-market dealings and legendary tavern brawls.",
		},
	}
	if err := store.Save("locations.json", locations); err != nil {
		log.Printf("seed locations: %v", err)
	}
}

func seedFactions() {
	existing, _ := store.Load[models.Faction]("factions.json")
	if len(existing) > 0 {
		return
	}
	factions := []models.Faction{
		{
			ID:          models.NewUUID(),
			Name:        "Verdant Accord",
			Alignment:   "Neutral Good",
			Description: "A druidic alliance sworn to protect the ancient forests and their creatures from exploitation.",
		},
		{
			ID:          models.NewUUID(),
			Name:        "Shadow Compact",
			Alignment:   "Chaotic Neutral",
			Description: "A loose confederation of thieves, spies, and information brokers who sell secrets to the highest bidder.",
		},
		{
			ID:          models.NewUUID(),
			Name:        "Iron Pact",
			Alignment:   "Lawful Neutral",
			Description: "A martial brotherhood of warriors who uphold a strict code of honour above all political allegiances.",
		},
	}
	if err := store.Save("factions.json", factions); err != nil {
		log.Printf("seed factions: %v", err)
	}
}

func seedEvents() {
	existing, _ := store.Load[models.Event]("events.json")
	if len(existing) > 0 {
		return
	}
	events := []models.Event{
		{
			ID:           models.NewUUID(),
			Title:        "The Siege of Eldenmoor",
			Date:         "342 AE",
			Description:  "A brutal three-month siege that ended with a heroic last stand on the cliff walls.",
			Participants: []string{"Aria Dawnwhisper", "Gornak the Unyielding"},
		},
		{
			ID:           models.NewUUID(),
			Title:        "Treaty of the Silent Pines",
			Date:         "289 AE",
			Description:  "A landmark peace accord signed in the ancient pine forest, ending decades of border conflicts.",
			Participants: []string{"Aria Dawnwhisper", "Thorne Ashveil"},
		},
		{
			ID:           models.NewUUID(),
			Title:        "The Sunken Archives Expedition",
			Date:         "401 AE",
			Description:  "A solo expedition to recover lost arcane texts from the submerged ruins beneath the Deepwater Trench.",
			Participants: []string{"Thorne Ashveil"},
		},
	}
	if err := store.Save("events.json", events); err != nil {
		log.Printf("seed events: %v", err)
	}
}
