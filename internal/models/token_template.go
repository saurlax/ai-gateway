package models

import (
	"encoding/json"

	"gorm.io/datatypes"
)

type TokenTemplate struct {
	ID                uint                      `gorm:"primarykey" json:"id"`
	Name              string                    `gorm:"size:64;not null" json:"name"`
	Models            string                    `gorm:"type:text" json:"models"`
	ExpiryDays        int                       `gorm:"not null;default:-1" json:"expiry_days"`
	Status            int                       `gorm:"not null;default:1" json:"status"`
	BYOKOnly          bool                      `gorm:"not null;default:false" json:"byok_only"`
	AllowedChannelIDs datatypes.JSONSlice[uint] `gorm:"type:text" json:"allowed_channel_ids"`
	AllowedGroupIDs   datatypes.JSONSlice[uint] `gorm:"type:text" json:"allowed_group_ids"`
	CreatedAt         int64                     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         int64                     `gorm:"autoUpdateTime" json:"updated_at"`
}

// AllowsGroup 报告该模版是否对指定用户组开放。
// 空白名单语义：未配置任何组 = 对所有组开放（向后兼容）。
func (t *TokenTemplate) AllowsGroup(groupID uint) bool {
	if len(t.AllowedGroupIDs) == 0 {
		return true
	}
	for _, id := range t.AllowedGroupIDs {
		if id == groupID {
			return true
		}
	}
	return false
}

// SyncFields 标记一次模版→令牌同步要覆盖哪些字段。
type SyncFields struct {
	Models   bool
	Channels bool
	BYOKOnly bool
}

// TokenFieldsEqualForFields 只比较 f 选中的字段；未选字段视为相等（不参与差异判定）。
func TokenFieldsEqualForFields(tpl *TokenTemplate, tok *Token, f SyncFields) bool {
	if f.Models && !modelsEqual(tpl.Models, tok.Models) {
		return false
	}
	if f.Channels && !channelsEqual([]uint(tpl.AllowedChannelIDs), []uint(tok.AllowedChannelIDs)) {
		return false
	}
	if f.BYOKOnly && tpl.BYOKOnly != tok.BYOKOnly {
		return false
	}
	return true
}

func TokenFieldsEqual(tplModelsJSON string, tplChannelIDs []uint, tok *Token) bool {
	if !modelsEqual(tplModelsJSON, tok.Models) {
		return false
	}
	return channelsEqual(tplChannelIDs, []uint(tok.AllowedChannelIDs))
}

func modelsEqual(a, b string) bool {
	aSet, aOK := parseModelSet(a)
	bSet, bOK := parseModelSet(b)
	if !aOK || !bOK {
		return a == b
	}
	if len(aSet) != len(bSet) {
		return false
	}
	for k := range aSet {
		if _, ok := bSet[k]; !ok {
			return false
		}
	}
	return true
}

func parseModelSet(s string) (map[string]struct{}, bool) {
	if s == "" {
		return map[string]struct{}{}, true
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil, false
	}
	set := make(map[string]struct{}, len(arr))
	for _, m := range arr {
		set[m] = struct{}{}
	}
	return set, true
}

func channelsEqual(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	aSet := make(map[uint]struct{}, len(a))
	for _, id := range a {
		aSet[id] = struct{}{}
	}
	for _, id := range b {
		if _, ok := aSet[id]; !ok {
			return false
		}
	}
	return true
}
