#!/usr/bin/env bash

set -e

# Cores para o output no terminal
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0;36m' # Reset / No Color

echo -e "${BLUE}=================================================="${NC}
echo -e "${BLUE}        ClipSync Daemon - Instalação Automática    "${NC}
echo -e "${BLUE}=================================================="${NC}

# Default server
DEFAULT_SERVER="https://wordpress.vps-kinghost.net"
SERVER_URL="${CLIPSYNC_SERVER:-$DEFAULT_SERVER}"

if [ -z "$CLIPSYNC_SERVER" ]; then
    read -p "Digite a URL do servidor ClipSync [padrão: $DEFAULT_SERVER]: " INPUT_SERVER
    if [ -n "$INPUT_SERVER" ]; then
        SERVER_URL="$INPUT_SERVER"
    fi
fi

echo -e "\n${YELLOW}-> Servidor configurado:${NC} $SERVER_URL"

# 2. Verificação do Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}[ERRO] Go não encontrado no sistema. Por favor, instale o Go para compilar o daemon.${NC}"
    exit 1
fi

# 3. Criação dos diretórios locais do usuário
BIN_DIR="$HOME/.local/bin"
SYSTEMD_DIR="$HOME/.config/systemd/user"

mkdir -p "$BIN_DIR"
mkdir -p "$SYSTEMD_DIR"

# 4. Compilação do Binário
echo -e "\n${YELLOW}-> Compilando...${NC}"
go build -o "$BIN_DIR/clipsync-daemon" .

chmod +x "$BIN_DIR/clipsync-daemon"
echo -e "${GREEN}✓ Binário gerado em:${NC} $BIN_DIR/clipsync-daemon"

# 5. Criação do Serviço no Systemd (User level)
SERVICE_FILE="$SYSTEMD_DIR/clipsync.service"

echo -e "\n${YELLOW}-> Criando serviço...${NC}"

cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=ClipSync Go Daemon
After=network.target

[Service]
ExecStart=$BIN_DIR/clipsync-daemon -server $SERVER_URL
Restart=always
RestartSec=5s

[Install]
WantedBy=default.target
EOF

echo -e "${GREEN}✓ Serviço Systemd criado em:${NC} $SERVICE_FILE"

# 6. Recarga e Ativação do Serviço
echo -e "\n${YELLOW}-> Ativando e iniciando o serviço...${NC}"
systemctl --user daemon-reload
systemctl --user enable clipsync.service
systemctl --user restart clipsync.service

echo -e "\n${GREEN}=================================================="${NC}
echo -e "${GREEN} Instalação concluída com sucesso! 🎉"${NC}
echo -e "${GREEN}=================================================="${NC}
echo -e "\nPara verificar o status do serviço e o PIN de pareamento (no 1º uso):"
echo -e "  ${YELLOW}journalctl --user -u clipsync -f${NC}\n"