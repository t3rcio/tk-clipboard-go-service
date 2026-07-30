package main

import (
	"flag"
	"log"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8000", "URL do servidor")
	flag.Parse()

	cfg, err := loadConfig(*serverURL)
	if err != nil {
		log.Fatalf("Erro ao carregar configuracoes %v", err)
	}

	log.Println("Iniciando Daemon...")
	if err := ensurePaired(cfg); err != nil {
		log.Fatalf("Erro durante pareamento %v", err)
	}

	startSync(cfg)
}