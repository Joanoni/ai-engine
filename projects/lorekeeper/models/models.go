package models

import (
	"crypto/rand"
	"fmt"
)

// newUUID generates a random UUID v4.
func NewUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Character represents a person or creature in the world.
type Character struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Faction     string `json:"faction"`
}

// Location represents a place in the world.
type Location struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	Description string `json:"description"`
}

// Faction represents an organisation or group.
type Faction struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Alignment   string `json:"alignment"`
	Description string `json:"description"`
}

// Event represents a historical or ongoing occurrence.
type Event struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Date         string   `json:"date"`
	Description  string   `json:"description"`
	Participants []string `json:"participants"`
}
