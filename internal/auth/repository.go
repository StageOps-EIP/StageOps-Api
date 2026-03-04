package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// UserRepository defines the persistence contract for users.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
}

// CouchConfig holds CouchDB connection settings read from environment.
type CouchConfig struct {
	BaseURL  string
	DB       string
	Username string
	Password string
}

// CouchDBRepository implements UserRepository via the CouchDB HTTP API.
type CouchDBRepository struct {
	cfg    CouchConfig
	client *http.Client
}

// NewCouchDBRepository creates a ready-to-use CouchDB repository.
func NewCouchDBRepository(cfg CouchConfig) *CouchDBRepository {
	return &CouchDBRepository{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// couchViewResponse models the JSON envelope returned by a CouchDB view.
type couchViewResponse struct {
	Rows []struct {
		Doc User `json:"doc"`
	} `json:"rows"`
}

// FindByEmail queries the existing _design/users/_view/by_email view.
// Returns ErrUserNotFound when no document matches.
func (r *CouchDBRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	params := url.Values{}
	// CouchDB view key must be a JSON-encoded value: "email@example.com"
	params.Set("key", fmt.Sprintf("%q", email))
	params.Set("include_docs", "true")

	rawURL := fmt.Sprintf(
		"%s/%s/_design/users/_view/by_email?%s",
		r.cfg.BaseURL, r.cfg.DB, params.Encode(),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.SetBasicAuth(r.cfg.Username, r.cfg.Password)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying CouchDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CouchDB returned status %d", resp.StatusCode)
	}

	var viewResp couchViewResponse
	if err := json.NewDecoder(resp.Body).Decode(&viewResp); err != nil {
		return nil, fmt.Errorf("decoding view response: %w", err)
	}

	if len(viewResp.Rows) == 0 {
		return nil, ErrUserNotFound
	}

	user := viewResp.Rows[0].Doc
	return &user, nil
}

// FindByID fetches a user document by its CouchDB _id.
// Returns ErrUserNotFound on a 404 response.
func (r *CouchDBRepository) FindByID(ctx context.Context, id string) (*User, error) {
	rawURL := fmt.Sprintf("%s/%s/%s", r.cfg.BaseURL, r.cfg.DB, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.SetBasicAuth(r.cfg.Username, r.cfg.Password)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying CouchDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrUserNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CouchDB returned status %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decoding user document: %w", err)
	}

	return &user, nil
}

// UpdateUser writes an existing user document back to CouchDB.
// The user must have a valid _rev field (returned by FindByEmail/FindByID)
// so CouchDB can detect write conflicts.
func (r *CouchDBRepository) UpdateUser(ctx context.Context, user *User) error {
	body, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("encoding user document: %w", err)
	}

	rawURL := fmt.Sprintf("%s/%s/%s", r.cfg.BaseURL, r.cfg.DB, user.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, rawURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.SetBasicAuth(r.cfg.Username, r.cfg.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("updating document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("document conflict on update")
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CouchDB returned status %d", resp.StatusCode)
	}

	return nil
}

// Create stores a new user document using PUT /<db>/<id>.
// A 409 Conflict from CouchDB is wrapped as ErrEmailAlreadyExists to
// handle the edge case of a concurrent duplicate registration.
func (r *CouchDBRepository) Create(ctx context.Context, user *User) error {
	body, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("encoding user document: %w", err)
	}

	rawURL := fmt.Sprintf("%s/%s/%s", r.cfg.BaseURL, r.cfg.DB, user.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, rawURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.SetBasicAuth(r.cfg.Username, r.cfg.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("creating document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("document conflict: %w", ErrEmailAlreadyExists)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("CouchDB returned status %d", resp.StatusCode)
	}

	return nil
}
