package main

import (
	"context"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/liamshahgholi/ticket-golang-project/internal/api"
	db "github.com/liamshahgholi/ticket-golang-project/internal/db/sqlc"
	"github.com/liamshahgholi/ticket-golang-project/internal/routes"
	"github.com/liamshahgholi/ticket-golang-project/internal/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	if config.APPDEBUG == "true" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Connect to the database
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.DBUSERNAME,
		config.DBPASSWORD,
		config.DBHOST,
		config.DBPORT,
		config.DBDATABASE,
	)

	connPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	defer connPool.Close()

	runDBMigration(config.MigrationURL, dsn)

	store := db.NewStore(connPool)

	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	// Register the routes
	if err := routes.SetupRoutes(server); err != nil {
		log.Fatal().Err(err).Msg("failed to set up routes")
	}

	if err := server.Start(config.APPPORT); err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}
