package ui

import (
	"regexp"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/richbl/go-ble-sync-cycle/internal/logger"
)

// Regex to find ANSI escape sequences
var ansiSplitRegex = regexp.MustCompile(`(\x1b\[[0-9;]*m)`)

// Map ANSI codes to human-readable GTK Tag names
var ansiToTag = map[string]string{
	"\x1b[31m": "red",     // Error
	"\x1b[32m": "green",   // Info
	"\x1b[33m": "yellow",  // Warn
	"\x1b[34m": "blue",    // Component
	"\x1b[35m": "magenta", // Fatal
	"\x1b[36m": "cyan",    // Debug
	"\x1b[37m": "white",   // Normal/Time
	"\x1b[0m":  "reset",   // Reset
}

// GuiLogWriter implements io.Writer to bridge application logs to a GTK TextView
type GuiLogWriter struct {
	textView *gtk.TextView
	buffer   *gtk.TextBuffer
}

// setupSessionLogSignals wires up event listeners for the Session Log view (Page 3)
func (sc *SessionController) setupSessionLogSignals() {
	logger.Debug(logger.GUI, "Session Log: signals setup complete")
}

// UpdateLogLevel updates the log level component in the view
func (sc *SessionController) UpdateLogLevel() {
	sc.UI.Page3.LogLevelRow.SetTitle(logger.LogLevel())
}

// NewGuiLogWriter creates a new writer for the specified TextView and initializes color tags
func NewGuiLogWriter(tv *gtk.TextView) *GuiLogWriter {

	w := &GuiLogWriter{
		textView: tv,
		buffer:   tv.Buffer(),
	}
	w.initTags()

	return w
}

// Write satisfies the io.Writer interface, parsing ANSI codes and inserts styled text
func (w *GuiLogWriter) Write(p []byte) (int, error) {

	fullText := string(p)

	glib.IdleAdd(func() {
		w.processAnsiAndInsert(fullText)
	})

	return len(p), nil
}

// processAnsiAndInsert parses the text for ANSI codes and inserts into the buffer
func (w *GuiLogWriter) processAnsiAndInsert(text string) {

	// Get the end iterator
	endIter := w.buffer.EndIter()
	var currentTag *gtk.TextTag

	matchesIdx := ansiSplitRegex.FindAllStringIndex(text, -1)
	cursor := 0

	// Iterate over matches
	for _, loc := range matchesIdx {

		start, end := loc[0], loc[1]
		code := text[start:end]

		// Insert text segment before the code
		if start > cursor {
			w.insertWithTag(endIter, text[cursor:start], currentTag)
		}

		// Update Tag
		tagName, ok := ansiToTag[code]
		if ok && tagName != "reset" {
			currentTag = w.buffer.TagTable().Lookup(tagName)
		} else {
			currentTag = nil
		}

		cursor = end
	}

	// Insert remaining text
	if cursor < len(text) {
		w.insertWithTag(endIter, text[cursor:], currentTag)
	}

}

// insertWithTag inserts text into the buffer with the specified tag
func (w *GuiLogWriter) insertWithTag(iter *gtk.TextIter, text string, tag *gtk.TextTag) {

	if tag == nil {
		w.buffer.Insert(iter, text)

		return
	}

	startOffset := iter.Offset()
	w.buffer.Insert(iter, text)
	startIter := w.buffer.IterAtOffset(startOffset)
	w.buffer.ApplyTag(tag, startIter, iter)

}

// initTags defines the color styles in the buffer's tag table
func (w *GuiLogWriter) initTags() {

	// Get the tag table
	table := w.buffer.TagTable()

	// Helper to create tags if they don't exist
	createTag := func(name, colorHex string) {
		if table.Lookup(name) == nil {
			tag := gtk.NewTextTag(name)
			tag.SetObjectProperty("foreground", colorHex)
			table.Add(tag)
		}
	}

	// Pretty colors
	createTag("red", "#ff5555")
	createTag("green", "#50fa7b")
	createTag("yellow", "#f1fa8c")
	createTag("blue", "#8be9fd")
	createTag("magenta", "#ff79c6")
	createTag("cyan", "#8be9fd")
	createTag("white", "#f8f8f2")

}
