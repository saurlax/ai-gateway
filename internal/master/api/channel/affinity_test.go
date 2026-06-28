package channel

import (
	"strconv"
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/models"
	"github.com/VaalaCat/ai-gateway/internal/pkg/eventbus"
	"github.com/stretchr/testify/require"
)

func ptrInt(i int) *int { return &i }

// TestCreate_AffinityValid verifies that a valid Affinity override is persisted on create.
func TestCreate_AffinityValid(t *testing.T) {
	db := setupTestDB(t)
	ctx := newTestContext(t, db, "")
	ctx.App.SetEventBus(eventbus.NewMemoryBus())
	h := &Handler{}

	enabled := true
	res, err := h.Create(ctx, CreateRequest{
		Name: "aff-ch",
		Affinity: &models.ChannelAffinity{
			Enabled: &enabled,
			TTLSec:  ptrInt(600),
		},
	})
	require.NoError(t, err)
	got := res.Value.Affinity.Data()
	require.NotNil(t, got.Enabled)
	require.True(t, *got.Enabled)
	require.NotNil(t, got.TTLSec)
	require.Equal(t, 600, *got.TTLSec)
}

// TestCreate_AffinityInvalidTTL verifies that a negative ttl_sec is rejected with BadRequest.
func TestCreate_AffinityInvalidTTL(t *testing.T) {
	db := setupTestDB(t)
	ctx := newTestContext(t, db, "")
	ctx.App.SetEventBus(eventbus.NewMemoryBus())
	h := &Handler{}

	_, err := h.Create(ctx, CreateRequest{
		Name:     "bad-aff-ch",
		Affinity: &models.ChannelAffinity{TTLSec: ptrInt(-1)},
	})
	require.Error(t, err)
}

// TestUpdate_AffinityValid verifies that a valid affinity in the PATCH body is round-tripped and persisted.
func TestUpdate_AffinityValid(t *testing.T) {
	db := setupTestDB(t)
	ctx := newTestContext(t, db, "")
	ctx.App.SetEventBus(eventbus.NewMemoryBus())
	h := &Handler{}

	ch := models.Channel{ChannelCore: models.ChannelCore{Name: "ch", Type: 1, Status: 1, Weight: 1}}
	require.NoError(t, db.Create(&ch).Error)

	req := UpdateRequest{ID: strconv.Itoa(int(ch.ID))}
	req.SetBodyMap(map[string]any{
		"affinity": map[string]any{
			"enabled": true,
			"ttl_sec": float64(300),
		},
	})
	got, err := h.Update(ctx, req)
	require.NoError(t, err)
	aff := got.Affinity.Data()
	require.NotNil(t, aff.Enabled)
	require.True(t, *aff.Enabled)
	require.NotNil(t, aff.TTLSec)
	require.Equal(t, 300, *aff.TTLSec)
}

// TestUpdate_AffinityInvalidTTL verifies that an out-of-range ttl_sec in PATCH is rejected.
func TestUpdate_AffinityInvalidTTL(t *testing.T) {
	db := setupTestDB(t)
	ctx := newTestContext(t, db, "")
	ctx.App.SetEventBus(eventbus.NewMemoryBus())
	h := &Handler{}

	ch := models.Channel{ChannelCore: models.ChannelCore{Name: "ch2", Type: 1, Status: 1, Weight: 1}}
	require.NoError(t, db.Create(&ch).Error)

	req := UpdateRequest{ID: strconv.Itoa(int(ch.ID))}
	req.SetBodyMap(map[string]any{
		"affinity": map[string]any{
			"ttl_sec": float64(99999),
		},
	})
	_, err := h.Update(ctx, req)
	require.Error(t, err)
}
