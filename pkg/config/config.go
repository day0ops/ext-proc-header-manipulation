package config

import "os"

var LogLevel = os.Getenv("LOG_LEVEL")
