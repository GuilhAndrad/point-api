package service

import (
	"context"
	"errors"
	"time"

	"github.com/GuilhAndrad/point-api/internal/model"
	"github.com/GuilhAndrad/point-api/internal/repository"
)

type PunchService interface {
    Punch(ctx context.Context, userID int64) (*model.TimeRecord, error)
    GetMyRecords(ctx context.Context, userID int64, from, to time.Time) ([]model.TimeRecord, error)
    GetDailyReport(ctx context.Context, userID int64, from, to time.Time) ([]model.DailyReport, error)
    GetAllRecords(ctx context.Context, from, to time.Time) ([]model.TimeRecord, error)
    GetAdminReport(ctx context.Context, from, to time.Time) (map[int64][]model.DailyReport, error)
}

type punchService struct {
    repo repository.RecordRepository
}

func NewPunchService(repo repository.RecordRepository) PunchService {
    return &punchService{repo: repo}
}

// Punch registra entrada ou saída automaticamente,
// alternando com base no último registro do funcionário.
func (s *punchService) Punch(ctx context.Context, userID int64) (*model.TimeRecord, error) {
    nextType := model.RecordEntry

    last, err := s.repo.FindLastByUserID(ctx, userID)
    if err == nil && last.Type == model.RecordEntry {
        // Último foi entrada → próximo é saída
        nextType = model.RecordExit
    }

    record := &model.TimeRecord{
        UserID:    userID,
        Type:      nextType,
        Timestamp: time.Now(),
    }

    if err := s.repo.Create(ctx, record); err != nil {
        return nil, err
    }
    return record, nil
}

func (s *punchService) GetMyRecords(ctx context.Context, userID int64, from, to time.Time) ([]model.TimeRecord, error) {
    return s.repo.FindByUserID(ctx, userID, from, to)
}

// GetDailyReport agrupa os registros por dia e calcula horas trabalhadas.
// Regra: pares entrada/saída somam as horas. Entrada sem saída = dia aberto.
func (s *punchService) GetDailyReport(ctx context.Context, userID int64, from, to time.Time) ([]model.DailyReport, error) {
    records, err := s.repo.FindByUserID(ctx, userID, from, to)
    if err != nil {
        return nil, err
    }

    byDay := map[string][]model.TimeRecord{}
    for _, r := range records {
        day := r.Timestamp.Format("2006-01-02")
        byDay[day] = append(byDay[day], r)
    }

    var report []model.DailyReport
    for day, dayRecords := range byDay {
        hours := calculateHours(dayRecords)
        report = append(report, model.DailyReport{
            Date:        day,
            Records:     dayRecords,
            HoursWorked: hours,
        })
    }
    return report, nil
}

func (s *punchService) GetAllRecords(ctx context.Context, from, to time.Time) ([]model.TimeRecord, error) {
    return s.repo.FindAll(ctx, from, to)
}

func (s *punchService) GetAdminReport(ctx context.Context, from, to time.Time) (map[int64][]model.DailyReport, error) {
    records, err := s.repo.FindAll(ctx, from, to)
    if err != nil {
        return nil, err
    }

    // userID → dia → registros
    userMap := map[int64]map[string][]model.TimeRecord{}

    for _, r := range records {
        if _, ok := userMap[r.UserID]; !ok {
            userMap[r.UserID] = map[string][]model.TimeRecord{}
        }

        day := r.Timestamp.Format("2006-01-02")
        userMap[r.UserID][day] = append(userMap[r.UserID][day], r)
    }

    result := map[int64][]model.DailyReport{}

    for userID, days := range userMap {
        for day, recs := range days {
            result[userID] = append(result[userID], model.DailyReport{
                Date:        day,
                Records:     recs,
                HoursWorked: calculateHours(recs),
            })
        }
    }

    return result, nil
}

// calculateHours soma pares entrada/saída. Lógica isolada e fácil de testar.
func calculateHours(records []model.TimeRecord) float64 {
    var total float64
    var entryTime time.Time
    var hasEntry bool

    for _, r := range records {
        if r.Type == model.RecordEntry {
            entryTime = r.Timestamp
            hasEntry = true
        } else if r.Type == model.RecordExit && hasEntry {
            total += r.Timestamp.Sub(entryTime).Hours()
            hasEntry = false
        }
    }
    return total
}

var ErrNotEmployee = errors.New("apenas funcionários podem registrar ponto")