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
	CipherPath        string    `json:"cipher_path"`
	PreviewHint       string    `json:"preview_hint"`
	RequirePassphrase bool      `json:"require_passphrase"`
	PassphraseHash    string    `json:"passphrase_hash"`
	AllowReopen       bool      `json:"allow_reopen"`
	DestroyOnOpen     bool      `json:"destroy_on_open"`
}
