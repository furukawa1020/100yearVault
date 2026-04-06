package vault

import (
	"time"
)

type State string

const (
	StateDraft      State = "Draft"
	StateSealed     State = "Sealed"
	StateUnlockable State = "Unlockable"
	StateOpened     State = "Opened"
	StateDestroyed  State = "Destroyed"
)

type Vault struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	State             State     `json:"state"`
	CreatedAt         time.Time `json:"created_at"`
	SealedAt          time.Time `json:"sealed_at"`
	UnlockAt          time.Time `json:"unlock_at"`
	OpenedAt          time.Time `json:"opened_at"`
	DeletedAt         time.Time `json:"deleted_at"`
	ContentType       string    `json:"content_type"`
	CipherPath        string    `json:"cipher_path"` // 最初の層のパス (廃止検討、すべて Layer に統一も可)
	PreviewHint       string    `json:"preview_hint"`
	RequirePassphrase bool      `json:"require_passphrase"`
	PassphraseHash    string    `json:"passphrase_hash"`
	AllowReopen       bool      `json:"allow_reopen"`
	DestroyOnOpen     bool      `json:"destroy_on_open"`

	// 2126年標準: STP 用拡張
	LayerCount int `json:"layer_count"`
}

type Layer struct {
	ID         string    `json:"id"`
	ParentID   string    `json:"parent_id"` // Vault.ID
	CipherPath string    `json:"cipher_path"`
	CreatedAt  time.Time `json:"created_at"`
}
