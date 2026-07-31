package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"golang.design/x/clipboard"
)

type WSMessage struct {
	Type           string `json:"type,omitempty"`
	Content        string `json:"content"`
	SenderDeviceID string `json:"sender_device_id,omitempty"`
	SenderDevice   string `json:"sender_device_name,omitempty"`
}

const MAX_RETRIES time.Duration = time.Duration(5)
const SYNC_TIME time.Duration = time.Duration(3)

func startSync(cfg *Config, db *DB) {
	err := clipboard.Init()
	if err != nil {
		log.Fatalf("Erro ao inicializar acesso ao Clipboard: %v", err)
	}

	u, err := url.Parse(cfg.ServerURL)
	if err != nil {
		log.Fatalf("URL invalida: %v", err)
	}

	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}

	wsURL := fmt.Sprintf("%s://%s/ws/clipboard/%s/%s", wsScheme, u.Host, cfg.UserID, cfg.DeviceID)

	for {
		log.Printf("Conectando ao WebSocket: %s\n", wsURL)
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			log.Printf("Erro na conexao WS: %v. Tentando novamente em 5s...", err)
			time.Sleep(MAX_RETRIES * time.Second)
			continue
		}

		log.Println("⚡ Conectado ao ClipSync Server!")

		// Channel para controlar cancelamento no disconnect
		ctx, cancel := context.WithCancel(context.Background())

		// Goroutine 1: Obtem mesgs do servidor, persiste no banco e atualiza tray
		go func() {
			defer cancel()
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("Conexao WS encerrada pelo servidor:", err)
					return
				}

				var msg WSMessage
				if err := json.Unmarshal(message, &msg); err == nil {
					if msg.Type == "clipboard_update" && msg.SenderDeviceID != cfg.DeviceID {
						fmt.Printf("📥 Recebido de [%s]: %s\n", msg.SenderDevice, msg.Content)
						clipboard.Write(clipboard.FmtText, []byte(msg.Content))
						db.Add(msg.Content)
						updateTrayMenu()
					}
				}
			}
		}()

		// Goroutine 2: Monitora o Clipboard Local -> Salva no SQLite -> Envia pro WS
		ch := clipboard.Watch(ctx, clipboard.FmtText)
		for data := range ch {
			text := strings.TrimSpace(string(data.Bytes))
			if text != "" {
				// Salva localmente
				db.Add(text)
				updateTrayMenu()

				payload := WSMessage{Content: text}
				payloadBytes, _ := json.Marshal(payload)

				err := conn.WriteMessage(websocket.TextMessage, payloadBytes)
				if err != nil {
					log.Println("Erro ao enviar mensagem pelo WS:", err)
					break
				}
				fmt.Printf("🚀 Enviado para a rede: %s\n", text)
			}
		}

		conn.Close()
		cancel()
		time.Sleep(SYNC_TIME * time.Second)
	}
}
