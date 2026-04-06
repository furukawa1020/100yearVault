package db

import (
	"database/sql"
	"vaultapp/internal/vault"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	
	schema := `
	CREATE TABLE IF NOT EXISTS vaults (
		id TEXT PRIMARY KEY,
		title TEXT,
		state TEXT,
		created_at DATETIME,
		sealed_at DATETIME,
		unlock_at DATETIME,
		opened_at DATETIME,
		deleted_at DATETIME,
		content_type TEXT,
		cipher_path TEXT,
		preview_hint TEXT,
		require_passphrase BOOLEAN,
		passphrase_hash TEXT,
		allow_reopen BOOLEAN,
		destroy_on_open BOOLEAN,
		layer_count INTEGER DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS vault_layers (
		id TEXT PRIMARY KEY,
		parent_id TEXT,
		cipher_path TEXT,
		created_at DATETIME,
		FOREIGN KEY(parent_id) REFERENCES vaults(id)
	);`
	
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveVault(v *vault.Vault) error {
	query := `INSERT OR REPLACE INTO vaults (
		id, title, state, created_at, sealed_at, unlock_at, opened_at, deleted_at,
		content_type, cipher_path, preview_hint, require_passphrase, passphrase_hash,
		allow_reopen, destroy_on_open, layer_count
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := s.db.Exec(query,
		v.ID, v.Title, string(v.State), v.CreatedAt, v.SealedAt, v.UnlockAt, v.OpenedAt, v.DeletedAt,
		v.ContentType, v.CipherPath, v.PreviewHint, v.RequirePassphrase, v.PassphraseHash,
		v.AllowReopen, v.DestroyOnOpen, v.LayerCount,
	)
	return err
}

func (s *Store) SaveLayer(l *vault.Layer) error {
	query := `INSERT OR REPLACE INTO vault_layers (id, parent_id, cipher_path, created_at) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, l.ID, l.ParentID, l.CipherPath, l.CreatedAt)
	if err != nil {
		return err
	}
	// VaultのLayerCountをインクリメント
	_, err = s.db.Exec("UPDATE vaults SET layer_count = layer_count + 1 WHERE id = ?", l.ParentID)
	return err
}

func (s *Store) ListLayers(parentID string) ([]*vault.Layer, error) {
	rows, err := s.db.Query("SELECT id, parent_id, cipher_path, created_at FROM vault_layers WHERE parent_id = ? ORDER BY created_at ASC", parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []*vault.Layer
	for rows.Next() {
		l := &vault.Layer{}
		if err := rows.Scan(&l.ID, &l.ParentID, &l.CipherPath, &l.CreatedAt); err != nil {
			return nil, err
		}
		layers = append(layers, l)
	}
	return layers, nil
}

func (s *Store) ListVaults() ([]*vault.Vault, error) {
	rows, err := s.db.Query("SELECT * FROM vaults WHERE state != 'Destroyed' ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var res []*vault.Vault
	for rows.Next() {
		v := &vault.Vault{}
		var stateStr string
		err := rows.Scan(
			&v.ID, &v.Title, &stateStr, &v.CreatedAt, &v.SealedAt, &v.UnlockAt, &v.OpenedAt, &v.DeletedAt,
			&v.ContentType, &v.CipherPath, &v.PreviewHint, &v.RequirePassphrase, &v.PassphraseHash,
			&v.AllowReopen, &v.DestroyOnOpen, &v.LayerCount,
		)
		if err != nil {
			return nil, err
		}
		v.State = vault.State(stateStr)
		res = append(res, v)
	}
	return res, nil
}
