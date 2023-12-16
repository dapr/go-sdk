package state

import (
	"context"
	"testing"
	"time"

	"github.com/dapr/go-sdk/actor/mock"
	"github.com/golang/mock/gomock"
)

func TestStateManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock.NewMockStateManager(ctrl)

	t.Run("Should Add without error", func(t *testing.T) {
		sm.EXPECT().Add("stateName", "value").Return(nil)
		sm.Add("stateName", "value")
	})

	t.Run("Should Remove without error", func(t *testing.T) {
		sm.EXPECT().Remove("stateName").Return(nil)
		sm.Remove("stateName")
	})

	t.Run("Should Get without error", func(t *testing.T) {
		sm.EXPECT().Get("test", "test").Return(nil)
		sm.Get("test", "test")
	})

	t.Run("Should Set without error", func(t *testing.T) {
		sm.EXPECT().Set("test", "test").Return(nil)
		sm.Set("test", "test")
	})

	t.Run("Should Contains without error", func(t *testing.T) {
		sm.EXPECT().Contains("test").Return(true, nil)
		sm.Contains("test")
	})
	t.Run("Should Save without error", func(t *testing.T) {
		sm.EXPECT().Save().Return(nil)
		sm.Save()
	})
	t.Run("Should Flush without error", func(t *testing.T) {
		sm.EXPECT().Flush()
		sm.Flush()
	})
}

func TestMockStateManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock.NewMockStateManagerContext(ctrl)

	ctx := context.Background()
	t.Run("Should Add without error", func(t *testing.T) {
		sm.EXPECT().Add(ctx, "stateName", "value").Return(nil)
		sm.Add(ctx, "stateName", "value")
	})

	t.Run("Should Remove without error", func(t *testing.T) {
		sm.EXPECT().Remove(ctx, "stateName").Return(nil)
		sm.Remove(ctx, "stateName")
	})

	t.Run("Should Get without error", func(t *testing.T) {
		sm.EXPECT().Get(ctx, "test", "test").Return(nil)
		sm.Get(ctx, "test", "test")
	})

	t.Run("Should Set without error", func(t *testing.T) {
		sm.EXPECT().Set(ctx, "test", "test").Return(nil)
		sm.Set(ctx, "test", "test")
	})
	t.Run("Should SetWithTTL without error", func(t *testing.T) {
		sm.EXPECT().SetWithTTL(ctx, "test", "test", time.Second).Return(nil)
		sm.SetWithTTL(ctx, "test", "test", time.Second)
	})
	t.Run("Should Contains without error", func(t *testing.T) {
		sm.EXPECT().Contains(ctx, "test").Return(true, nil)
		sm.Contains(ctx, "test")
	})
	t.Run("Should Save without error", func(t *testing.T) {
		sm.EXPECT().Save(ctx).Return(nil)
		sm.Save(ctx)
	})
	t.Run("Should Flush without error", func(t *testing.T) {
		sm.EXPECT().Flush(ctx)
		sm.Flush(ctx)
	})
}
