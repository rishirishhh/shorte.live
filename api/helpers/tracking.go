package helpers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ivinayakg/shorte.live/api/database"
	"github.com/ivinayakg/shorte.live/api/timescale"
	"gopkg.in/mgo.v2/bson"
)

type TrackEventType string

const track_event_redis_key = "track_event"

type TrackerType struct {
	maxEvents      int
	eventsLength   int
	flushFrequency time.Duration
	mutex          sync.Mutex
}

var Tracker *TrackerType

func flushToDB(events []*database.ClickEvent) {
	if len(events) > 0 {
		if err := timescale.InsertClickEventsBulk(events); err != nil {
			log.Fatal(err)
		}
	}
}

func (eq *TrackerType) flush() {
	eq.mutex.Lock()
	defer eq.mutex.Unlock()

	jsondata, err := Redis.Client.LRange(context.TODO(), track_event_redis_key, 0, int64(eq.eventsLength)).Result()
	if err != nil {
		log.Fatal(err)
	}

	Redis.Client.LTrim(context.TODO(), track_event_redis_key, int64(eq.eventsLength), -1)
	eq.eventsLength = 0

	var data []*database.ClickEvent

	for _, v := range jsondata {
		var temp database.ClickEvent
		bson.Unmarshal([]byte(v), &temp)
		data = append(data, &temp)
	}

	go flushToDB(data)
}

func (eq *TrackerType) CaptureRedirectEvent(device string, ip string, os string, referrer string, urlId string, timestamp int64) {
	fmt.Println(ip)
	// add the ip-2-geo location handler here
	geo := "null"

	data := database.ClickEvent{URLId: urlId, Device: device, Geo: database.CountryName(geo), OS: os, Referrer: referrer, Timestamp: database.UnixTime(timestamp)}

	jsonData, _ := bson.Marshal(data)

	// Push the entire slice as a single element into the Redis list
	err := Redis.Client.LPush(context.TODO(), track_event_redis_key, jsonData).Err()
	if err != nil {
		log.Fatal(err)
	}

	eq.eventsLength += 1
	if eq.eventsLength >= eq.maxEvents {
		eq.flush()
	}
}

func (eq *TrackerType) StartFlush() {
	ticker := time.NewTicker(eq.flushFrequency)
	defer ticker.Stop()

	for range ticker.C {
		eq.flush()
	}
}

func SetupTracker(dur time.Duration, maxEvents int, eventsLength int) {
	Redis.Client.LTrim(context.TODO(), track_event_redis_key, int64(eventsLength), -1)
	Tracker = &TrackerType{
		maxEvents:      maxEvents,
		eventsLength:   eventsLength,
		flushFrequency: dur,
	}
}
