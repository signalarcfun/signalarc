package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const checkTimeout = 3 * time.Second

type DB struct {
	pool *pgxpool.Pool
}

type SchemaValidation struct {
	Status           string   `json:"status"`
	MigrationVersion int      `json:"migration_version"`
	Dirty            bool     `json:"dirty"`
	MissingTables    []string `json:"missing_tables"`
	MissingColumns   []string `json:"missing_columns"`
}

var expectedPhase2Tables = []string{
	"users",
	"wallets",
	"markets",
	"positions",
	"trades",
	"liquidity",
	"resolutions",
	"settlements",
	"oracle_events",
	"audit_logs",
	"api_keys",
	"webhooks",
	"agent_access",
	"agent_wallets",
	"agent_intents",
	"agent_executions",
	"analytics_indexer_state",
	"analytics_markets",
	"analytics_events",
	"analytics_summary_cache",
}

var expectedMarketColumns = []string{
	"cover_image_url",
	"market_contract_address",
	"market_deployment_tx_hash",
	"market_factory_address",
	"resolver_address",
	"onchain_deployment_status",
}

func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	db := &DB{pool: pool}
	if err := db.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Ping(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, checkTimeout)
	defer cancel()

	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

func (db *DB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

func (db *DB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...)
}

func (db *DB) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := db.pool.Exec(ctx, sql, args...)
	return err
}

func (db *DB) ValidateSchema(parent context.Context) (SchemaValidation, error) {
	ctx, cancel := context.WithTimeout(parent, checkTimeout)
	defer cancel()

	result := SchemaValidation{
		Status:         "error",
		MissingTables:  []string{},
		MissingColumns: []string{},
	}

	var migrationsExists bool
	err := db.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
				AND table_name = 'schema_migrations'
		)
	`).Scan(&migrationsExists)
	if err != nil {
		return result, fmt.Errorf("check schema_migrations table: %w", err)
	}
	if !migrationsExists {
		result.MissingTables = append(result.MissingTables, "schema_migrations")
		return result, nil
	}

	if err := db.pool.QueryRow(ctx, `SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&result.MigrationVersion, &result.Dirty); err != nil {
		return result, fmt.Errorf("read schema_migrations: %w", err)
	}

	rows, err := db.pool.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
			AND table_name = ANY($1)
	`, expectedPhase2Tables)
	if err != nil {
		return result, fmt.Errorf("read phase 2 tables: %w", err)
	}
	defer rows.Close()

	existing := make(map[string]bool, len(expectedPhase2Tables))
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return result, fmt.Errorf("scan phase 2 table: %w", err)
		}
		existing[tableName] = true
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("iterate phase 2 tables: %w", err)
	}

	for _, tableName := range expectedPhase2Tables {
		if !existing[tableName] {
			result.MissingTables = append(result.MissingTables, tableName)
		}
	}

	columnRows, err := db.pool.Query(ctx, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
			AND table_name = 'markets'
			AND column_name = ANY($1)
	`, expectedMarketColumns)
	if err != nil {
		return result, fmt.Errorf("read market columns: %w", err)
	}
	defer columnRows.Close()

	existingColumns := make(map[string]bool, len(expectedMarketColumns))
	for columnRows.Next() {
		var columnName string
		if err := columnRows.Scan(&columnName); err != nil {
			return result, fmt.Errorf("scan market column: %w", err)
		}
		existingColumns[columnName] = true
	}
	if err := columnRows.Err(); err != nil {
		return result, fmt.Errorf("iterate market columns: %w", err)
	}

	for _, columnName := range expectedMarketColumns {
		if !existingColumns[columnName] {
			result.MissingColumns = append(result.MissingColumns, "markets."+columnName)
		}
	}

	if result.MigrationVersion == 21 && !result.Dirty && len(result.MissingTables) == 0 && len(result.MissingColumns) == 0 {
		result.Status = "ok"
	}

	return result, nil
}
