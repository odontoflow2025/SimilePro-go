package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// TimezoneInjector extracts X-Timezone from the request and assigns the Location object into the Gin context
func TimezoneInjector() gin.HandlerFunc {
	return func(c *gin.Context) {
		tzHeader := c.GetHeader("X-Timezone")
		if tzHeader == "" {
			// Fallback to UTC if client doesn't send anything
			tzHeader = "UTC"
		}

		location, err := time.LoadLocation(tzHeader)
		if err != nil {
			log.Printf("[WARNING] Invalid timezone '%s' received, falling back to UTC. Error: %v\n", tzHeader, err)
			location = time.UTC
		}

		// Set the location object tightly into the request's context
		c.Set("timezone_location", location)

		c.Next()
	}
}

// GetLocationFromContext cleanly fetches the parsed timezone or defaults to UTC
func GetLocationFromContext(c *gin.Context) *time.Location {
	locVal, exists := c.Get("timezone_location")
	if !exists {
		return time.UTC
	}
	
	loc, ok := locVal.(*time.Location)
	if !ok {
		return time.UTC
	}
	return loc
}
