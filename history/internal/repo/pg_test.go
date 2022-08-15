package repo

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/mazitovt/logger"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"mtsbank/history/internal/config"
	"testing"
	"time"
)

func TestRepoPG_CreatePrincipal(t *testing.T) {

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14.3-alpine3.16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "history",
			"POSTGRES_USER":     "history",
			"POSTGRES_PASSWORD": "history",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(context.TODO(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Fatal(err)
	}

	ip, err := container.Host(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	mappedPort, err := container.MappedPort(context.TODO(), "5432")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ip, mappedPort.Port())
	if err := migrate(fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s", "history", "history", ip, mappedPort.Port(), "history", "disable")); err != nil {
		t.Fatal("migrate: ", err)
	}

	r, err := NewRepoPG(&config.PostgresConfig{
		Host:     ip,
		Port:     mappedPort.Port(),
		User:     "history",
		Password: "history",
		DBname:   "history",
		Sslmode:  "disable",
	}, logger.New(logger.Debug))

	if err != nil {
		t.Fatal(err)
	}

	if err = r.Insert(context.Background(), []RegistryRow{
		{"EURUSD", time.Now(), 45},
		{"EURUSD", time.Now(), 45},
		{"USDRUB", time.Now(), 2},
	}); err != nil {
		t.Fatal("Insert: ", err)
	}

	f, err := r.hasCurrencyPair(context.Background(), "RUB")
	t.Log(f, err)
	f, err = r.hasCurrencyPair(context.Background(), "EURUSD")
	t.Log(f, err)

	byTime, err := r.GetByTime(context.Background(), "EURUSD", time.Now().Truncate(time.Hour), time.Now().Add(time.Hour))
	t.Log(byTime, err)

	cur, err := r.Currencies(context.Background())
	t.Log(cur, err)

}

func migrate(dsn string) error {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			fmt.Printf("Rollback: err: %v\n", err)
		}
	}()

	q := `CREATE TABLE currency_pair(
                      name text PRIMARY KEY 
);

CREATE TABLE registry(
                        name text REFERENCES currency_pair(name) NOT NULL,
                        creation_time timestamp NOT NULL,
                        rate INT NOT NULL,
                        PRIMARY KEY (name, creation_time)

);

INSERT INTO currency_pair(name) VALUES ('EURUSD'),('USDRUB'),('USDJPY');
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
