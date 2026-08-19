package main

import (
	"fmt"
	"log"
	"time"

	"fyne.io/systray"
	"golang.design/x/clipboard"
)

const MAX_ITEMS_ON_TRAY = 10
const MAX_CHARS = 30

type ClipSyncMenuItem struct {
	mitem      *systray.MenuItem
	id         string
	text       string
	created_at time.Time
}

var (
	mHistoryTitle *systray.MenuItem
	mHistoryItems []*ClipSyncMenuItem
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
	systray.SetTooltip("ClipSync Daemon")

	mHistoryTitle = systray.AddMenuItem("📋 Histórico Recente", "Lista de itens copiados")
	mHistoryTitle.Disable()
	systray.AddSeparator()

	mHistoryItems = make([]*ClipSyncMenuItem, MAX_ITEMS_ON_TRAY)
	for i := 0; i < MAX_ITEMS_ON_TRAY; i++ {
		mHistoryItems[i] = &ClipSyncMenuItem{}
		_item := systray.AddMenuItem("", "")
		mHistoryItems[i].mitem = _item
		mHistoryItems[i].mitem.Hide()
		mHistoryItems[i].text = ""

		// MenuItem actions
		go func(i int, menuItem *ClipSyncMenuItem) {
			for range menuItem.mitem.ClickedCh {
				clipboard.Write(clipboard.FmtText, []byte(menuItem.text))
				log.Printf("Copiado do histórico da bandeja: %s\n", menuItem.text)
			}
		}(i, mHistoryItems[i])
	}

	systray.AddSeparator()
	mClearHistory = systray.AddMenuItem("🗑️ Limpar Histórico", "Apaga o histórico local")
	mQuit = systray.AddMenuItem("❌ Sair", "Encerra o ClipSync")

	// MenuItem ordinary actions
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

	items, err := dbRef.List(MAX_ITEMS_ON_TRAY)
	if err != nil {
		log.Println("Erro ao ler histórico para a bandeja:", err)
		return
	}

	for i := 0; i < MAX_ITEMS_ON_TRAY; i++ {
		if i < len(items) {
			itemText := items[i].Content
			preview := itemText

			if len(preview) > MAX_CHARS {
				preview = preview[:MAX_CHARS] + "..."
			}
			mHistoryItems[i].text = items[i].Content
			mHistoryItems[i].created_at = items[i].CreatedAt

			mHistoryItems[i].mitem.SetTitle(fmt.Sprintf("%d. %s", i+1, preview))
			mHistoryItems[i].mitem.SetTooltip(itemText)
			mHistoryItems[i].mitem.Show()

			// Trata clique no item
			// go func(menuItem *systray.MenuItem, text string) {
			// 	for range menuItem.ClickedCh {
			// 		clipboard.Write(clipboard.FmtText, []byte(text))
			// 		log.Printf("Copiado do histórico da bandeja: %s\n", text)
			// 	}
			// }(mHistoryItems[i], itemText)

		} else {
			mHistoryItems[i].mitem.Hide()
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
