package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

// Service для управления миграциями базы данных
type Service struct {
	db             *sql.DB
	migrationsPath string
}

// NewService создает новый экземпляр сервиса миграций
func NewService(db *sql.DB, migrationsPath string) *Service {
	return &Service{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

// Migration представляет информацию о миграции
type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

// MigrationStatus представляет статус миграции
type MigrationStatus struct {
	Version   int
	Name      string
	Applied   bool
	AppliedAt *string
}

// Apply применяет все непримененные миграции
func (s *Service) Apply() error {
	log.Println("Начинаем применение миграций...")

	// Создаем таблицу для отслеживания миграций
	if err := s.createMigrationsTable(); err != nil {
		return fmt.Errorf("не удалось создать таблицу миграций: %w", err)
	}

	// Загружаем все миграции
	migrations, err := s.loadMigrations()
	if err != nil {
		return fmt.Errorf("не удалось загрузить миграции: %w", err)
	}

	// Получаем список примененных миграций
	appliedVersions, err := s.getAppliedVersions()
	if err != nil {
		return fmt.Errorf("не удалось получить список примененных миграций: %w", err)
	}

	// Применяем миграции в порядке версий
	appliedCount := 0
	for _, migration := range migrations {
		if appliedVersions[migration.Version] {
			log.Printf("Миграция %d (%s) уже применена, пропускаем", migration.Version, migration.Name)
			continue
		}

		log.Printf("Применяем миграцию %d (%s)...", migration.Version, migration.Name)
		if err := s.applyMigration(migration); err != nil {
			return fmt.Errorf("ошибка применения миграции %d (%s): %w", migration.Version, migration.Name, err)
		}

		// Отмечаем миграцию как примененную
		if err := s.markAsApplied(migration.Version, migration.Name); err != nil {
			return fmt.Errorf("не удалось отметить миграцию %d как примененную: %w", migration.Version, err)
		}

		log.Printf("Миграция %d (%s) успешно применена", migration.Version, migration.Name)
		appliedCount++
	}

	if appliedCount == 0 {
		log.Println("Все миграции уже применены")
	} else {
		log.Printf("Применено миграций: %d", appliedCount)
	}

	return nil
}

// createMigrationsTable создает таблицу для отслеживания примененных миграций
func (s *Service) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	_, err := s.db.Exec(query)
	return err
}

// loadMigrations загружает все миграции из папки migrations
func (s *Service) loadMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(s.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать папку миграций %s: %w", s.migrationsPath, err)
	}

	migrationsMap := make(map[int]Migration)

	// Сначала загружаем все up миграции
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// Пропускаем файлы с суффиксом _down.sql на первом проходе
		if strings.HasSuffix(filename, "_down.sql") {
			continue
		}

		// Парсим номер миграции из имени файла (например, 001_create_users_table.sql)
		if !strings.HasSuffix(filename, ".sql") {
			continue
		}

		// Извлекаем номер из начала имени файла
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("Предупреждение: не удалось распарсить версию из файла %s, пропускаем", filename)
			continue
		}

		// Читаем содержимое файла миграции
		filePath := filepath.Join(s.migrationsPath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать файл миграции %s: %w", filePath, err)
		}

		name := strings.TrimSuffix(filename, ".sql")
		migrationsMap[version] = Migration{
			Version: version,
			Name:    name,
			UpSQL:   string(content),
		}
	}

	// Теперь загружаем down миграции
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// Обрабатываем только файлы с суффиксом _down.sql
		if !strings.HasSuffix(filename, "_down.sql") {
			continue
		}

		// Извлекаем номер из начала имени файла
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("Предупреждение: не удалось распарсить версию из файла %s, пропускаем", filename)
			continue
		}

		// Читаем содержимое файла down миграции
		filePath := filepath.Join(s.migrationsPath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("не удалось прочитать файл down миграции %s: %w", filePath, err)
		}

		// Добавляем DownSQL к существующей миграции или создаем новую
		if migration, exists := migrationsMap[version]; exists {
			migration.DownSQL = string(content)
			migrationsMap[version] = migration
		} else {
			name := strings.TrimSuffix(filename, "_down.sql")
			migrationsMap[version] = Migration{
				Version: version,
				Name:    name,
				DownSQL: string(content),
			}
		}
	}

	// Сортируем миграции по версии
	var migrations []Migration
	for version := range migrationsMap {
		migrations = append(migrations, migrationsMap[version])
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedVersions возвращает множество версий примененных миграций
func (s *Service) getAppliedVersions() (map[int]bool, error) {
	rows, err := s.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if rows != nil {
			if err := rows.Close(); err != nil {
				fmt.Println("close rows err:", err)
			}
		}
	}(rows)

	appliedVersions := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		appliedVersions[version] = true
	}

	return appliedVersions, rows.Err()
}

// applyMigration применяет конкретную миграцию
func (s *Service) applyMigration(migration Migration) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			fmt.Println("rollback:", err)
		}
	}(tx)

	// Выполняем SQL миграции
	if _, err := tx.Exec(migration.UpSQL); err != nil {
		return fmt.Errorf("ошибка выполнения SQL: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("не удалось зафиксировать транзакцию: %w", err)
	}

	return nil
}

// markAsApplied отмечает миграцию как примененную
func (s *Service) markAsApplied(version int, name string) error {
	query := `
		INSERT INTO schema_migrations (version, name, applied_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (version) DO NOTHING
	`
	_, err := s.db.Exec(query, version, name)
	return err
}

// Down откатывает последнюю примененную миграцию
func (s *Service) Down() error {
	log.Println("Начинаем откат миграций...")

	// Создаем таблицу для отслеживания миграций (если еще не создана)
	if err := s.createMigrationsTable(); err != nil {
		return fmt.Errorf("не удалось создать таблицу миграций: %w", err)
	}

	// Получаем последнюю примененную миграцию
	var version int
	var name string
	err := s.db.QueryRow("SELECT version, name FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version, &name)
	if err == sql.ErrNoRows {
		return fmt.Errorf("нет примененных миграций для отката")
	}
	if err != nil {
		return fmt.Errorf("не удалось получить последнюю миграцию: %w", err)
	}

	// Загружаем все миграции
	migrations, err := s.loadMigrations()
	if err != nil {
		return fmt.Errorf("не удалось загрузить миграции: %w", err)
	}

	// Находим миграцию для отката
	var migration *Migration
	for i := range migrations {
		if migrations[i].Version == version {
			migration = &migrations[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("миграция %d (%s) не найдена в файлах", version, name)
	}

	if migration.DownSQL == "" {
		return fmt.Errorf("для миграции %d (%s) не найден файл отката", version, name)
	}

	log.Printf("Откатываем миграцию %d (%s)...", version, name)

	// Откатываем миграцию
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			fmt.Println("rollback:", err)
		}
	}(tx)

	if _, err := tx.Exec(migration.DownSQL); err != nil {
		return fmt.Errorf("ошибка выполнения SQL отката: %w", err)
	}

	// Удаляем запись о миграции
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
		return fmt.Errorf("не удалось удалить запись о миграции: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("не удалось зафиксировать транзакцию: %w", err)
	}

	log.Printf("Миграция %d (%s) успешно откачена", version, name)
	return nil
}

// Status возвращает статус всех миграций
func (s *Service) Status() ([]MigrationStatus, error) {
	// Создаем таблицу для отслеживания миграций (если еще не создана)
	if err := s.createMigrationsTable(); err != nil {
		return nil, fmt.Errorf("не удалось создать таблицу миграций: %w", err)
	}

	// Загружаем все миграции
	migrations, err := s.loadMigrations()
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить миграции: %w", err)
	}

	// Получаем список примененных миграций
	rows, err := s.db.Query("SELECT version, name, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список примененных миграций: %w", err)
	}
	defer func(rows *sql.Rows) {
		if rows != nil {
			if err := rows.Close(); err != nil {
				fmt.Println("close rows:", err)
			}
		}
	}(rows)

	appliedMap := make(map[int]struct {
		name      string
		appliedAt string
	})
	for rows.Next() {
		var version int
		var name string
		var appliedAt string
		if err := rows.Scan(&version, &name, &appliedAt); err != nil {
			return nil, err
		}
		appliedMap[version] = struct {
			name      string
			appliedAt string
		}{name: name, appliedAt: appliedAt}
	}

	// Формируем статус для каждой миграции
	statuses := make([]MigrationStatus, 0, len(migrations))
	for _, migration := range migrations {
		status := MigrationStatus{
			Version: migration.Version,
			Name:    migration.Name,
			Applied: false,
		}
		if applied, ok := appliedMap[migration.Version]; ok {
			status.Applied = true
			status.AppliedAt = &applied.appliedAt
		}
		statuses = append(statuses, status)
	}

	return statuses, rows.Err()
}
