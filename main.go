package main

import (
	"context"
	"database/sql"
	"fmt"
	"local/tmo/api"
	"local/tmo/db"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config holds application configuration
type Config struct {
	DBDSN        string
	GatewayURL   string
	PollDuration time.Duration
	Logger       *log.Logger
}

// GatewayPoller handles the polling of the gateway and storing data
type GatewayPoller struct {
	config    Config
	db        *sql.DB
	apiClient api.IClient
	queries   *db.Queries
}

// NewGatewayPoller creates a new GatewayPoller
func NewGatewayPoller(config Config) *GatewayPoller {
	if config.Logger == nil {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		config.Logger = logger
	}

	return &GatewayPoller{
		config: config,
	}
}

// Initialize sets up the database connection and API client
func (p *GatewayPoller) Initialize(ctx context.Context) error {
	// Set up database
	var err error
	p.db, err = newDB(ctx, p.config.DBDSN)
	if err != nil {
		return fmt.Errorf("database initialization failed: %w", err)
	}

	// Set up API client
	p.apiClient = api.NewClient(p.config.GatewayURL, p.config.Logger)
	err = p.apiClient.Login(ctx)
	if err != nil {
		return fmt.Errorf("API login failed: %w", err)
	}

	// Set up queries
	p.queries = db.New(p.db)

	return nil
}

// Run starts the polling loop
func (p *GatewayPoller) Run(ctx context.Context) error {
	// Initial poll
	err := p.Poll(ctx)
	if err != nil {
		return err
	}

	timer := time.NewTimer(p.chooseDuration())

	for {
		select {
		case <-timer.C:
			err := p.Poll(ctx)
			if err != nil {
				if strings.Contains(err.Error(), "network is unreachable") {
					p.config.Logger.Println("Network unreachable, is the Gateway down?")
				} else {
					return err
				}
			}
			timer.Reset(p.chooseDuration())
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Poll fetches data from the gateway and stores it in the database
func (p *GatewayPoller) Poll(ctx context.Context) error {
	gateway, err := p.apiClient.GetGateway(ctx)
	if err != nil {
		return fmt.Errorf("error getting gateway from API: %w", err)
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	queries := p.queries.WithTx(tx)

	device, err := p.loadDevice(ctx, queries, gateway.Device)
	if err != nil {
		return fmt.Errorf("error loading device: %w", err)
	}

	snapshot, err := p.loadSnapshot(ctx, queries, device, gateway.Time)
	if err != nil {
		return fmt.Errorf("error loading snapshot: %w", err)
	}

	err = p.loadSignal(ctx, queries, snapshot, "4G", gateway.Signal.FourG)
	if err != nil {
		return fmt.Errorf("error loading 4G signal: %w", err)
	}

	err = p.loadSignal(ctx, queries, snapshot, "5G", gateway.Signal.FiveG)
	if err != nil {
		return fmt.Errorf("error loading 5G signal: %w", err)
	}

	return tx.Commit()
}

// loadDevice returns the existing device or creates a new one
func (p *GatewayPoller) loadDevice(ctx context.Context, queries *db.Queries, apiDevice api.Device) (db.Device, error) {
	device, err := queries.GetDevice(ctx, db.GetDeviceParams{
		Serial:          apiDevice.Serial,
		SoftwareVersion: apiDevice.SoftwareVersion,
	})

	if err == nil {
		return device, nil
	}

	if err != sql.ErrNoRows {
		return db.Device{}, err
	}

	p.config.Logger.Println("Creating new device")
	return queries.CreateDevice(ctx, db.CreateDeviceParams{
		FriendlyName:    apiDevice.FriendlyName,
		HardwareVersion: apiDevice.HardwareVersion,
		Isenabled:       apiDevice.IsEnabled,
		IsmeshSupported: apiDevice.IsMeshSupported,
		Macid:           apiDevice.MacID,
		Manufacturer:    apiDevice.Manufacturer,
		ManufacturerOui: apiDevice.ManufacturerOUI,
		Model:           apiDevice.Model,
		Name:            apiDevice.Name,
		Role:            apiDevice.Role,
		Serial:          apiDevice.Serial,
		SoftwareVersion: apiDevice.SoftwareVersion,
		Type:            apiDevice.Type,
		UpdateState:     apiDevice.UpdateState,
	})
}

// loadSnapshot inserts a new snapshot record into the database
func (p *GatewayPoller) loadSnapshot(ctx context.Context, queries *db.Queries, device db.Device, apiTime api.Time) (db.Snapshot, error) {
	return queries.CreateSnapshot(ctx, db.CreateSnapshotParams{
		Deviceid:  device.ID,
		CreatedAt: time.Unix(int64(apiTime.LocalTime), 0),
		Uptime:    int64(apiTime.UpTime),
	})
}

// loadSignal inserts a new signal record into the database
func (p *GatewayPoller) loadSignal(ctx context.Context, queries *db.Queries, snapshot db.Snapshot, statName string, stats api.SignalStats) error {
	if statName != "4G" && statName != "5G" {
		return fmt.Errorf("invalid statName: %s", statName)
	}

	if len(stats.Bands) != 1 {
		return fmt.Errorf("expected 1 band, got %+v", stats.Bands)
	}

	_, err := queries.CreateSignal(ctx, db.CreateSignalParams{
		Snapshotid:  snapshot.ID,
		Generation:  statName,
		AntennaUsed: stats.AntennaUsed,
		Band:        stats.Bands[0],
		Bars:        stats.Bars,
		Cid:         int64(stats.Cid),
		Enbid:       int64(stats.ENBID),
		Gnbid:       int64(stats.GNBID),
		Rsrp:        int64(stats.Rsrp),
		Rsrq:        int64(stats.Rsrq),
		Rssi:        int64(stats.Rssi),
		Sinr:        int64(stats.Sinr),
	})

	return err
}

// chooseDuration determines polling frequency based on time of day
func (p *GatewayPoller) chooseDuration() time.Duration {
	midnightDuration := 1 * time.Minute
	if p.config.PollDuration < midnightDuration {
		return p.config.PollDuration
	}

	now := time.Now()
	if now.Hour() >= 23 || now.Hour() < 3 {
		return midnightDuration
	}

	return p.config.PollDuration
}

// newDB initializes and tests a database connection
func newDB(ctx context.Context, dsn string) (*sql.DB, error) {
	sqlDb, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	sqlDb.SetMaxOpenConns(1)
	sqlDb.SetMaxIdleConns(1)

	if err = sqlDb.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return sqlDb, nil
}

// defaultPollDuration retrieves the polling duration from environment or uses default
func defaultPollDuration() time.Duration {
	s := os.Getenv("GATEWAY_POLL_FREQ")
	if s == "" {
		s = "5m"
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		log.Fatal("Invalid GATEWAY_POLL_FREQ")
	}

	return duration
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	config := Config{
		DBDSN:        "file:tmo.db?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL",
		GatewayURL:   "http://192.168.12.1/TMI/v1",
		PollDuration: defaultPollDuration(),
		Logger:       logger,
	}

	ctx := context.Background()

	poller := NewGatewayPoller(config)

	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := poller.Initialize(timeoutCtx)
	if err != nil {
		logger.Fatalf("Failed to initialize poller: %v", err)
	}

	err = poller.Run(ctx)
	if err != nil {
		logger.Fatalf("Poller exited with error: %v", err)
	}
}
