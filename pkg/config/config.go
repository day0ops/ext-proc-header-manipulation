package config

import "os"

var LogLevel = os.Getenv("LOG_LEVEL")
var PodName = os.Getenv("POD_NAME")
