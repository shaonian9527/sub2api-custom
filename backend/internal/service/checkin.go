package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	DailyCheckinReward = 1.0
	DailyCheckinPrefix = "DAILY-CHECKIN-"
	DailyCheckinNote   = "daily checkin reward"
)

var (
	ErrAlreadyCheckedIn = infraerrors.Conflict("ALREADY_CHECKED_IN", "already checked in today")
	ErrCheckinDisabled  = infraerrors.Forbidden("CHECKIN_DISABLED", "daily checkin is disabled")
	shanghaiLocation    = func() *time.Location {
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			return time.FixedZone("CST", 8*3600)
		}
		return loc
	}()
)

type CheckinStatus struct {
	Enabled             bool     `json:"enabled"`
	CheckedIn           bool     `json:"checked_in"`
	RewardAmount        float64  `json:"reward_amount"`
	BaseRewardAmount    float64  `json:"base_reward_amount"`
	ConsecutiveDays     int      `json:"consecutive_days"`
	NextRewardAmount    float64  `json:"next_reward_amount"`
	Timezone            string   `json:"timezone"`
	Today               string   `json:"today"`
	CheckedInAt         *string  `json:"checked_in_at,omitempty"`
	BalanceAfter        *float64 `json:"balance_after,omitempty"`
	CheckinCode         *string  `json:"checkin_code,omitempty"`
	HistoryPagePath     string   `json:"history_page_path,omitempty"`
}

type CheckinResult struct {
	Message           string  `json:"message"`
	RewardAmount      float64 `json:"reward_amount"`
	BaseRewardAmount  float64 `json:"base_reward_amount"`
	BonusRewardAmount float64 `json:"bonus_reward_amount"`
	NewBalance        float64 `json:"new_balance"`
	Today             string  `json:"today"`
	Timezone          string  `json:"timezone"`
	ConsecutiveDays   int     `json:"consecutive_days"`
	CheckedInAt       string  `json:"checked_in_at"`
	Code              string  `json:"code"`
}

type CheckinHistoryItem struct {
	ID              int64   `json:"id"`
	Code            string  `json:"code"`
	RewardAmount    float64 `json:"reward_amount"`
	BaseReward      float64 `json:"base_reward_amount"`
	BonusReward     float64 `json:"bonus_reward_amount"`
	ConsecutiveDays int     `json:"consecutive_days"`
	Day             string  `json:"day"`
	UsedAt          string  `json:"used_at"`
}

func shanghaiNow() time.Time {
	return time.Now().In(shanghaiLocation)
}

func localDayString(t time.Time) string {
	return t.In(shanghaiLocation).Format("2006-01-02")
}

func buildDailyCheckinCode(userID int64, day string) string {
	return fmt.Sprintf("%s%d-%s", DailyCheckinPrefix, userID, day)
}

func consecutiveBonus(days int) float64 {
	switch {
	case days >= 30:
		return 3
	case days >= 14:
		return 2
	case days >= 7:
		return 1
	default:
		return 0
	}
}

func totalReward(base float64, days int) float64 {
	return base + consecutiveBonus(days)
}

func parseCheckinMeta(note string) (day string, consecutive int, baseReward float64, bonusReward float64) {
	baseReward = 0
	bonusReward = 0
	if note == "" {
		return
	}
	parts := strings.Split(note, "|")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "day":
			day = kv[1]
		case "streak":
			fmt.Sscanf(kv[1], "%d", &consecutive)
		case "base":
			fmt.Sscanf(kv[1], "%f", &baseReward)
		case "bonus":
			fmt.Sscanf(kv[1], "%f", &bonusReward)
		}
	}
	return
}

func buildCheckinNote(day string, streak int, baseReward, bonusReward float64) string {
	return fmt.Sprintf("%s|day=%s|streak=%d|base=%.2f|bonus=%.2f", DailyCheckinNote, day, streak, baseReward, bonusReward)
}

func (s *RedeemService) listCheckinHistory(ctx context.Context, userID int64, limit int) ([]RedeemCode, error) {
	items, err := s.redeemRepo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]RedeemCode, 0, len(items))
	for _, item := range items {
		if strings.HasPrefix(item.Code, DailyCheckinPrefix) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *RedeemService) getCurrentStreak(ctx context.Context, userID int64, today string) int {
	history, err := s.listCheckinHistory(ctx, userID, 60)
	if err != nil || len(history) == 0 {
		return 0
	}

	streak := 0
	expected := mustParseLocalDay(today)
	for _, item := range history {
		day, storedStreak, _, _ := parseCheckinMeta(item.Notes)
		if day == "" {
			continue
		}
		parsed := mustParseLocalDay(day)
		if parsed.Equal(expected) {
			if storedStreak > 0 {
				return storedStreak
			}
			streak++
			expected = expected.AddDate(0, 0, -1)
			continue
		}
		if parsed.Equal(expected.AddDate(0, 0, -1)) {
			streak++
			expected = expected.AddDate(0, 0, -1)
			continue
		}
		break
	}
	return streak
}

func mustParseLocalDay(day string) time.Time {
	t, err := time.ParseInLocation("2006-01-02", day, shanghaiLocation)
	if err != nil {
		return shanghaiNow()
	}
	return t
}

func (s *RedeemService) GetCheckinStatus(ctx context.Context, userID int64) (*CheckinStatus, error) {
	today := localDayString(shanghaiNow())
	code := buildDailyCheckinCode(userID, today)
	enabled := true
	baseReward := DailyCheckinReward
	if s.settingService != nil {
		enabled = s.settingService.IsDailyCheckinEnabled(ctx)
		baseReward = s.settingService.GetDailyCheckinReward(ctx)
	}

	currentStreak := s.getCurrentStreak(ctx, userID, today)
	nextStreak := currentStreak
	if nextStreak == 0 {
		nextStreak = 1
	}
	nextReward := totalReward(baseReward, nextStreak)

	redeemCode, err := s.redeemRepo.GetByCode(ctx, code)
	if err != nil {
		if err == ErrRedeemCodeNotFound {
			return &CheckinStatus{
				Enabled:          enabled,
				CheckedIn:        false,
				RewardAmount:     nextReward,
				BaseRewardAmount: baseReward,
				ConsecutiveDays:  currentStreak,
				NextRewardAmount: nextReward,
				Timezone:         "Asia/Shanghai",
				Today:            today,
				HistoryPagePath:  "/checkin-history",
			}, nil
		}
		return nil, fmt.Errorf("get checkin status: %w", err)
	}

	day, streak, base, bonus := parseCheckinMeta(redeemCode.Notes)
	if day == "" {
		day = today
	}
	if streak == 0 {
		streak = currentStreak
	}
	if base == 0 && redeemCode.Value >= 0 {
		base = baseReward
	}
	status := &CheckinStatus{
		Enabled:          enabled,
		CheckedIn:        redeemCode.Status == StatusUsed,
		RewardAmount:     redeemCode.Value,
		BaseRewardAmount: base,
		ConsecutiveDays:  streak,
		NextRewardAmount: totalReward(baseReward, streak),
		Timezone:         "Asia/Shanghai",
		Today:            day,
		HistoryPagePath:  "/checkin-history",
	}
	if bonus > 0 && status.RewardAmount == 0 {
		status.RewardAmount = base + bonus
	}
	if redeemCode.UsedAt != nil {
		ts := redeemCode.UsedAt.In(shanghaiLocation).Format(time.RFC3339)
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
	enabled := true
	baseReward := DailyCheckinReward
	if s.settingService != nil {
		enabled = s.settingService.IsDailyCheckinEnabled(ctx)
		baseReward = s.settingService.GetDailyCheckinReward(ctx)
	}
	if !enabled {
		return nil, ErrCheckinDisabled
	}

	today := localDayString(shanghaiNow())
	code := buildDailyCheckinCode(userID, today)

	if existing, err := s.redeemRepo.GetByCode(ctx, code); err == nil {
		if existing.Status == StatusUsed {
			return nil, ErrAlreadyCheckedIn
		}
	} else if err != ErrRedeemCodeNotFound {
		return nil, fmt.Errorf("precheck checkin code: %w", err)
	}

	previousStreak := 0
	yesterday := mustParseLocalDay(today).AddDate(0, 0, -1).Format("2006-01-02")
	yesterdayCode := buildDailyCheckinCode(userID, yesterday)
	if y, err := s.redeemRepo.GetByCode(ctx, yesterdayCode); err == nil && y.Status == StatusUsed {
		_, storedStreak, _, _ := parseCheckinMeta(y.Notes)
		if storedStreak > 0 {
			previousStreak = storedStreak
		} else {
			previousStreak = 1
		}
	}
	streak := previousStreak + 1
	bonusReward := consecutiveBonus(streak)
	rewardAmount := baseReward + bonusReward

	createErr := s.redeemRepo.Create(ctx, &RedeemCode{
		Code:   code,
		Type:   RedeemTypeBalance,
		Value:  rewardAmount,
		Status: StatusUnused,
		Notes:  buildCheckinNote(today, streak, baseReward, bonusReward),
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

	checkedInAt := shanghaiNow().Format(time.RFC3339)
	if redeemCode != nil && redeemCode.UsedAt != nil {
		checkedInAt = redeemCode.UsedAt.In(shanghaiLocation).Format(time.RFC3339)
	}

	return &CheckinResult{
		Message:           "checkin successful",
		RewardAmount:      rewardAmount,
		BaseRewardAmount:  baseReward,
		BonusRewardAmount: bonusReward,
		NewBalance:        user.Balance,
		Today:             today,
		Timezone:          "Asia/Shanghai",
		ConsecutiveDays:   streak,
		CheckedInAt:       checkedInAt,
		Code:              code,
	}, nil
}

func (s *RedeemService) GetCheckinHistory(ctx context.Context, userID int64, page, pageSize int) ([]CheckinHistoryItem, *pagination.PaginationResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	params := pagination.NewPaginationParams(page, pageSize)
	items, result, err := s.redeemRepo.ListByUserPaginated(ctx, userID, params, RedeemTypeBalance)
	if err != nil {
		return nil, nil, fmt.Errorf("get checkin history: %w", err)
	}
	filtered := make([]CheckinHistoryItem, 0, len(items))
	for _, item := range items {
		if !strings.HasPrefix(item.Code, DailyCheckinPrefix) {
			continue
		}
		day, streak, baseReward, bonusReward := parseCheckinMeta(item.Notes)
		usedAt := ""
		if item.UsedAt != nil {
			usedAt = item.UsedAt.In(shanghaiLocation).Format(time.RFC3339)
		}
		filtered = append(filtered, CheckinHistoryItem{
			ID:              item.ID,
			Code:            item.Code,
			RewardAmount:    item.Value,
			BaseReward:      baseReward,
			BonusReward:     bonusReward,
			ConsecutiveDays: streak,
			Day:             day,
			UsedAt:          usedAt,
		})
	}
	return filtered, result, nil
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
