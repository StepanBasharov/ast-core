package postgres

const (
	// users_cv

	CreateCVQuery = `
		INSERT INTO users_cv (uuid, first_name, last_name, cv_title, specialization, work_experience, raw_text)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	UpdateCVQuery = `
		UPDATE users_cv
		SET first_name = $2, last_name = $3, cv_title = $4, specialization = $5, work_experience = $6, raw_text = $7
		WHERE uuid = $1;`

	GetCVByIDQuery = `
		SELECT uuid, first_name, last_name, cv_title, specialization, work_experience, raw_text
		FROM users_cv
		WHERE uuid = $1;`

	DeleteCVQuery = `
		DELETE FROM users_cv
		WHERE uuid = $1;`

	// skills

	CreateSkillQuery = `
		INSERT INTO skills (uuid, name)
		VALUES ($1, $2)
		ON CONFLICT (name) DO NOTHING;`

	GetSkillByNameQuery = `
		SELECT uuid, name
		FROM skills
		WHERE name = $1;`

	// cv_skills (join table)

	LinkSkillToCVQuery = `
		INSERT INTO cv_skills (cv_uuid, skill_uuid)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;`

	UnlinkAllSkillsFromCVQuery = `
		DELETE FROM cv_skills
		WHERE cv_uuid = $1;`

	GetSkillsByCVIDQuery = `
		SELECT s.uuid, s.name
		FROM skills s
		JOIN cv_skills cs ON cs.skill_uuid = s.uuid
		WHERE cs.cv_uuid = $1;`
)
