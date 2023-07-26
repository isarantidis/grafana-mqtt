package plugin

import (
	"strconv"
	"strings"
	"time"
)

/**
 */
func NewSubscriptionModel(path string) (SubscriptionModel, error) {
	// Split the path into elements
	var err error
	result := SubscriptionModel{}
	elements := strings.Split(path, ".")
	for _, str := range elements {
		pair := strings.Split(str, "=")
		if len(pair) == 2 {
			key := pair[0]
			value := pair[1]
			switch key {
			case "topic":
				result.topic = value
			case "useInterval":
				result.useInterval, err = getBooleanFromString(value)
			case "includeSchema":
				result.includeSchema, err = getBooleanFromString(value)
			case "interval":
				var interval time.Duration
				interval, err = time.ParseDuration(value)
				result.interval = int(interval.Milliseconds())
			}
		}
	}
	return result, err
}

func getBooleanFromString(value string) (bool, error) {
	bvalue, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}
	return bvalue, nil
}
