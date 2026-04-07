package vault

import (
	"time"
)

// ReminiscenceState は記憶の「存在感」を表す
type ReminiscenceState string

const (
	StateEcho    ReminiscenceState = "Echo"    // 遠い残響
	StatePulse   ReminiscenceState = "Pulse"   // 鼓動
	StateRadiant ReminiscenceState = "Radiant" // 輝き
)

// MemoryFragment は100年分の宇宙を構成する一つの星（記憶）
type MemoryFragment struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	Aura              ReminiscenceState `json:"aura"`
	CreatedAt         time.Time         `json:"created_at"`
	Luminosity        float32           `json:"luminosity"` // 0.0 ~ 1.0 (時間の経過や共鳴で変化)
	
	// 技術的背面（Age暗号化は「時層」の保護として維持）
	CipherPath        string            `json:"cipher_path"`
	PreviewHint       string            `json:"preview_hint"`
	RequirePassphrase bool              `json:"require_passphrase"`
	
	// 2126年標準: 地層 (Layers)
	LayerCount int `json:"layer_count"`
}

type Layer struct {
	ID         string    `json:"id"`
	ParentID   string    `json:"parent_id"` // MemoryFragment.ID
	CipherPath string    `json:"cipher_path"`
	CreatedAt  time.Time `json:"created_at"`
}
