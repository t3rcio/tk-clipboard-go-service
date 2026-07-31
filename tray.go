package main

import (
	"fmt"
	"log"

	"fyne.io/systray"
	"golang.design/x/clipboard"
)

var (
	mHistoryTitle *systray.MenuItem
	mHistoryItems []*systray.MenuItem
	mClearHistory *systray.MenuItem
	mQuit         *systray.MenuItem
	dbRef         *DB
)

func setupTray(db *DB) {
	dbRef = db
	systray.Run(onReady, onExit)
}

func onReady() {
	// Ícone genérico de prancheta (PNG simplificado)
	systray.SetTemplateIcon(iconData, iconData)
	systray.SetTitle("ClipSync")
	systray.SetTooltip("ClipSync Clipboard Manager")

	mHistoryTitle = systray.AddMenuItem("📋 Histórico Recente", "Lista de itens copiados")
	mHistoryTitle.Disable()
	systray.AddSeparator()

	// Cria slots para os 10 últimos itens
	mHistoryItems = make([]*systray.MenuItem, 10)
	for i := 0; i < 10; i++ {
		mHistoryItems[i] = systray.AddMenuItem("", "")
		mHistoryItems[i].Hide()
	}

	systray.AddSeparator()
	mClearHistory = systray.AddMenuItem("🗑️ Limpar Histórico", "Apaga o histórico local")
	mQuit = systray.AddMenuItem("❌ Sair", "Encerra o ClipSync")

	// Goroutine para escutar ações nos menus
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			case <-mClearHistory.ClickedCh:
				if err := dbRef.Clear(); err == nil {
					updateTrayMenu()
				}
			}
		}
	}()

	updateTrayMenu()
}

func updateTrayMenu() {
	if dbRef == nil {
		return
	}

	items, err := dbRef.List(10)
	if err != nil {
		log.Println("Erro ao ler histórico para a bandeja:", err)
		return
	}

	for i := 0; i < 10; i++ {
		if i < len(items) {
			itemText := items[i].Content
			preview := itemText
			if len(preview) > 30 {
				preview = preview[:30] + "..."
			}

			mHistoryItems[i].SetTitle(fmt.Sprintf("%d. %s", i+1, preview))
			mHistoryItems[i].SetTooltip(itemText)
			mHistoryItems[i].Show()

			// Trata clique no item
			go func(menuItem *systray.MenuItem, text string) {
				for range menuItem.ClickedCh {
					clipboard.Write(clipboard.FmtText, []byte(text))
					log.Printf("Copiado do histórico da bandeja: %s\n", text)
				}
			}(mHistoryItems[i], itemText)

		} else {
			mHistoryItems[i].Hide()
		}
	}
}

func onExit() {
	// Cleanup ao fechar
}

// Ícone mínimo de backup em byte array (clipboard icon)
var iconData = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0xf3, 0xff, 0x61, 0x00, 0x00, 0x00,
	0x19, 0x49, 0x44, 0x41, 0x54, 0x38, 0x8d, 0x63, 0x60, 0x18, 0x05, 0xa3,
	0x60, 0x14, 0x8c, 0x82, 0x51, 0x30, 0x0a, 0x61, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x03, 0x00, 0x18, 0xdd, 0x02, 0x21, 0x98, 0xd0, 0x9e, 0xdf, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}