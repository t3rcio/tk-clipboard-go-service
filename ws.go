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

func startSync(cfg *Config) {
	// Inicializa o acesso ao Clipboard nativo do SO
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

		// Goroutine 1: Escuta mensagens do Servidor
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
					}
				}
			}
		}()

		// Goroutine 2: Escuta alterações no Clipboard do SO local
		ch := clipboard.Watch(ctx, clipboard.FmtText)
		for data := range ch {
			text := strings.TrimSpace(string(data.Bytes))
			if text != "" {
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