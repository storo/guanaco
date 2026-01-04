package ui

import (
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/i18n"
	"github.com/storo/guanaco/internal/logger"
	"github.com/storo/guanaco/internal/store"
)

// Sidebar displays the list of chats.
type Sidebar struct {
	*gtk.Box

	listBox       *gtk.ListBox
	scrolled      *gtk.ScrolledWindow
	emptyState    *gtk.Box
	newChatButton *gtk.Button
	chats         []*store.Chat

	// Dependencies
	db     *store.DB
	window *gtk.Window

	// Callbacks
	onChatSelected func(*store.Chat)
	onChatDeleted  func(int64)
	onSettings     func()
}

// NewSidebar creates a new sidebar.
func NewSidebar(db *store.DB) *Sidebar {
	sb := &Sidebar{
		db: db,
	}

	sb.Box = gtk.NewBox(gtk.OrientationVertical, 0)
	sb.SetVExpand(true)

	sb.setupUI()

	return sb
}

func (sb *Sidebar) setupUI() {
	// Header
	header := gtk.NewBox(gtk.OrientationHorizontal, 8)
	header.SetMarginTop(12)
	header.SetMarginBottom(12)
	header.SetMarginStart(12)
	header.SetMarginEnd(12)

	title := gtk.NewLabel(i18n.T("Chats"))
	title.AddCSSClass("title-3")
	title.SetHExpand(true)
	title.SetXAlign(0)
	header.Append(title)

	// New Chat button
	sb.newChatButton = gtk.NewButton()
	sb.newChatButton.SetIconName("list-add-symbolic")
	sb.newChatButton.SetTooltipText(i18n.T("New Chat"))
	sb.newChatButton.AddCSSClass("flat")
	header.Append(sb.newChatButton)

	sb.Append(header)

	// Separator
	separator := gtk.NewSeparator(gtk.OrientationHorizontal)
	sb.Append(separator)

	// Chat list
	sb.listBox = gtk.NewListBox()
	sb.listBox.SetSelectionMode(gtk.SelectionSingle)
	sb.listBox.AddCSSClass("navigation-sidebar")
	sb.listBox.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		if row == nil {
			return
		}

		idx := row.Index()
		if idx >= 0 && idx < len(sb.chats) {
			if sb.onChatSelected != nil {
				sb.onChatSelected(sb.chats[idx])
			}
		}
	})

	sb.scrolled = gtk.NewScrolledWindow()
	sb.scrolled.SetChild(sb.listBox)
	sb.scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	sb.scrolled.SetVExpand(true)
	sb.Append(sb.scrolled)

	// Empty state (hidden by default)
	sb.emptyState = gtk.NewBox(gtk.OrientationVertical, 8)
	sb.emptyState.SetVExpand(true)
	sb.emptyState.SetVAlign(gtk.AlignCenter)
	sb.emptyState.SetMarginStart(16)
	sb.emptyState.SetMarginEnd(16)

	emptyIcon := gtk.NewImageFromIconName("chat-message-new-symbolic")
	emptyIcon.SetIconSize(gtk.IconSizeLarge)
	emptyIcon.AddCSSClass("dim-label")
	sb.emptyState.Append(emptyIcon)

	emptyTitle := gtk.NewLabel(i18n.T("No conversations yet"))
	emptyTitle.AddCSSClass("dim-label")
	sb.emptyState.Append(emptyTitle)

	emptyDesc := gtk.NewLabel(i18n.T("Start a new chat to begin"))
	emptyDesc.AddCSSClass("dim-label")
	emptyDesc.AddCSSClass("caption")
	sb.emptyState.Append(emptyDesc)

	sb.emptyState.SetVisible(false)
	sb.Append(sb.emptyState)

	// === FOOTER ===
	footerSeparator := gtk.NewSeparator(gtk.OrientationHorizontal)
	sb.Append(footerSeparator)

	footer := gtk.NewBox(gtk.OrientationVertical, 4)
	footer.SetMarginTop(8)
	footer.SetMarginBottom(8)
	footer.SetMarginStart(8)
	footer.SetMarginEnd(8)

	// Settings button
	settingsBtn := gtk.NewButton()
	settingsBtn.SetChild(sb.createFooterButtonContent("preferences-system-symbolic", i18n.T("Settings")))
	settingsBtn.AddCSSClass("flat")
	settingsBtn.ConnectClicked(func() {
		if sb.onSettings != nil {
			sb.onSettings()
		}
	})
	footer.Append(settingsBtn)

	sb.Append(footer)
}

// createFooterButtonContent creates a horizontal box with icon and label for footer buttons.
func (sb *Sidebar) createFooterButtonContent(iconName, label string) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)

	icon := gtk.NewImageFromIconName(iconName)
	box.Append(icon)

	labelWidget := gtk.NewLabel(label)
	labelWidget.SetHExpand(true)
	labelWidget.SetXAlign(0)
	box.Append(labelWidget)

	return box
}

// LoadChats loads and displays chats from the database.
func (sb *Sidebar) LoadChats() {
	if sb.db == nil {
		return
	}

	chats, err := sb.db.ListChats()
	if err != nil {
		return
	}

	sb.setChats(chats)
}

func (sb *Sidebar) setChats(chats []*store.Chat) {
	// Clear existing
	for {
		row := sb.listBox.RowAtIndex(0)
		if row == nil {
			break
		}
		sb.listBox.Remove(row)
	}

	sb.chats = chats

	// Show/hide empty state
	hasChats := len(chats) > 0
	sb.scrolled.SetVisible(hasChats)
	sb.emptyState.SetVisible(!hasChats)

	// Add chat rows
	for _, chat := range chats {
		row := sb.createChatRow(chat)
		sb.listBox.Append(row)
	}
}

func (sb *Sidebar) createChatRow(chat *store.Chat) *gtk.ListBoxRow {
	row := gtk.NewListBoxRow()

	box := gtk.NewBox(gtk.OrientationVertical, 2)
	box.SetMarginTop(8)
	box.SetMarginBottom(8)
	box.SetMarginStart(12)
	box.SetMarginEnd(8)

	// Header with title and delete button
	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 4)

	// Title
	titleLabel := gtk.NewLabel(chat.Title)
	titleLabel.SetXAlign(0)
	titleLabel.SetHExpand(true)
	titleLabel.SetEllipsize(3) // PANGO_ELLIPSIZE_END
	titleLabel.AddCSSClass("heading")
	headerBox.Append(titleLabel)

	// Delete button
	deleteBtn := gtk.NewButton()
	deleteBtn.SetIconName("user-trash-symbolic")
	deleteBtn.AddCSSClass("flat")
	deleteBtn.AddCSSClass("circular")
	deleteBtn.SetTooltipText(i18n.T("Delete chat"))
	deleteBtn.SetVAlign(gtk.AlignCenter)

	chatID := chat.ID // capture for closure
	deleteBtn.ConnectClicked(func() {
		sb.deleteChat(chatID)
	})
	headerBox.Append(deleteBtn)

	box.Append(headerBox)

	// Preview of last message
	if sb.db != nil {
		if messages, err := sb.db.GetMessages(chat.ID); err == nil && len(messages) > 0 {
			lastMsg := messages[len(messages)-1]
			preview := truncatePreview(lastMsg.Content, 40)

			previewLabel := gtk.NewLabel(preview)
			previewLabel.SetXAlign(0)
			previewLabel.SetEllipsize(3) // PANGO_ELLIPSIZE_END
			previewLabel.AddCSSClass("dim-label")
			previewLabel.AddCSSClass("caption")
			box.Append(previewLabel)
		}
	}

	// Model subtitle (smaller, dimmer)
	modelLabel := gtk.NewLabel(chat.Model)
	modelLabel.SetXAlign(0)
	modelLabel.AddCSSClass("dim-label")
	modelLabel.AddCSSClass("caption")
	modelLabel.SetOpacity(0.6)
	box.Append(modelLabel)

	row.SetChild(box)
	return row
}

// truncatePreview truncates text for preview display.
func truncatePreview(s string, maxLen int) string {
	// Remove newlines for preview
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}

// AddChat adds a new chat to the list if not already present.
func (sb *Sidebar) AddChat(chat *store.Chat) {
	// Check if chat already exists
	for _, c := range sb.chats {
		if c.ID == chat.ID {
			return // Already in list
		}
	}

	sb.chats = append([]*store.Chat{chat}, sb.chats...)
	row := sb.createChatRow(chat)
	sb.listBox.Prepend(row)
}

// SelectChat selects a chat in the list.
func (sb *Sidebar) SelectChat(chat *store.Chat) {
	for i, c := range sb.chats {
		if c.ID == chat.ID {
			row := sb.listBox.RowAtIndex(i)
			if row != nil {
				sb.listBox.SelectRow(row)
			}
			break
		}
	}
}

// OnChatSelected sets the callback for when a chat is selected.
func (sb *Sidebar) OnChatSelected(callback func(*store.Chat)) {
	sb.onChatSelected = callback
}

// Refresh reloads the chat list.
func (sb *Sidebar) Refresh() {
	sb.LoadChats()
}

// OnNewChat sets the callback for when the new chat button is clicked.
func (sb *Sidebar) OnNewChat(callback func()) {
	sb.newChatButton.ConnectClicked(callback)
}

// OnChatDeleted sets the callback for when a chat is deleted.
func (sb *Sidebar) OnChatDeleted(callback func(int64)) {
	sb.onChatDeleted = callback
}

// deleteChat shows a confirmation dialog and deletes a chat if confirmed.
func (sb *Sidebar) deleteChat(chatID int64) {
	if sb.db == nil {
		return
	}

	// Create confirmation dialog
	dialog := adw.NewMessageDialog(sb.window, i18n.T("Delete Chat?"), i18n.T("This conversation will be permanently deleted. This action cannot be undone."))
	dialog.AddResponse("cancel", i18n.T("Cancel"))
	dialog.AddResponse("delete", i18n.T("Delete"))
	dialog.SetResponseAppearance("delete", adw.ResponseDestructive)
	dialog.SetDefaultResponse("cancel")
	dialog.SetCloseResponse("cancel")

	dialog.ConnectResponse(func(response string) {
		if response == "delete" {
			sb.confirmDeleteChat(chatID)
		}
	})

	dialog.Present()
}

// confirmDeleteChat actually deletes the chat after confirmation.
func (sb *Sidebar) confirmDeleteChat(chatID int64) {
	if err := sb.db.DeleteChat(chatID); err != nil {
		logger.Error("Failed to delete chat", "chatID", chatID, "error", err)
		return
	}

	logger.Info("Chat deleted", "chatID", chatID)

	// Notify listener
	if sb.onChatDeleted != nil {
		sb.onChatDeleted(chatID)
	}

	// Refresh the list
	sb.Refresh()
}

// OnSettings sets the callback for when the settings button is clicked.
func (sb *Sidebar) OnSettings(callback func()) {
	sb.onSettings = callback
}

// SetWindow sets the parent window reference for dialogs.
func (sb *Sidebar) SetWindow(window *gtk.Window) {
	sb.window = window
}
