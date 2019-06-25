package logger

import "testing"

func TestLogger(t *testing.T) {
	log := Get("test")
	log.Info("info")
	log.Warn("info")
	log.Error("info")
}
