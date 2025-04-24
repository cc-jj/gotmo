package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"local/tmo/api"
	"local/tmo/db"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// MockAPIClient implements a mock version of the API client for testing
type MockAPIClient struct {
	gateway api.GatewayResponse
	delay   time.Duration
}

func (m *MockAPIClient) GetGateway(ctx context.Context) (api.GatewayResponse, error) {
	time.Sleep(m.delay)
	return m.gateway, nil
}

func (m *MockAPIClient) Login(ctx context.Context) error {
	return nil
}

// setupTestDatabase creates a temporary SQLite database for benchmarking
func setupTestDatabase(b *testing.B) (*sql.DB, string) {
	// Create a temporary file for the SQLite database
	tmpFile, err := os.CreateTemp("", "benchmark-*.db")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	dsn := tmpFile.Name() + "?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL"

	// Connect to the SQLite database
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}

	// Create the schema
	err = createSchema(sqlDB)
	if err != nil {
		b.Fatalf("Failed to create schema: %v", err)
	}

	return sqlDB, dsn
}

// createSchema sets up the database schema for testing
func createSchema(db *sql.DB) error {
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("Failed to read schema file: %w", err)
	}
	_, err = db.Exec(string(schema))
	return err
}

// setupBenchmark creates a GatewayPoller with a real SQLite database
func setupBenchmark(b *testing.B) (*GatewayPoller, context.Context, func()) {
	// Create a real SQLite database
	sqlDB, dbDsn := setupTestDatabase(b)

	// Create a mock API client with test data
	mockAPIClient := &MockAPIClient{
		gateway: api.GatewayResponse{
			Device: api.Device{
				Serial:          "ABC123",
				SoftwareVersion: "1.0.0",
				FriendlyName:    "Test Gateway",
				HardwareVersion: "HW1",
				IsEnabled:       true,
				IsMeshSupported: true,
				MacID:           "00:11:22:33:44:55",
				Manufacturer:    "TestMfg",
				ManufacturerOUI: "TestOUI",
				Model:           "TestModel",
				Name:            "TestName",
				Role:            "Gateway",
				Type:            "TestType",
				UpdateState:     "Updated",
			},
			Time: api.Time{
				LocalTime: 1617235200, // April 1, 2021 UTC
				UpTime:    3600,       // 1 hour
			},
			Signal: api.Signal{
				FourG: api.SignalStats{
					AntennaUsed: "internal",
					Bands:       []string{"B2"},
					Bars:        3,
					Cid:         12345,
					ENBID:       67890,
					GNBID:       0,
					Rsrp:        -90,
					Rsrq:        -12,
					Rssi:        -80,
					Sinr:        15,
				},
				FiveG: api.SignalStats{
					AntennaUsed: "internal",
					Bands:       []string{"n41"},
					Bars:        4,
					Cid:         54321,
					ENBID:       0,
					GNBID:       98765,
					Rsrp:        -85,
					Rsrq:        -10,
					Rssi:        -75,
					Sinr:        20,
				},
			},
		},
	}

	// Create the GatewayPoller
	poller := &GatewayPoller{
		config: Config{
			DBDSN:        dbDsn,
			GatewayURL:   "http://localhost",
			PollDuration: 5 * time.Minute,
			Logger:       log.New(io.Discard, "", 0), // Silent logger for benchmarks
		},
		db:        sqlDB,
		apiClient: mockAPIClient,
		queries:   db.New(sqlDB),
	}

	// Return cleanup function
	cleanup := func() {
		sqlDB.Close()
	}

	// Return the poller, context and cleanup function
	return poller, context.Background(), cleanup
}

// BenchmarkPoll benchmarks the Poll method
func BenchmarkPoll(b *testing.B) {
	poller, ctx, cleanup := setupBenchmark(b)
	defer cleanup()

	for b.Loop() {
		err := poller.Poll(ctx)
		if err != nil {
			b.Fatalf("Poll failed: %v", err)
		}
	}
}

// BenchmarkWithVariableData tests Poll with different data each time
func BenchmarkPollWithVariableData(b *testing.B) {
	poller, ctx, cleanup := setupBenchmark(b)
	defer cleanup()

	// Get original gateway for modification
	mockClient := poller.apiClient.(*MockAPIClient)
	originalGateway := mockClient.gateway

	i := 0
	for b.Loop() {
		// Modify the gateway data for each iteration
		i++
		gateway := originalGateway
		gateway.Device.Serial = fmt.Sprintf("ABC%d", i)
		gateway.Time.LocalTime = int(time.Now().Unix())
		gateway.Time.UpTime = int(3600 + i)
		gateway.Signal.FourG.Rsrp = -90 - i%10
		gateway.Signal.FiveG.Rsrp = -85 - i%10

		mockClient.gateway = gateway

		err := poller.Poll(ctx)
		if err != nil {
			b.Fatalf("Poll failed: %v", err)
		}
	}
}

// BenchmarkPollParallel benchmarks the Poll method with parallel execution
func BenchmarkPollParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine needs its own setup
		poller, ctx, cleanup := setupBenchmark(b)
		defer cleanup()

		counter := 0
		for pb.Next() {
			// Modify serial to avoid conflicts in parallel runs
			mockClient := poller.apiClient.(*MockAPIClient)
			gateway := mockClient.gateway
			gateway.Device.Serial = fmt.Sprintf("PARALLEL-%d-%d", b.N, counter)
			mockClient.gateway = gateway

			err := poller.Poll(ctx)
			if err != nil {
				b.Fatalf("Parallel Poll failed: %v", err)
			}
			counter++
		}
	})
}

// BenchmarkTransactionOnly benchmarks just the database transaction overhead
func BenchmarkTransactionOnly(b *testing.B) {
	poller, ctx, cleanup := setupBenchmark(b)
	defer cleanup()

	for b.Loop() {
		tx, err := poller.db.BeginTx(ctx, nil)
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		// Do minimal work inside transaction
		_, err = tx.Exec("SELECT 1")
		if err != nil {
			tx.Rollback()
			b.Fatalf("Failed to execute simple query: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			b.Fatalf("Failed to commit transaction: %v", err)
		}
	}
}

// Create wrapper for API client with delay
type DelayedAPIClient struct {
	originalClient *MockAPIClient
	delayDuration  time.Duration
}

// BenchmarkRealWorldScenario simulates a more realistic polling scenario
func BenchmarkRealWorldScenario(b *testing.B) {
	poller, ctx, cleanup := setupBenchmark(b)
	defer cleanup()

	// Create a delayed API client
	delayedAPIClient := MockAPIClient{
		gateway: poller.apiClient.(*MockAPIClient).gateway,
		delay:   100 * time.Millisecond,
	}

	poller.apiClient = &delayedAPIClient

	i := 0
	for b.Loop() {
		// Modify the gateway data for each iteration to simulate changing data
		i++
		gateway := delayedAPIClient.gateway
		gateway.Device.Serial = fmt.Sprintf("ABC%d", i)
		gateway.Time.LocalTime = int(time.Now().Unix())
		gateway.Time.UpTime = int(3600 + i)

		delayedAPIClient.gateway = gateway

		err := poller.Poll(ctx)
		if err != nil {
			b.Fatalf("Poll failed: %v", err)
		}

		// Add a small delay between polls to simulate real-world polling frequency
		if i < b.N-1 {
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// BenchmarkInsertionVolume tests how performance scales with more data in the database
func BenchmarkInsertionVolume(b *testing.B) {
	// Only run a few iterations for this test as we'll pre-populate the database
	if b.N > 10 {
		b.N = 10
	}

	poller, ctx, cleanup := setupBenchmark(b)
	defer cleanup()

	// Pre-populate the database with many entries
	prePopulateEntries := 1000
	b.Logf("Pre-populating database with %d entries", prePopulateEntries)

	mockClient := poller.apiClient.(*MockAPIClient)
	originalGateway := mockClient.gateway

	// Disable timer during population
	b.StopTimer()

	for i := range prePopulateEntries {
		gateway := originalGateway
		gateway.Device.Serial = fmt.Sprintf("VOLUME-%d", i)
		gateway.Time.LocalTime = int(time.Now().Unix() - int64(i*60)) // One minute apart
		mockClient.gateway = gateway

		err := poller.Poll(ctx)
		if err != nil {
			b.Fatalf("Pre-population failed: %v", err)
		}
	}

	// Reset to original gateway
	mockClient.gateway = originalGateway

	// Start timer for actual benchmark
	b.StartTimer()

	for b.Loop() {
		err := poller.Poll(ctx)
		if err != nil {
			b.Fatalf("Poll failed: %v", err)
		}
	}
}
