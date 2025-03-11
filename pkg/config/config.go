package config

import "os"

var LogLevel = os.Getenv("LOG_LEVEL")
var InjectNewHeader = os.Getenv("NEW_HEADER")
