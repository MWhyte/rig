package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

// copyNoticeDuration is how long the "Copied" confirmation stays on screen.
const copyNoticeDuration = 2 * time.Second

// copyNotice tracks the outcome of the most recent clipboard copy so the UI
// can briefly confirm it. seq increments on every copy, so an expiry from an
// earlier press never clears a newer notice.
type copyNotice struct {
	visible bool
	err     error
	seq     int
}

// copyNoticeExpiredMsg hides the notice shown by the copy with the same seq.
type copyNoticeExpiredMsg struct{ seq int }

// copy writes text to the clipboard, records the outcome and returns a
// command that hides the notice after copyNoticeDuration.
func (n *copyNotice) copy(text string) tea.Cmd {
	n.seq++
	n.visible = true
	n.err = clipboard.WriteAll(text)

	seq := n.seq
	return tea.Tick(copyNoticeDuration, func(time.Time) tea.Msg {
		return copyNoticeExpiredMsg{seq: seq}
	})
}

// expire hides the notice if it still belongs to the copy identified by seq.
func (n *copyNotice) expire(seq int) {
	if seq == n.seq {
		n.visible = false
	}
}

// reset clears the notice, e.g. when its modal closes.
func (n *copyNotice) reset() {
	n.visible = false
	n.err = nil
}
