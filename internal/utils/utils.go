package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func RaToDegrees(ra string) (float64, error) {
	parts := strings.Split(ra, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid RA format, expected HH:MM:SS")
	}

	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid hours: %v", err)
	}

	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %v", err)
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid seconds: %v", err)
	}

	// Convert to decimal hours then to degrees
	decimalHours := hours + (minutes / 60) + (seconds / 3600)
	degrees := decimalHours * 15

	return degrees, nil
}

func DecToDegrees(dec string) (float64, error) {
	replacer := strings.NewReplacer("°", "", "'", "", "\"", "")
	decClean := replacer.Replace(dec)

	fields := strings.Fields(decClean)
	if len(fields) != 3 {
		return 0, fmt.Errorf("invalid declination format, expected D° M' S\"")
	}

	// Check if the degrees are negative
	sign := 1.0
	if strings.HasPrefix(fields[0], "-") {
		sign = -1.0
	}

	deg, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid degrees: %v", err)
	}

	min, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid arcminutes: %v", err)
	}

	sec, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid arcseconds: %v", err)
	}

	// Use absolute value for components, apply sign at the end
	absDeg := math.Abs(deg) + (min / 60) + (sec / 3600)
	return sign * absDeg, nil
}
