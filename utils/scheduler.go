package utils

import (
	"context"
	"event-horizon/store"
	"log"
	"time"
)

/** *********************  EVENT CLEANUP SCHEDULER   ********************

This scheduler runs in the background and periodically deletes expired events
from the database to keep it clean and efficient.


 **************************************/

// StartEventCleanupScheduler starts a background job that deletes expired events periodically
func StartEventCleanupScheduler(eventStore *store.EventStore) {

	//! Run  every hour
	ticker := time.NewTicker(1 * time.Hour)

	//! RUN IN CONCURRENT GO ROUTINE
	go func() {
		//! Run on startup
		runCleanup(eventStore)

		//! run periodically
		for range ticker.C {
			runCleanup(eventStore)
		}
	}()

	log.Println("EVENT CLEANUP STARTED")
}

// ! CLEAN UP FUNCTION
func runCleanup(eventStore *store.EventStore) {
	ctx := context.Background()
	deletedCount, err := eventStore.DeleteExpiredEvents(ctx)

	if err != nil {
		log.Printf("Error cleaning up expired events: %v", err)
		return
	}

	if deletedCount > 0 {
		log.Printf("Successfully deleted %d expired event(s)", deletedCount)
	}
}
