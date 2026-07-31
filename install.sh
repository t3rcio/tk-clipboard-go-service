#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # Reset / No Color

REPO_URL="https://github.com/t3rcio/tk-clipboard-manager.git"
TMP_DIR=$(mktemp -d -t clipsync-build-XXXXXX)


cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo -e "${BLUE}=================================================="${NC}
echo -e "${BLUE}        ClipSync Daemon - Instalação Automática    "${NC}
echo -e "${BLUE}=================================================="${NC}

DEFAULT_SERVER="http://wordpress.vps-kinghost.net"
SERVER_URL="${CLIPSYNC_SERVER:-$DEFAULT_SERVER}"

if [ -z "$CLIPSYNC_SERVER" ]; then
    read -p "Digite a URL do servidor ClipSync [padrão: $DEFAULT_SERVER]: " INPUT_SERVER
    if [ -n "$INPUT_SERVER" ]; then
        SERVER_URL="$INPUT_SERVER"
    fi
fi

echo -e "\n${YELLOW}-> Servidor configurado:${NC} $SERVER_URL"

if ! command -v go &> /dev/null; then
    echo -e "${RED}[ERRO] Go não encontrado no sistema. Por favor, instale o Go 1.18+ para compilar o daemon.${NC}"
    exit 1
fi

if ! command -v git &> /dev/null; then
    echo -e "${RED}[ERRO] Git não encontrado. Instale o git para continuar a instalação.${NC}"
    exit 1
fi


BIN_DIR="$HOME/.local/bin"
SYSTEMD_DIR="$HOME/.config/systemd/user"

mkdir -p "$BIN_DIR"
mkdir -p "$SYSTEMD_DIR"


echo -e "\n${YELLOW}-> Obtendo o código-fonte...${NC}"
git clone --depth 1 "$REPO_URL" "$TMP_DIR/repo" >/dev/null 2>&1

# echo -e "${YELLOW}-> Compilando o binário Go...${NC}"
# cd "$TMP_DIR/repo"

echo -e "${YELLOW}-> Localizando módulo Go...${NC}"
GOMOD_PATH=$(find "$TMP_DIR/repo" -name "go.mod" -print -quit)

if [ -z "$GOMOD_PATH" ]; then
    echo -e "${RED}[ERRO] Não foi possível encontrar o arquivo go.mod no repositório.${NC}"
    exit 1
fi

GO_PROJECT_DIR=$(dirname "$GOMOD_PATH")
cd "$GO_PROJECT_DIR"

go build -o "$BIN_DIR/clipsync-daemon" .

chmod +x "$BIN_DIR/clipsync-daemon"
echo -e "${GREEN}✓ Binário gerado em:${NC} $BIN_DIR/clipsync-daemon"


SERVICE_FILE="$SYSTEMD_DIR/clipsync.service"

echo -e "\n${YELLOW}-> Criando serviço Systemd do usuário...${NC}"

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


echo -e "\n${YELLOW}-> Ativando e iniciando o serviço...${NC}"
systemctl --user daemon-reload
systemctl --user enable clipsync.service
systemctl --user restart clipsync.service

echo -e "\n${GREEN}=================================================="${NC}
echo -e "${GREEN} Instalação concluída com sucesso! 🎉"${NC}
echo -e "${GREEN}=================================================="${NC}
echo -e "\nPara verificar os logs e pegar o PIN de pareamento:"
echo -e "  ${YELLOW}journalctl --user -u clipsync -f${NC}\n"