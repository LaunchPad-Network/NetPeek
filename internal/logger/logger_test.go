package logger

import (
	"testing"

	"github.com/LaunchPad-Network/NetPeek/internal/config"
)

func TestNew(t *testing.T) {
	config.Development()
	log := New("LogUnitTesting")
	log.WithField("test", "test").Debug("This is a debug log.")
	log.Info("This is an info log.")
	log.Warn("This is a warn log.")
	log.Error("This is an error log.")
	// log.Fatal("This is a fatal log.")
}

func BenchmarkNew(b *testing.B) {
	config.Development()
	log := New("LogUnitTesting")
	for i := 0; i < b.N; i++ {
		log.WithField("test", "test").Debug("This is a debug log.")
		log.Info("This is an info log.")
		log.Warn("This is a warn log.")
		log.Error("This is an error log.")
		//log.Fatal("This is a fatal log.")
	}
}
