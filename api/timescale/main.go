package timescale

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ivinayakg/shorte.live/api/database"
	"github.com/jackc/pgx/v5"
)

var TimescaleDB *pgx.Conn

// connect to database using a single connection
func SetupTimeScale() {
	/***********************************************/
	/* Single Connection to TimescaleDB/ PostgreSQL */
	/***********************************************/
	ctx := context.Background()
	connStr := os.Getenv("TIMESCALE_CONN")
	db, err := pgx.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	click_events_table_name := os.Getenv("CLICK_EVENTS_TABLE_NAME")

	// create click events table
	_, err = db.Exec(ctx, fmt.Sprintf("CREATE TABLE IF not exists %v (url_id VARCHAR(50), geo TEXT,device TEXT,os TEXT,referrer TEXT,timestamp TIMESTAMPTZ NOT NULL);", click_events_table_name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create Table failed: %v\n", err)
		os.Exit(1)
	}

	// checks if hypertable exists for click events
	var click_events_hypertable_exists bool
	err = db.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM timescaledb_information.hypertables WHERE hypertable_name = '%v');", click_events_table_name)).Scan(&click_events_hypertable_exists)
	fmt.Println(click_events_hypertable_exists)
	if err != nil {
		fmt.Println(err)
		// fmt.Fprintf(os.Stderr, "Create %v index failed: %v\n", click_events_url_timestamp_index_name, err)
		os.Exit(1)
	}

	//create hypertable if not already there
	if !click_events_hypertable_exists {
		_, err := db.Exec(ctx, fmt.Sprintf("SELECT create_hypertable('%v', by_range('timestamp'))", click_events_table_name))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Create hypertable failed: %v\n", err)
			os.Exit(1)
		}
	}

	click_events_url_timestamp_index_name := "ix_url_timestamp"
	// create index on url_id and time
	_, err = db.Exec(ctx, fmt.Sprintf("CREATE INDEX IF not EXISTS %v ON %v (url_id, timestamp DESC);", click_events_url_timestamp_index_name, click_events_table_name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create %v index failed: %v\n", click_events_url_timestamp_index_name, err)
		os.Exit(1)
	}

	fmt.Println("Connected to TimescaleDB")
	TimescaleDB = db
}

func InsertClickEventsBulk(events []*database.ClickEvent) *error {
	ctx := context.Background()
	queryInsertClickEventsData := `INSERT INTO click_events (url_id, geo, device, os, referrer, timestamp) VALUES ($1, $2, $3, $4, $5, $6);`

	connStr := os.Getenv("TIMESCALE_CONN")
	db, err := pgx.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return &err
	}

	//create batch
	batch := &pgx.Batch{}
	//load insert statements into batch queue
	for _, event := range events {
		batch.Queue(queryInsertClickEventsData, &event.URLId, &event.Geo, &event.Device, &event.OS, &event.Referrer, &event.Timestamp)
	}

	//send batch to connection pool
	clickEventsInsertBatch := db.SendBatch(ctx, batch)
	//execute statements in batch queue
	_, err = clickEventsInsertBatch.Exec()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to execute statement in batch queue for click events %v\n", err)
		return &err
	}
	fmt.Println("Successfully click events data batch inserted")

	return nil
}

func QueryClickEvents(urlId string, startTime int64, endTime int64) (*map[string]map[string]int, error) {
	ctx := context.Background()
	queryClickEventsData := `SELECT device, os, geo, referrer FROM click_events WHERE url_id = $1 AND timestamp >= to_timestamp($2) AND timestamp <= to_timestamp($3);`
	// queryClickEventsData := `SELECT COUNT(geo) AS geo_count, COUNT(device) AS device_count, COUNT(os) AS os_count, COUNT(referrer) AS referrer_count FROM click_events WHERE url_id = $1 AND timestamp >= to_timestamp($2) AND timestamp <= to_timestamp($3);`
	// queryClickEventsData := `SELECT
	// (SELECT json_agg(row_to_json(t)) FROM (SELECT device, COUNT(device) AS device_count FROM click_events WHERE url_id = $1 AND timestamp >= to_timestamp($2) AND timestamp <= to_timestamp($3) GROUP BY device) t) AS device_counts,
	// (SELECT json_agg(row_to_json(t)) FROM (SELECT os, COUNT(os) AS os_count FROM click_events WHERE url_id = $1 AND timestamp >= to_timestamp($2) AND timestamp <= to_timestamp($3) GROUP BY os) t) AS os_counts,
	// (SELECT json_agg(row_to_json(t)) FROM (SELECT geo, COUNT(geo) AS geo_count FROM click_events WHERE url_id = $1 AND timestamp >= to_timestamp($2) AND timestamp <= to_timestamp($3) GROUP BY geo) t) AS geo_counts,
	// (SELECT json_agg(row_to_json(t)) FROM (SELECT referrer, COUNT(referrer) AS referrer_count FROM click_events WHERE url_id = $1 AND timestamp >= to_timestamp($2) AND timestamp <= to_timestamp($3) GROUP BY referrer) t) AS referrer_counts;`

	rows, err := TimescaleDB.Query(ctx, queryClickEventsData, urlId, startTime, endTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to execute query for click events %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var clickEvents []*database.ClickEvent
	for rows.Next() {
		var device, os, geo, referrer string
		if err := rows.Scan(&device, &os, &geo, &referrer); err != nil {
			log.Fatalf("Failed to scan row: %v\n", err)
		}
		clickEvents = append(clickEvents, &database.ClickEvent{Device: device, OS: os, Geo: database.CountryName(geo), Referrer: referrer})
	}

	clickCounts := map[string]map[string]int{
		"device_counts":   make(map[string]int),
		"os_counts":       make(map[string]int),
		"geo_counts":      make(map[string]int),
		"referrer_counts": make(map[string]int),
	}

	for _, event := range clickEvents {
		clickCounts["device_counts"][event.Device]++
		clickCounts["os_counts"][event.OS]++
		clickCounts["geo_counts"][string(event.Geo)]++
		clickCounts["referrer_counts"][event.Referrer]++
		clickCounts["device_counts"]["total"]++
		clickCounts["os_counts"]["total"]++
		clickCounts["geo_counts"]["total"]++
		clickCounts["referrer_counts"]["total"]++
	}

	return &clickCounts, nil
}
