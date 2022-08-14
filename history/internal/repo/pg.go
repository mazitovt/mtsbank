package repo

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"mtsbank/history/internal/config"
	"mtsbank/history/logger"
	"strings"
	"time"
)

var _ Repo = (*RepoPG)(nil)

type RepoPG struct {
	tableName string
	db        *sql.DB
	logger    logger.Logger
}

func (r *RepoPG) Currencies(ctx context.Context) ([]string, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: true})
	if err != nil {
		r.logger.Debug("BeginTx: err: %s", err)
		return nil, err
	}

	defer func() {
		if err = tx.Rollback(); err != nil {
			r.logger.Error("Rollback: err: %s", err)
		}
	}()

	q := "SELECT name FROM currency_pair"
	r.logger.Info("RepoPG.GetByTime: query: %s", q)

	rows, err := tx.QueryContext(ctx, q)
	if err != nil {
		r.logger.Debug("Tx.QueryContext: err: %s", err)
		return nil, err
	}

	var currencies []string

	defer rows.Close()
	for rows.Next() {
		currency := ""
		err = rows.Scan(&currency)
		if err != nil {
			r.logger.Debug("Row.Scan: ", err)
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	if rows.Err() != nil {
		r.logger.Debug("Rows.Err: ", err)
		return nil, err
	}

	if err = rows.Err(); err != nil {
		r.logger.Debug("Tx.ExecContext: select failed: %currencies", err)
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		r.logger.Debug("Tx.Commit: err: %s", err)
		return nil, err
	}

	return currencies, nil
}

func NewRepoPG(cfg *config.PostgresConfig, logger logger.Logger) (*RepoPG, error) {
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBname, cfg.Sslmode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &RepoPG{
		tableName: "currency_pair",
		db:        db,
		logger:    logger,
	}, nil
}

func (r *RepoPG) Insert(ctx context.Context, data []RegistryRow) error {
	r.logger.Debug("RepoPg.Insert: start")
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			r.logger.Error("Rollback: err: %s", err)
		}
	}()

	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*3)
	for i, v := range data {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", 3*i+1, 3*i+2, 3*i+3))
		valueArgs = append(valueArgs, v.CurrencyPair, v.Time.Round(time.Microsecond), v.Rate)
	}
	stmt := fmt.Sprintf("INSERT INTO registry(name, creation_time, rate) VALUES %s ON CONFLICT DO NOTHING",
		strings.Join(valueStrings, ","))

	// TODO: stmt[:len(stmt)%50]
	r.logger.Info("RepoPG.Insert: query: %s", stmt[:])
	r.logger.Info("%v", valueArgs)

	if _, err = tx.ExecContext(ctx, stmt, valueArgs...); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Tx.Commit: err: %s", err)
		return err
	}

	return nil
}

func (r *RepoPG) GetByTime(ctx context.Context, currencyPair string, start time.Time, end time.Time) ([]RegistryRow, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: true})
	if err != nil {
		r.logger.Debug("BeginTx: err: %s", err)
		return nil, err
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			r.logger.Error("Rollback: err: %s", err)
		}
	}()

	q := "SELECT name, creation_time, rate FROM registry WHERE name = $1 AND creation_time >= $2 AND creation_time <= $3 ORDER BY creation_time"
	r.logger.Info("RepoPG.GetByTime: query: %s", q)

	rows, err := tx.QueryContext(ctx, q, currencyPair, start, end)
	if err != nil {
		r.logger.Debug("Tx.QueryContext: err: %s", err)
		return nil, err
	}

	v := []RegistryRow{}

	defer rows.Close()
	for rows.Next() {
		row := RegistryRow{}
		err := rows.Scan(&row.CurrencyPair, &row.Time, &row.Rate)
		if err != nil {
			r.logger.Debug("Row.Scan: ", err)
			return nil, err
		}
		v = append(v, row)
	}

	err = rows.Err()

	if err != nil {
		r.logger.Debug("Rows.Err: ", err)
		return nil, err
	}

	if err := rows.Err(); err != nil {
		r.logger.Debug("Tx.ExecContext: select failed: %v", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		r.logger.Debug("Tx.Commit: err: %s", err)
		return nil, err
	}

	return v, nil
}

func (r *RepoPG) hasCurrencyPair(ctx context.Context, currencyPair string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: true})
	if err != nil {
		r.logger.Error("BeginTx: err: %s", err)
		return false, err
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			r.logger.Error("Rollback: err: %s", err)
		}
	}()

	q := "select exists(select 1 from currency_pair where name=$1)"
	r.logger.Info("RepoPG.GetPrincipal: query: %s", q)

	row := tx.QueryRowContext(ctx, q, currencyPair)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		r.logger.Error("Row.Scan: err: %s", err)
		return false, err
	}
	if err := row.Err(); err != nil {
		r.logger.Error("Tx.ExecContext: select failed: %v", err)
		return false, err
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Tx.Commit: err: %s", err)
		return false, err
	}

	return exists, nil
}

// TODO: remove migration to file
func (r *RepoPG) Migrate() error {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			// TODO: replace with logger
			fmt.Printf("Rollback: err: %v\n", err)
		}
	}()

	q := `CREATE TABLE currency_pair(
                      name text PRIMARY KEY 
);

CREATE TABLE registry(
                        name text REFERENCES currency_pair(name) NOT NULL,
                        creation_time timestamptz NOT NULL,
                        rate INT NOT NULL,
                        PRIMARY KEY (name, creation_time)

);

INSERT INTO currency_pair(name) VALUES ('EURUSD');
`

	_, err = tx.ExecContext(context.Background(), q)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
