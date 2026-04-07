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
	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		title TEXT,
		aura TEXT,
		created_at DATETIME,
		luminosity REAL,
		cipher_path TEXT,
		preview_hint TEXT,
		require_passphrase BOOLEAN,
		layer_count INTEGER DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS memory_layers (
		id TEXT PRIMARY KEY,
		parent_id TEXT,
		cipher_path TEXT,
		created_at DATETIME,
		FOREIGN KEY(parent_id) REFERENCES memories(id)
	);`
	
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveMemory(m *vault.MemoryFragment) error {
	query := `INSERT OR REPLACE INTO memories (
		id, title, aura, created_at, luminosity,
		cipher_path, preview_hint, require_passphrase, layer_count
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := s.db.Exec(query,
		m.ID, m.Title, string(m.Aura), m.CreatedAt, m.Luminosity,
		m.CipherPath, m.PreviewHint, m.RequirePassphrase, m.LayerCount,
	)
	return err
}

func (s *Store) SaveLayer(l *vault.Layer) error {
	query := `INSERT OR REPLACE INTO memory_layers (id, parent_id, cipher_path, created_at) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, l.ID, l.ParentID, l.CipherPath, l.CreatedAt)
	if err != nil {
		return err
	}
	// MemoryのLayerCountをインクリメント
	_, err = s.db.Exec("UPDATE memories SET layer_count = layer_count + 1 WHERE id = ?", l.ParentID)
	return err
}

func (s *Store) ListLayers(parentID string) ([]*vault.Layer, error) {
	rows, err := s.db.Query("SELECT id, parent_id, cipher_path, created_at FROM memory_layers WHERE parent_id = ? ORDER BY created_at ASC", parentID)
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

func (s *Store) ListMemories() ([]*vault.MemoryFragment, error) {
	rows, err := s.db.Query("SELECT * FROM memories ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var res []*vault.MemoryFragment
	for rows.Next() {
		m := &vault.MemoryFragment{}
		var auraStr string
		err := rows.Scan(
			&m.ID, &m.Title, &auraStr, &m.CreatedAt, &m.Luminosity,
			&m.CipherPath, &m.PreviewHint, &m.RequirePassphrase, &m.LayerCount,
		)
		if err != nil {
			return nil, err
		}
		m.Aura = vault.ReminiscenceState(auraStr)
		res = append(res, m)
	}
	return res, nil
}
