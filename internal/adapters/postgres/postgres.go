package postgres

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/domain"
	"backend/pkg/log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdapterPostgres struct {
	pool *pgxpool.Pool
	log  log.Logger
}

func NewAdapterPostgres(ctx context.Context, logger log.Logger, dsn string) (*AdapterPostgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error(
			"Error connecting to database",
			log.FieldLogger{Key: "Error", Value: err},
		)

		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &AdapterPostgres{pool: pool, log: logger}, nil
}

func (a *AdapterPostgres) Create(ctx context.Context, cv *domain.CV) error {
	cv.UUID = uuid.New()

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			if !errors.Is(errRollback, pgx.ErrTxClosed) {
				a.log.Error(
					"Error rolling back transaction",
					log.FieldLogger{Key: "Error", Value: errRollback},
				)
			}
		}
	}()

	_, err = tx.Exec(ctx, CreateCVQuery,
		cv.UUID,
		cv.FirstName,
		cv.LastName,
		cv.CVTitle,
		cv.Specialization,
		cv.WorkExperience,
		cv.RawText,
	)
	if err != nil {
		return fmt.Errorf("insert cv: %w", err)
	}

	if err := a.upsertAndLinkSkills(ctx, tx, cv); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (a *AdapterPostgres) Update(ctx context.Context, cv *domain.CV) error {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			if !errors.Is(errRollback, pgx.ErrTxClosed) {
				a.log.Error(
					"Error rolling back transaction",
					log.FieldLogger{Key: "Error", Value: errRollback},
				)
			}
		}
	}()

	_, err = tx.Exec(ctx, UpdateCVQuery,
		cv.UUID,
		cv.FirstName,
		cv.LastName,
		cv.CVTitle,
		cv.Specialization,
		cv.WorkExperience,
		cv.RawText,
	)
	if err != nil {
		return fmt.Errorf("update cv: %w", err)
	}

	if _, err = tx.Exec(ctx, UnlinkAllSkillsFromCVQuery, cv.UUID); err != nil {
		return fmt.Errorf("unlink skills: %w", err)
	}

	if err := a.upsertAndLinkSkills(ctx, tx, cv); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (a *AdapterPostgres) GetByID(ctx context.Context, id uuid.UUID) (domain.CV, error) {
	var cv domain.CV

	row := a.pool.QueryRow(ctx, GetCVByIDQuery, id)
	err := row.Scan(
		&cv.UUID,
		&cv.FirstName,
		&cv.LastName,
		&cv.CVTitle,
		&cv.Specialization,
		&cv.WorkExperience,
		&cv.RawText,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CV{}, fmt.Errorf("cv not found: %w", err)
	}
	if err != nil {
		return domain.CV{}, fmt.Errorf("scan cv: %w", err)
	}

	skills, err := a.getSkillsByCV(ctx, id)
	if err != nil {
		return domain.CV{}, err
	}
	cv.Skills = skills

	return cv, nil
}

func (a *AdapterPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	// cv_skills удаляются каскадом (ON DELETE CASCADE)
	_, err := a.pool.Exec(ctx, DeleteCVQuery, id)
	if err != nil {
		return fmt.Errorf("delete cv: %w", err)
	}
	return nil
}

// upsertAndLinkSkills создаёт отсутствующие навыки и привязывает их к CV.
func (a *AdapterPostgres) upsertAndLinkSkills(ctx context.Context, tx pgx.Tx, cv *domain.CV) error {
	for i, skill := range cv.Skills {
		// Пытаемся вставить навык; если уже есть — пропускаем (ON CONFLICT DO NOTHING).
		newID := uuid.New()
		_, err := tx.Exec(ctx, CreateSkillQuery, newID, skill.Name)
		if err != nil {
			return fmt.Errorf("upsert skill %q: %w", skill.Name, err)
		}

		// Получаем реальный UUID навыка (мог быть создан ранее).
		var s domain.Skill
		if err := tx.QueryRow(ctx, GetSkillByNameQuery, skill.Name).Scan(&s.UUID, &s.Name); err != nil {
			return fmt.Errorf("get skill %q: %w", skill.Name, err)
		}

		cv.Skills[i] = s

		if _, err := tx.Exec(ctx, LinkSkillToCVQuery, cv.UUID, s.UUID); err != nil {
			return fmt.Errorf("link skill %q: %w", skill.Name, err)
		}
	}
	return nil
}

// getSkillsByCV возвращает все навыки, привязанные к CV.
func (a *AdapterPostgres) getSkillsByCV(ctx context.Context, cvID uuid.UUID) ([]domain.Skill, error) {
	rows, err := a.pool.Query(ctx, GetSkillsByCVIDQuery, cvID)
	if err != nil {
		return nil, fmt.Errorf("query skills: %w", err)
	}
	defer rows.Close()

	var skills []domain.Skill
	for rows.Next() {
		var s domain.Skill
		if err := rows.Scan(&s.UUID, &s.Name); err != nil {
			return nil, fmt.Errorf("scan skill: %w", err)
		}
		skills = append(skills, s)
	}
	return skills, rows.Err()
}
