package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type DeviceCodeRequest struct {
	DeviceName string `json:"device_name"`
}

type DeviceCodeResponse struct {
	Code string `json:"code"`
	ExpiresIn int `json:"expires_in"`
}

type PairStatusResponse struct {
	Status   string `json:"status"`
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
}

func ensurePaired(cfg *Config) error {
	if cfg.UserID != "" && cfg.DeviceID != "" {
		return nil
	}

	hostname, _ := os.Hostname()
	deviceName := fmt.Sprintf("Go Daemon (%s)", hostname)

	reqBody, _ := json.Marshal(DeviceCodeRequest{DeviceName: deviceName})
	resp, err := http.Post(cfg.ServerURL+"/api/auth/device-code", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("Falha ao conectar ao servidor: %w", err)
	}
	defer resp.Body.Close()

	var codeRes DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&codeRes); err != nil {
		return fmt.Errorf("Resposta invalida do servidor: %w", err)
	}

	fmt.Printf("🔑 CÓDIGO DE PAREAMENTO:  [ %s ]\n", codeRes.Code)	
	fmt.Println("1. Abra o PWA do ClipSync no navegador/smartphone.")
	fmt.Println("2. Digite o PIN acima no campo 'Parear Computador'.")
	fmt.Println("Aguardando autorização...")

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		statusURL := fmt.Sprintf("%s/api/auth/device-code/%s/status", cfg.ServerURL, codeRes.Code)
		resp, err := http.Get(statusURL)
		if err != nil {
			continue
		}

		var statusRes PairStatusResponse
		json.NewDecoder(resp.Body).Decode(&statusRes)
		resp.Body.Close()

		if statusRes.Status == "approved" {
			cfg.UserID = statusRes.UserID
			cfg.DeviceID = statusRes.DeviceID
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("erro ao salvar configuracao: %w", err)
			}
			fmt.Println("\n✅ Dispositivo pareado com sucesso!")
			return nil
		}
	}

	return nil
}