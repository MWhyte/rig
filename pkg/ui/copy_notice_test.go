package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrwhyte/rig/pkg/identifier"
)

func TestCopyNoticeExpireIgnoresStaleSeq(t *testing.T) {
	n := copyNotice{visible: true, seq: 2}

	n.expire(1)
	if !n.visible {
		t.Fatal("expiry from an earlier copy hid the current notice")
	}

	n.expire(2)
	if n.visible {
		t.Fatal("expiry for the current copy did not hide the notice")
	}
}

func TestRenderIdentifyFooter(t *testing.T) {
	tests := []struct {
		name   string
		notice copyNotice
		want   string
	}{
		{name: "hints by default", notice: copyNotice{}, want: "y: copy"},
		{name: "confirms a copy", notice: copyNotice{visible: true}, want: "Copied to clipboard"},
		{name: "reports a failure", notice: copyNotice{visible: true, err: errors.New("no clipboard")}, want: "Couldn't copy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{
				identifyTrack: &identifier.Track{Title: "Song", Artist: "Band"},
				identifyCopy:  tt.notice,
			}
			if got := m.renderIdentifyFooter(); !strings.Contains(got, tt.want) {
				t.Errorf("footer = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}
