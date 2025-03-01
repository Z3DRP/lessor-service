package dac

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Z3DRP/lessor-service/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Persister interface {
	GetDB() *sql.DB
	GetBunDB() *bun.DB
}

type Store struct {
	db  *sql.DB
	BdB bun.DB
}

func (s Store) GetDB() *sql.DB {
	return s.db
}

func (s Store) GetBunDB() *bun.DB {
	return &s.BdB
}

type StoreBuilder struct {
	db  *sql.DB
	bdb bun.DB
}

func NewBuilder() *StoreBuilder {
	return &StoreBuilder{}
}

func (b *StoreBuilder) SetDB(db *sql.DB) *StoreBuilder {
	b.db = db
	return b
}

func (b *StoreBuilder) SetBunDB() *StoreBuilder {
	if b.db == nil {
		panic("SetDb must be called first")
	}
	b.bdb = *bun.NewDB(b.db, pgdialect.New())
	return b
}

func (b *StoreBuilder) Build() Persister {
	if b.db == nil {
		panic("Database connecciton must be set before building")
	}
	return &Store{
		db:  b.db,
		BdB: b.bdb,
	}
}

func InitStore(db *sql.DB) Store {
	return Store{
		db:  db,
		BdB: *bun.NewDB(db, pgdialect.New()),
	}
}

func Con() *sql.DB {
	dbc := "postgres://postgres:zroot_1119@18.226.170.114:5432/alessor?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dbc)))
	return sqldb
}

func (s Store) TestConnection() error {
	return s.db.Ping()
}

func DbCon(dbConf *config.DbConfig) (*sql.DB, error) {
	rootCertPool := x509.NewCertPool()
	path := dbConf.SslRoot
	pem, err := os.ReadFile(path)

	if err != nil {
		log.Printf("failed to read CA certificate: %v", err)
		return nil, err
	}

	if !rootCertPool.AppendCertsFromPEM(pem) {
		log.Println("fialed to append CA certificate")
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs:            rootCertPool,
		InsecureSkipVerify: true, // might need to change if set to false current config will reject bc handshake
	}

	pgcon := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(fmt.Sprintf("%v:%v", dbConf.Host, dbConf.Port)),
		pgdriver.WithTLSConfig(tlsConfig),
		pgdriver.WithUser(dbConf.DbUsr),
		pgdriver.WithPassword(dbConf.DbPwd),
		pgdriver.WithDatabase(dbConf.DbName),
		pgdriver.WithDialTimeout(time.Second*time.Duration(dbConf.DialTimeout)),
		pgdriver.WithReadTimeout(time.Second*time.Duration(dbConf.ReadTimeout)),
		pgdriver.WithWriteTimeout(time.Second*time.Duration(dbConf.WriteTimeout)),
	)
	db := sql.OpenDB(pgcon)
	db.SetMaxOpenConns(dbConf.MaxOpenConns)
	db.SetMaxIdleConns(dbConf.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(dbConf.ConnTimeout) * time.Minute)

	err = db.Ping()
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		return nil, err
	}

	return db, nil
}
