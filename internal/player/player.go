package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	utils "github.com/sina96/ytunes/internal/utils"
)

type Player interface {
	Play(url string) (Metadata, error)
	Pause() error
	Resume() error
	Stop() error
	Wait() error
	IsPlaying() bool
	IsPaused() bool
	Position() (float64, error)
}

type Metadata struct {
	Title           string
	Duration        string
	DurationSeconds int
}

type mpvPlayer struct {
	cmd       *exec.Cmd
	paused    bool
	ipcConn   net.Conn
	ipcReader *bufio.Reader
	ipcPath   string
	nextReqID int
	mutex     sync.Mutex
}

type ipcReply struct {
	Error     string          `json:"error"`
	Data      json.RawMessage `json:"data"`
	RequestID int             `json:"request_id"`
}

func (p *mpvPlayer) sendIPC(command []any) (json.RawMessage, error) {
	if p.ipcConn == nil {
		return nil, fmt.Errorf("mpv ipc: not connected")
	}

	p.nextReqID++
	id := p.nextReqID

	message, err := json.Marshal(map[string]any{
		"command":    command,
		"request_id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal IPC command: %w", err)
	}

	if _, err := p.ipcConn.Write(append(message, '\n')); err != nil {
		return nil, fmt.Errorf("failed to send IPC command: %w", err)
	}

	line, err := p.ipcReader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	var reply ipcReply
	if err := json.Unmarshal([]byte(line), &reply); err != nil {
		return nil, err
	}
	if reply.Error != "success" {
		return nil, fmt.Errorf("mpv ipc error: %s", reply.Error)
	}
	return reply.Data, nil

}

func newSocketPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("ytunes-socket-%d-%d", os.Getpid(), time.Now().UnixNano()))
}

func buildPlayCommand(url string, ipcPath string) (*exec.Cmd, string, error) {

	trimmedUrl := strings.TrimSpace(url)
	if !utils.IsYouTubeURL(trimmedUrl) {
		return nil, "", fmt.Errorf("not a YouTube URL")
	}

	mpvPath, err := exec.LookPath("mpv")
	if err != nil {
		return nil, "", fmt.Errorf("mpv not found. Please install it and try again.")
	}

	cmd := exec.Command(mpvPath, "--no-video", "--ytdl-format=bestaudio", "--input-ipc-server="+ipcPath, trimmedUrl)
	return cmd, trimmedUrl, nil
}

func connectIPC(path string) (net.Conn, error) {
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.Dial("unix", path)
		if err == nil {
			return conn, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("failed to connect to IPC socket: %w", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func fetchMetadata(url string) (Metadata, error) {
	ytbdlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		return Metadata{}, fmt.Errorf("yt-dlp not found. Please install it and try again.")
	}
	cmd := exec.Command(ytbdlpPath, "--no-playlist", "--print", "title,duration", url)
	stdout, err := cmd.Output()
	if err != nil {
		return Metadata{}, err
	}
	lines := strings.Split(string(stdout), "\n")
	if len(lines) < 2 {
		return Metadata{}, fmt.Errorf("unexpected output from yt-dlp")
	}
	title := lines[0]
	duration := lines[1]
	seconds, err := strconv.Atoi(strings.TrimSpace(duration))
	if err != nil {
		return Metadata{}, err
	}
	return Metadata{Title: title, Duration: duration, DurationSeconds: seconds}, nil
}

func (p *mpvPlayer) Play(url string) (Metadata, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.paused = false

	ipcPath := newSocketPath()
	playCmd, url, err := buildPlayCommand(url, ipcPath)
	if err != nil {
		return Metadata{}, err
	}

	meta, err := fetchMetadata(url)
	if err != nil {
		return Metadata{}, err
	}

	p.cmd = playCmd
	p.ipcPath = ipcPath

	err = p.cmd.Start()
	if err != nil {
		return Metadata{}, err
	}

	conn, err := connectIPC(ipcPath)
	if err != nil {
		return Metadata{}, err
	}
	p.ipcConn = conn
	p.ipcReader = bufio.NewReader(conn)

	return meta, nil
}

func (p *mpvPlayer) Pause() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.cmd == nil {
		return fmt.Errorf("no command to pause")
	}
	if p.paused {
		return nil
	}
	if _, err := p.sendIPC([]any{"set_property", "pause", true}); err != nil {
		return err
	}
	p.paused = true
	return nil
}

func (p *mpvPlayer) Resume() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.cmd == nil {
		return fmt.Errorf("no command to resume")
	}
	if !p.paused {
		return nil
	}
	if _, err := p.sendIPC([]any{"set_property", "pause", false}); err != nil {
		return err
	}
	p.paused = false
	return nil
}

func (p *mpvPlayer) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.cmd == nil {
		return fmt.Errorf("no command to stop")
	}
	if p.paused {
		if p.cmd.Process != nil {
			err := p.cmd.Process.Signal(syscall.SIGCONT)
			if err != nil {
				return err
			}
		}
	}
	p.paused = false
	if p.cmd.Process != nil {
		err := p.cmd.Process.Signal(syscall.SIGINT)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *mpvPlayer) Wait() error {
	p.mutex.Lock()
	cmd := p.cmd
	p.mutex.Unlock()

	if cmd == nil {
		return fmt.Errorf("no command to wait")
	}
	err := cmd.Wait()
	p.mutex.Lock()
	p.cmd = nil
	p.paused = false
	if p.ipcConn != nil {
		p.ipcConn.Close()
		p.ipcConn = nil
		p.ipcReader = nil
	}
	if p.ipcPath != "" {
		os.Remove(p.ipcPath)
		p.ipcPath = ""
	}
	p.mutex.Unlock()
	return err
}

func (p *mpvPlayer) Position() (float64, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.cmd == nil {
		return 0, fmt.Errorf("no command to query")
	}

	data, err := p.sendIPC([]any{"get_property", "time-pos"})
	if err != nil {
		return 0, err
	}

	var seconds float64
	if err := json.Unmarshal(data, &seconds); err != nil {
		return 0, err
	}
	return seconds, nil
}

func (p *mpvPlayer) IsPlaying() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.cmd != nil && !p.paused
}

func (p *mpvPlayer) IsPaused() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.cmd != nil && p.paused
}

// New returns a new Player instance using the mpvPlayer implementation.
func New() Player {
	return &mpvPlayer{}
}
