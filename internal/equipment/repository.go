package equipment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stageops/backend/internal/couch"
)

const docType = "equipment"
const designDoc = "equipment"
const viewAll = "all"

type Repository struct {
	db *couch.Client
}

func NewRepository(cfg couch.Config) *Repository {
	return &Repository{db: couch.New(cfg)}
}

func (r *Repository) List(ctx context.Context) ([]Equipment, error) {
	var docs []doc
	if err := r.db.ListByView(ctx, designDoc, viewAll, &docs); err != nil {
		return nil, fmt.Errorf("listing equipment: %w", err)
	}
	return toPublicSlice(docs), nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Equipment, error) {
	var d doc
	if err := r.db.GetDoc(ctx, id, &d); err != nil {
		if errors.Is(err, couch.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("fetching equipment %s: %w", id, err)
	}
	return toPublic(&d), nil
}

func (r *Repository) Create(ctx context.Context, input *Input) (*Equipment, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	id := fmt.Sprintf("equipment::%s", uuid.New().String())
	now := time.Now().UTC()
	d := &doc{
		ID:                id,
		Type:              docType,
		Name:              input.Name,
		Category:          input.Category,
		QRCode:            input.QRCode,
		Status:            input.Status,
		Location:          input.Location,
		Zone:              input.Zone,
		EventID:           input.EventID,
		ResponsiblePerson: input.ResponsiblePerson,
		LastCheck:         input.LastCheck,
		Position:          input.Position,
		Notes:             input.Notes,
	}
	if d.LastCheck == nil {
		d.LastCheck = &now
	}

	if err := r.db.PutDoc(ctx, id, d); err != nil {
		return nil, fmt.Errorf("creating equipment: %w", err)
	}
	return toPublic(d), nil
}

func (r *Repository) Update(ctx context.Context, id string, input *Input) (*Equipment, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	// Fetch current doc to get _rev.
	var current doc
	if err := r.db.GetDoc(ctx, id, &current); err != nil {
		if errors.Is(err, couch.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("fetching equipment for update: %w", err)
	}

	current.Name = input.Name
	current.Category = input.Category
	current.QRCode = input.QRCode
	current.Status = input.Status
	current.Location = input.Location
	current.Zone = input.Zone
	current.EventID = input.EventID
	current.ResponsiblePerson = input.ResponsiblePerson
	current.LastCheck = input.LastCheck
	current.Position = input.Position
	current.Notes = input.Notes

	if err := r.db.PutDoc(ctx, id, &current); err != nil {
		return nil, fmt.Errorf("updating equipment: %w", err)
	}
	return toPublic(&current), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	var current doc
	if err := r.db.GetDoc(ctx, id, &current); err != nil {
		if errors.Is(err, couch.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("fetching equipment for delete: %w", err)
	}

	if err := r.db.DeleteDoc(ctx, id, current.Rev); err != nil {
		if errors.Is(err, couch.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("deleting equipment: %w", err)
	}
	return nil
}

func validateInput(input *Input) error {
	if input.Name == "" {
		return &ValidationError{Message: "Le champ 'name' est requis."}
	}
	if !validCategories[input.Category] {
		return ErrInvalidCategory
	}
	if input.Status != "" && !validStatuses[input.Status] {
		return ErrInvalidStatus
	}
	if input.Status == "" {
		input.Status = "ok"
	}
	return nil
}

type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }
