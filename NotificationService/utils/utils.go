package utils

import (
	"time"
)

var ISTLocation *time.Location
func TimeStringToUnix(timeStr string) (int64, error) {
    // Define a custom layout that matches the input time string format
    const layout = "2006-01-02 15:04:05"
    t, err := time.ParseInLocation(layout, timeStr, ISTLocation)
    if err != nil {
        return 0, err
    }
    return t.Unix(), nil
}
