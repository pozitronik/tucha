package service

import (
	"errors"
	"testing"

	"github.com/pozitronik/tucha/internal/domain/entity"
	"github.com/pozitronik/tucha/internal/testutil/mock"
)

func TestQuotaService_GetUsage_underQuota(t *testing.T) {
	svc := NewQuotaService(
		&mock.NodeRepositoryMock{
			TotalSizeFunc: func(userID int64) (int64, error) { return 500, nil },
		},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return &entity.User{ID: 1, QuotaBytes: 1000}, nil
			},
		},
	)

	usage, err := svc.GetUsage(1)
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if usage.Overquota {
		t.Error("Overquota = true, want false")
	}
	if usage.BytesUsed != 500 {
		t.Errorf("BytesUsed = %d, want 500", usage.BytesUsed)
	}
	if usage.BytesTotal != 1000 {
		t.Errorf("BytesTotal = %d, want 1000", usage.BytesTotal)
	}
}

func TestQuotaService_GetUsage_overQuota(t *testing.T) {
	svc := NewQuotaService(
		&mock.NodeRepositoryMock{
			TotalSizeFunc: func(userID int64) (int64, error) { return 1500, nil },
		},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return &entity.User{ID: 1, QuotaBytes: 1000}, nil
			},
		},
	)

	usage, err := svc.GetUsage(1)
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if !usage.Overquota {
		t.Error("Overquota = false, want true")
	}
}

func TestQuotaService_GetUsage_exactQuota(t *testing.T) {
	svc := NewQuotaService(
		&mock.NodeRepositoryMock{
			TotalSizeFunc: func(userID int64) (int64, error) { return 1000, nil },
		},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return &entity.User{ID: 1, QuotaBytes: 1000}, nil
			},
		},
	)

	usage, err := svc.GetUsage(1)
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	// used == quota is NOT over quota (uses > not >=).
	if usage.Overquota {
		t.Error("Overquota at exact quota = true, want false")
	}
}

func TestQuotaService_GetUsage_userNotFound(t *testing.T) {
	svc := NewQuotaService(
		&mock.NodeRepositoryMock{},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) { return nil, nil },
		},
	)

	_, err := svc.GetUsage(999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetUsage(unknown user) error = %v, want ErrNotFound", err)
	}
}

func TestQuotaService_CheckQuota(t *testing.T) {
	svc := NewQuotaService(
		&mock.NodeRepositoryMock{
			TotalSizeFunc: func(userID int64) (int64, error) { return 800, nil },
		},
		&mock.UserRepositoryMock{
			GetByIDFunc: func(id int64) (*entity.User, error) {
				return &entity.User{ID: 1, QuotaBytes: 1000}, nil
			},
		},
	)

	// 800 + 100 = 900 < 1000: not over.
	over, err := svc.CheckQuota(1, 100)
	if err != nil {
		t.Fatalf("CheckQuota: %v", err)
	}
	if over {
		t.Error("CheckQuota(100) = true, want false")
	}

	// 800 + 300 = 1100 > 1000: over.
	over, err = svc.CheckQuota(1, 300)
	if err != nil {
		t.Fatalf("CheckQuota: %v", err)
	}
	if !over {
		t.Error("CheckQuota(300) = false, want true")
	}
}
