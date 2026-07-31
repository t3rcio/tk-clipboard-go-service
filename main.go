package main

import (
	"flag"
	"log"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8000", "URL do servidor ClipSync")
	flag.Parse()

	cfg, err := loadConfig(*serverURL)
	if err != nil {
		log.Fatalf("Erro ao carregar configuracoes: %v", err)
	}

	db, err := initDB()
	if err != nil {
		log.Fatalf("Erro ao inicializar SQLite: %v", err)
	}
	defer db.Close()

	log.Println("Iniciando ClipSync Daemon...")

	if err := ensurePaired(cfg); err != nil {
		log.Fatalf("Erro durante o pareamento: %v", err)
	}

	// Inicia a sincronização do WebSocket em segundo plano
	go startSync(cfg, db)

	// Inicia a System Tray na thread principal
	setupTray(db)
}