package equipment

import (
	"errors"
	"time"
)

var (
	ErrNotFound       = errors.New("equipment not found")
	ErrInvalidStatus   = errors.New("invalid equipment status")
	ErrInvalidCategory = errors.New("invalid equipment category")
)

var validStatuses = map[string]bool{
	"ok": true, "to-check": true, "hs": true, "repair": true,
}
var validCategories = map[string]bool{
	"sound": true, "light": true, "video": true,
	"set": true, "safety": true, "rigging": true,
}

type Position struct {
	X float64  `json:"x"`
	Y float64  `json:"y"`
	Z *float64 `json:"z,omitempty"`
}

type doc struct {
	ID                string     `json:"_id"`
	Rev               string     `json:"_rev,omitempty"`
	Type              string     `json:"type"`
	Name              string     `json:"name"`
	Category          string     `json:"category"`
	QRCode            string     `json:"qrCode"`
	Status            string     `json:"status"`
	Location          string     `json:"location"`
	Zone              string     `json:"zone,omitempty"`
	EventID           string     `json:"eventId,omitempty"`
	ResponsiblePerson string     `json:"responsiblePerson,omitempty"`
	LastCheck         *time.Time `json:"lastCheck,omitempty"`
	Position          *Position  `json:"position,omitempty"`
	Notes             string     `json:"notes,omitempty"`
}

type Equipment struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Category          string     `json:"category"`
	QRCode            string     `json:"qrCode"`
	Status            string     `json:"status"`
	Location          string     `json:"location"`
	Zone              string     `json:"zone,omitempty"`
	EventID           string     `json:"eventId,omitempty"`
	ResponsiblePerson string     `json:"responsiblePerson,omitempty"`
	LastCheck         *time.Time `json:"lastCheck,omitempty"`
	Position          *Position  `json:"position,omitempty"`
	Notes             string     `json:"notes,omitempty"`
}

type Input struct {
	Name              string     `json:"name"`
	Category          string     `json:"category"`
	QRCode            string     `json:"qrCode"`
	Status            string     `json:"status"`
	Location          string     `json:"location"`
	Zone              string     `json:"zone"`
	EventID           string     `json:"eventId"`
	ResponsiblePerson string     `json:"responsiblePerson"`
	LastCheck         *time.Time `json:"lastCheck"`
	Position          *Position  `json:"position"`
	Notes             string     `json:"notes"`
}

func toPublic(d *doc) *Equipment {
	return &Equipment{
		ID:                d.ID,
		Name:              d.Name,
		Category:          d.Category,
		QRCode:            d.QRCode,
		Status:            d.Status,
		Location:          d.Location,
		Zone:              d.Zone,
		EventID:           d.EventID,
		ResponsiblePerson: d.ResponsiblePerson,
		LastCheck:         d.LastCheck,
		Position:          d.Position,
		Notes:             d.Notes,
	}
}

func toPublicSlice(docs []doc) []Equipment {
	out := make([]Equipment, len(docs))
	for i := range docs {
		out[i] = *toPublic(&docs[i])
	}
	return out
}
