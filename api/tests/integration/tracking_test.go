package integration_tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ivinayakg/shorte.live/api/database"
	"github.com/ivinayakg/shorte.live/api/timescale"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestURLRedirectTracking(t *testing.T) {
	resp, err := RedirecthttpClient.Get(ServerURL + "/" + URLFixture.Short)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	destinationURL := URLFixture.Destination

	assert.Equal(t, resp.StatusCode, http.StatusMovedPermanently, "Excpected status code to be 301")
	assert.Contains(t, resp.Header.Get("Location"), destinationURL, "Expected redirect to destination url")

	result := &database.ClickEvent{}

	for result.URLId == "" {
		clickEventSelectQuery := `SELECT url_id, geo, device, os, referrer, timestamp FROM click_events WHERE url_id = $1 LIMIT 1;`

		fmt.Println(clickEventSelectQuery, URLFixture.ID)
		err = timescale.TimescaleDB.QueryRow(context.TODO(), clickEventSelectQuery, URLFixture.ID.Hex()).Scan(&result.URLId, &result.Geo, &result.Device, &result.OS, &result.Referrer, &result.Timestamp)
		if err != nil && err != pgx.ErrNoRows {
			t.Log(err)
			t.Fail()
		}
		time.Sleep(time.Second * 2)
	}

	fmt.Println((*result).URLId == URLFixture.ID.Hex())

	assert.Equal(t, (*result).URLId, URLFixture.ID.Hex(), "Expected URL ID to be the same")
}
