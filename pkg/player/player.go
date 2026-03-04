package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// State represents the current player state
type State int

const (
	StateStopped State = iota
	StatePlaying
	StatePaused
)

// MPVResponse represents a response from MPV IPC
type MPVResponse struct {
	RequestID *int        `json:"request_id"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error"`
}

// Metadata represents stream metadata
type Metadata struct {
	Title      string  // Current song (ICY title)
	Genre      string  // ICY genre
	BufferSecs float64 // Seconds of buffered audio ahead of playback position
	ActualKbps float64 // Actual decoded audio bitrate in kbps
}

// Player interface defines the audio player operations
type Player interface {
	Play(url string) error
	Pause() error
	Resume() error
	Stop() error
	SetVolume(volume int) error
	GetVolume() (int, error)
	GetState() State
	IsPlaying() bool
	GetMetadata() (*Metadata, error)
	Close() error
}

// MPVPlayer is an mpv-based audio player
type MPVPlayer struct {
	cmd        *exec.Cmd
	stdout     io.ReadCloser
	socketPath string
	state      State
	currentURL string
	volume     int
	mu         sync.RWMutex
}

// NewMPVPlayer creates a new mpv-based player
func NewMPVPlayer() (*MPVPlayer, error) {
	// Check if mpv is available
	if _, err := exec.LookPath("mpv"); err != nil {
		return nil, fmt.Errorf("mpv not found in PATH: %w (install with: brew install mpv)", err)
	}

	// Create socket path in temp directory
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("rig-mpv-%d.sock", os.Getpid()))

	return &MPVPlayer{
		socketPath: socketPath,
		state:      StateStopped,
		volume:     70,
	}, nil
}

// Play starts playing the given URL
func (p *MPVPlayer) Play(url string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop current playback if any
	if p.cmd != nil {
		p.stopLocked()
	}

	// Start mpv with IPC server and battery optimization flags
	p.cmd = exec.Command("mpv",
		"--no-video",                            // Audio only
		"--no-terminal",                         // Don't take over terminal
		"--input-ipc-server="+p.socketPath,      // IPC socket for control
		"--volume="+fmt.Sprintf("%d", p.volume), // Set volume
		"--cache=yes",                           // Enable caching
		"--cache-secs=5",                        // Small buffer for streaming
		"--demuxer-max-bytes=500K",              // Limit memory usage
		"--audio-buffer=0.2",                    // Small audio buffer
		url,
	)

	// Capture stdout/stderr for debugging
	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	p.stdout = stdout

	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	// Read stderr in background for error detection
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// Could log errors here if needed
			_ = scanner.Text()
		}
	}()

	p.currentURL = url
	p.state = StatePlaying

	return nil
}

// Pause pauses playback
func (p *MPVPlayer) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != StatePlaying {
		return fmt.Errorf("not playing")
	}

	if err := p.sendCommand(map[string]interface{}{
		"command": []interface{}{"set_property", "pause", true},
	}); err != nil {
		return err
	}

	p.state = StatePaused
	return nil
}

// Resume resumes playback
func (p *MPVPlayer) Resume() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != StatePaused {
		return fmt.Errorf("not paused")
	}

	if err := p.sendCommand(map[string]interface{}{
		"command": []interface{}{"set_property", "pause", false},
	}); err != nil {
		return err
	}

	p.state = StatePlaying
	return nil
}

// Stop stops playback
func (p *MPVPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.stopLocked()
}

// stopLocked stops playback (must be called with lock held)
func (p *MPVPlayer) stopLocked() error {
	if p.cmd == nil || p.state == StateStopped {
		return nil
	}

	// Send quit command
	_ = p.sendCommand(map[string]interface{}{
		"command": []interface{}{"quit"},
	})

	// Kill process if still running
	if p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}

	// Wait for process to exit
	_ = p.cmd.Wait()

	p.cmd = nil
	p.state = StateStopped
	p.currentURL = ""

	// Clean up socket
	_ = os.Remove(p.socketPath)

	return nil
}

// SetVolume sets the volume (0-100)
func (p *MPVPlayer) SetVolume(volume int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	p.volume = volume

	if p.state == StatePlaying || p.state == StatePaused {
		if err := p.sendCommand(map[string]interface{}{
			"command": []interface{}{"set_property", "volume", volume},
		}); err != nil {
			return err
		}
	}

	return nil
}

// GetVolume returns the current volume
func (p *MPVPlayer) GetVolume() (int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.volume, nil
}

// GetState returns the current player state
func (p *MPVPlayer) GetState() State {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.state
}

// IsPlaying returns true if currently playing
func (p *MPVPlayer) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.state == StatePlaying
}

// Close closes the player and cleans up resources
func (p *MPVPlayer) Close() error {
	return p.Stop()
}

// connectWithRetry connects to the MPV socket with retry logic
func (p *MPVPlayer) connectWithRetry() (net.Conn, error) {
	if p.socketPath == "" {
		return nil, fmt.Errorf("no socket path")
	}

	maxRetries := 10
	var conn net.Conn
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = net.Dial("unix", p.socketPath)
		if err == nil {
			return conn, nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to connect to mpv socket after retries: %w", err)
}

// sendCommand sends a JSON command to mpv via IPC socket
func (p *MPVPlayer) sendCommand(cmd map[string]interface{}) error {
	conn, err := p.connectWithRetry()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))

	// Encode and send command
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(cmd); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// getProperty queries an MPV property and returns the response
func (p *MPVPlayer) getProperty(property string) (interface{}, error) {
	conn, err := p.connectWithRetry()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Set read/write deadlines
	conn.SetDeadline(time.Now().Add(2 * time.Second))

	// Send command
	cmd := map[string]interface{}{
		"command":    []interface{}{"get_property", property},
		"request_id": 1,
	}
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	decoder := json.NewDecoder(conn)
	var response MPVResponse
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "success" && response.Error != "" {
		return nil, fmt.Errorf("MPV error: %s", response.Error)
	}

	return response.Data, nil
}

// GetMetadata retrieves current playback metadata
func (p *MPVPlayer) GetMetadata() (*Metadata, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.state == StateStopped {
		return nil, fmt.Errorf("not playing")
	}

	metadata := &Metadata{}

	// Get song title from icy-title only
	// media-title fallback removed because it returns ugly URLs/filenames
	if title, err := p.getProperty("metadata/icy-title"); err == nil {
		if titleStr, ok := title.(string); ok {
			metadata.Title = titleStr
		}
	}

	if genre, err := p.getProperty("metadata/icy-genre"); err == nil {
		if genreStr, ok := genre.(string); ok {
			metadata.Genre = genreStr
		}
	}

	if cacheTime, err := p.getProperty("demuxer-cache-time"); err == nil {
		if secs, ok := cacheTime.(float64); ok {
			metadata.BufferSecs = secs
		}
	}

	if bitrate, err := p.getProperty("audio-bitrate"); err == nil {
		if bps, ok := bitrate.(float64); ok {
			metadata.ActualKbps = bps / 1000
		}
	}

	return metadata, nil
}
