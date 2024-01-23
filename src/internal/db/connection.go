package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	pgx_migrate "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

type DbConfig struct {
	Host         string
	Port         string
	Username     string
	DatabaseName string
	Password     string
	Timezone     string
}

func CreateDbConnection(ctx context.Context, cfg *DbConfig) (*pgx.Conn, error) {
	psqlCfg, err := getDbConnectionString(cfg)
	if err != nil {
		return nil, err
	}

	connConfig, err := pgx.ParseConfig(psqlCfg)
	if err != nil {
		return nil, err
	}

	// err = db_common.AddRootCertToConfig(&connConfig.Config, getRootCertLocation())
	// if err != nil {
	// 	return nil, err
	// }

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func MigrateSchema(ctx context.Context, cfg *DbConfig) (ok bool, err error) {
	psqlCfg := getMigrationConnectionString(cfg)

	p := &pgx_migrate.Postgres{}
	d, err := p.Open(psqlCfg)
	if err != nil {
		return false, err
	}
	m, err := migrate.NewWithDatabaseInstance("file://database/migration", "pgx", d)
	if err != nil {
		return false, err
	}
	m.Up()
	return true, nil
}

func getDbConnectionString(cfg *DbConfig) (string, error) {
	if cfg == nil {
		cfg = &DbConfig{}
	}

	if len(cfg.Username) == 0 {
		cfg.Username = "postgres" // constants.DatabaseUsername
	}

	psqlCfgMap := map[string]string{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"user":     cfg.Username,
		"dbname":   cfg.DatabaseName,
		"password": cfg.Password,
	}

	psqlInfo := []string{}
	for k, v := range psqlCfgMap {
		psqlInfo = append(psqlInfo, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(psqlInfo, " "), nil
}

func getMigrationConnectionString(cfg *DbConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
}
