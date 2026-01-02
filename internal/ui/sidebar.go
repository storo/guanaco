package ui

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/store"
)

// Sidebar displays the list of chats.
type Sidebar struct {
	*gtk.Box

	listBox  *gtk.ListBox
	scrolled *gtk.ScrolledWindow
	chats    []*store.Chat

	// Dependencies
	db *store.DB

	// Callbacks
	onChatSelected func(*store.Chat)
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

	title := gtk.NewLabel("Chats")
	title.AddCSSClass("title-3")
	title.SetHExpand(true)
	title.SetXAlign(0)
	header.Append(title)

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

	// Add chat rows
	for _, chat := range chats {
		row := sb.createChatRow(chat)
		sb.listBox.Append(row)
	}
}

func (sb *Sidebar) createChatRow(chat *store.Chat) *gtk.ListBoxRow {
	row := gtk.NewListBoxRow()

	box := gtk.NewBox(gtk.OrientationVertical, 4)
	box.SetMarginTop(8)
	box.SetMarginBottom(8)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	// Title
	titleLabel := gtk.NewLabel(chat.Title)
	titleLabel.SetXAlign(0)
	titleLabel.SetEllipsize(3) // PANGO_ELLIPSIZE_END
	titleLabel.AddCSSClass("heading")
	box.Append(titleLabel)

	// Model subtitle
	modelLabel := gtk.NewLabel(chat.Model)
	modelLabel.SetXAlign(0)
	modelLabel.AddCSSClass("dim-label")
	modelLabel.AddCSSClass("caption")
	box.Append(modelLabel)

	row.SetChild(box)
	return row
}

// AddChat adds a new chat to the list.
func (sb *Sidebar) AddChat(chat *store.Chat) {
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
