package tools

import (
	"fmt"
	"strings"
	"time"
)

// Format Time with Timezone for IP Address
func LookupTimezone(givenTime time.Time, address string) string {
	style := "01/02/2006 03:04pm"
	loc := time.UTC
	if ok, result := GeolocateIPV4(address); ok {
		loc = time.FixedZone("UTC", int(result.TimezoneOffset))
	} else if ok, result := GeolocateIPV6(address); ok {
		loc = time.FixedZone("UTC", int(result.TimezoneOffset))
	}
	return givenTime.In(loc).Format(style)
}

// Determine Location (approximate) using IP Address
func LookupLocation(address string) string {
	if ok, result := GeolocateIPV4(address); ok {
		return fmt.Sprintf(
			"%s, %s, %s",
			GeolocateString(result.CityIndex),
			GeolocateString(result.RegionIndex),
			GeolocateString(result.CountryIndex),
		)
	}
	if ok, result := GeolocateIPV6(address); ok {
		return fmt.Sprintf(
			"%s, %s, %s",
			GeolocateString(result.CityIndex),
			GeolocateString(result.RegionIndex),
			GeolocateString(result.CountryIndex),
		)
	}
	return "Unavailable"
}

// Determine Device (approximate) using User Agent
func LookupBrowser(ua string) string {
	name := "Unknown Browser"
	switch {
	case strings.Contains(ua, "Chrome"):
		name = "Chrome"
	case strings.Contains(ua, "Firefox"):
		name = "Firefox"
	case strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome"):
		name = "Safari"
	case strings.Contains(ua, "Edge"):
		name = "Edge"
	}

	os := "Unknown OS"
	switch {
	case strings.Contains(ua, "Windows NT"):
		os = "Windows"
	case strings.Contains(ua, "Mac OS X"):
		os = "macOS"
	case strings.Contains(ua, "Android"):
		os = "Android"
	case strings.Contains(ua, "iPhone"), strings.Contains(ua, "iPad"), strings.Contains(ua, "iOS"):
		os = "iOS"
	case strings.Contains(ua, "Linux"):
		os = "Linux"
	}

	return fmt.Sprintf("%s on %s", name, os)
}
