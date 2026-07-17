package workspace

import (
	"database/sql"
	"errors"
)

// Repository interface defines database operations
type Repository interface {
	Create(workspace *Workspace) error
	GetByID(id string) (*Workspace, error)
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new instance of the workspace repository
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(w *Workspace) error {
	query := `INSERT INTO workspaces (id, name, description, owner_id, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(query, w.ID, w.Name, w.Description, w.OwnerID, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *repository) GetByID(id string) (*Workspace, error) {
	w := &Workspace{}
	query := `SELECT id, name, description, owner_id, created_at, updated_at FROM workspaces WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&w.ID, &w.Name, &w.Description, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Or a custom ErrNotFound
		}
		return nil, err
	}
	return w, nil
}
