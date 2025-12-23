package config

import (
	"io"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)


func InitLogger() {
	logRotator := &lumberjack.Logger{
		Filename:   "logs/app.log",   // Nama file log
		MaxSize:    10,          // 10 Megabytes sebelum dipotong
		MaxBackups: 5,           // Simpan 5 file backup lama
		MaxAge:     30,          // Hapus backup yang lebih tua dari 30 hari
		Compress:   true,        // Compress jadi .gz
	}

	multiWriter := io.MultiWriter(os.Stdout, logRotator)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}