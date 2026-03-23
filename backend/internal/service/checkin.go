package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	DailyCheckinReward = 1.0
	DailyCheckinPrefix = "DAILY-CHECKIN-"
	DailyCheckinNote   = "daily checkin reward"
)

var ErrAlreadyCheckedIn = infraerrors.Conflict("ALREADY_CHECKED_IN", "already checked in today")

type CheckinStatus struct {
	CheckedIn      bool     `json:"checked_in"`
	RewardAmount   float64  `json:"reward_amount"`
	Today          string   `json:"today"`
	CheckedInAt    *string  `json:"checked_in_at,omitempty"`
	BalanceAfter   *float64 `json:"balance_after,omitempty"`
	CheckinCode    *string  `json:"checkin_code,omitempty"`
}

type CheckinResult struct {
	Message      string  `json:"message"`
	RewardAmount float64 `json:"reward_amount"`
	NewBalance   float64 `json:"new_balance"`
	Today        string  `json:"today"`
	CheckedInAt  string  `json:"checked_in_at"`
	Code         string  `json:"code"`
}

func utcDayString(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func buildDailyCheckinCode(userID int64, day string) string {
	return fmt.Sprintf("%s%d-%s", DailyCheckinPrefix, userID, day)
}

func (s *RedeemService) GetCheckinStatus(ctx context.Context, userID int64) (*CheckinStatus, error) {
	day := utcDayString(time.Now())
	code := buildDailyCheckinCode(userID, day)
	rewardAmount := DailyCheckinReward
	if s.settingService != nil {
		rewardAmount = s.settingService.GetDailyCheckinReward(ctx)
	}

	redeemCode, err := s.redeemRepo.GetByCode(ctx, code)
	if err != nil {
		if err == ErrRedeemCodeNotFound {
			return &CheckinStatus{
				CheckedIn:    false,
				RewardAmount: rewardAmount,
				Today:        day,
			}, nil
		}
		return nil, fmt.Errorf("get checkin status: %w", err)
	}

	status := &CheckinStatus{
		CheckedIn:    redeemCode.Status == StatusUsed,
		RewardAmount: rewardAmount,
		Today:        day,
	}
	if redeemCode.UsedAt != nil {
		ts := redeemCode.UsedAt.UTC().Format(time.RFC3339)
		status.CheckedInAt = &ts
	}
	if redeemCode.Code != "" {
		status.CheckinCode = &redeemCode.Code
	}
	if status.CheckedIn {
		user, userErr := s.userRepo.GetByID(ctx, userID)
		if userErr == nil {
			balance := user.Balance
			status.BalanceAfter = &balance
		}
	}
	return status, nil
}

func (s *RedeemService) Checkin(ctx context.Context, userID int64) (*CheckinResult, error) {
	day := utcDayString(time.Now())
	code := buildDailyCheckinCode(userID, day)
	rewardAmount := DailyCheckinReward
	if s.settingService != nil {
		rewardAmount = s.settingService.GetDailyCheckinReward(ctx)
	}

	if existing, err := s.redeemRepo.GetByCode(ctx, code); err == nil {
		if existing.Status == StatusUsed {
			return nil, ErrAlreadyCheckedIn
		}
	} else if err != ErrRedeemCodeNotFound {
		return nil, fmt.Errorf("precheck checkin code: %w", err)
	}

	createErr := s.redeemRepo.Create(ctx, &RedeemCode{
		Code:   code,
		Type:   RedeemTypeBalance,
		Value:  rewardAmount,
		Status: StatusUnused,
		Notes:  DailyCheckinNote,
	})
	if createErr != nil && !isDuplicateCreateErr(createErr) {
		return nil, fmt.Errorf("create checkin code: %w", createErr)
	}

	redeemCode, err := s.Redeem(ctx, userID, code)
	if err != nil {
		if err == ErrRedeemCodeUsed {
			return nil, ErrAlreadyCheckedIn
		}
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user after checkin: %w", err)
	}

	checkedInAt := time.Now().UTC().Format(time.RFC3339)
	if redeemCode != nil && redeemCode.UsedAt != nil {
		checkedInAt = redeemCode.UsedAt.UTC().Format(time.RFC3339)
	}

	return &CheckinResult{
		Message:      "checkin successful",
		RewardAmount: rewardAmount,
		NewBalance:   user.Balance,
		Today:        day,
		CheckedInAt:  checkedInAt,
		Code:         code,
	}, nil
}

func isDuplicateCreateErr(err error) bool {
	if err == nil {
		return false
	}
	if dbent.IsConstraintError(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
