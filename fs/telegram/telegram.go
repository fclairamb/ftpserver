// Package telegram provides a telegram access layer
package telegram

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"

	"os"
	"sync/atomic"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/spf13/afero"

	"github.com/fclairamb/ftpserver/config/confpar"
	"gopkg.in/telebot.v3/middleware"
)

// ErrNotImplemented is returned when something is not implemented
var ErrNotImplemented = errors.New("not implemented")

// ErrNotFound is returned when something is not found
var ErrNotFound = errors.New("not found")

// ErrInvalidParameter is returned when a parameter is invalid
var ErrInvalidParameter = errors.New("invalid parameter")

// defaultMaxPartSize is the default maximum size of each telegram upload part (49 MB, telegram limit is 50 MB)
const defaultMaxPartSize = 49 * 1024 * 1024

const sendRetryAttempts = 3
const sendRetryDelay = 1200 * time.Millisecond

// Fs is a write-only afero.Fs implementation using telegram as backend
type Fs struct {
	// Bot is the telegram bot instance
	Bot *tele.Bot
	// ChatID is the telegram chat ID to send files to
	ChatID int64
	// Logger is the logger, obviously
	Logger *slog.Logger
	// MaxPartSize is the maximum size of each part when splitting large files (bytes)
	MaxPartSize int64
	// TempDir is the optional directory used to store multipart temporary files
	TempDir string

	// fakeFs is a lightweight fake filesystem intended for store temporary info about files
	// since some ftp clients expect to perform mkdir() + stat() on files and directories before upload
	fakeFs *fakeFilesystem
}

// File is the afero.File implementation
type File struct {
	// Path is the file path
	Path string
	// Content is the file content
	Content []byte
	// PartNumber is the uploaded part counter for multipart uploads
	PartNumber int
	// PartTempFiles stores temporary part file paths before final upload
	PartTempFiles []string
	// PartTempDir stores the temporary directory used for multipart spool
	PartTempDir string
	// TotalWritten stores total bytes written across the transfer
	TotalWritten int64
	// Fs is the parent Fs
	Fs *Fs
	// At is the current position in the file
	At int64
}

// imageExtensions is the list of supported image extensions
var imageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif"}

// videoExtensions is the list of supported video extensions
var videoExtensions = []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".mpeg", ".mpg", ".m4v", ".3gp", ".3g2"}

// textExtensions is the list of supported text extensions
var textExtensions = []string{".txt", ".md"}

// audioExtensions is the list of supported audio extensions
var audioExtensions = []string{".mp3", ".ogg", ".flac", ".wav", ".m4a", ".opus"}

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access, logger *slog.Logger) (afero.Fs, error) {

	token := access.Params["Token"]
	if token == "" {
		return nil, fmt.Errorf("parameter Token is empty")
	}

	chatID, err := strconv.ParseInt(access.Params["ChatID"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ChatID parameter: %v", err)
	}

	// Parse MaxPartSize param, default to defaultMaxPartSize if not set
	var maxPartSize int64 = defaultMaxPartSize
	partSizeStr := access.Params["MaxPartSize"]
	if partSizeStr != "" {
		parsed, parseErr := strconv.ParseInt(partSizeStr, 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid MaxPartSize parameter: %v", parseErr)
		}
		if parsed <= 0 {
			return nil, fmt.Errorf("invalid MaxPartSize parameter: must be > 0")
		}
		maxPartSize = parsed
	}

	tempDir := strings.TrimSpace(access.Params["TempDir"])

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		logger.Error("telegram bot initialization", "err", err)
		return nil, err
	}
	bot.Use(middleware.Logger())
	bot.Use(middleware.AutoRespond())

	bot.Handle("/start", startHandler)
	bot.Handle("/help", helpHandler)

	go func() {
		// Run bot in the background
		bot.Start()
	}()

	fs := &Fs{
		Bot:         bot,
		Logger:      logger,
		ChatID:      chatID,
		MaxPartSize: maxPartSize,
		TempDir:     tempDir,
		fakeFs:      newFakeFilesystem(),
	}

	return fs, nil
}

// Name of the file
func (f *File) Name() string { return f.Path }

// Close closes the file transfer and does the actual transfer to telegram.
// If the file is larger than maxPartSize, it will be split into multiple parts.
func (f *File) Close() error {
	if f.Fs == nil {
		return ErrNotFound
	}

	if entry := f.Fs.fakeFs.stat(f.Path); entry != nil && entry.IsDir() {
		// Some FTP clients open/close directory paths; never upload directories to Telegram.
		return nil
	}

	basePath := filepath.Base(f.Path)
	defer f.cleanupTempParts()

	if f.PartNumber == 0 && len(f.Content) == 0 {
		// Telegram rejects empty files; skip upload for empty payloads.
		if f.Fs.fakeFs.stat(f.Path) == nil {
			f.Fs.fakeFs.create(f.Path)
		}
		f.Fs.fakeFs.setSize(f.Path, 0)
		return nil
	}

	if f.PartNumber == 0 {
		// Single-part upload path
		err := f.sendContent(f.Content, basePath, basePath, false)
		if err != nil {
			return err
		}
	} else {
		// Multipart upload path: write final remainder, then upload all parts with partXofY naming.
		if len(f.Content) > 0 {
			if err := f.spillNextPartToDisk(f.Content); err != nil {
				return err
			}
		}

		totalParts := len(f.PartTempFiles)
		partFilenames := make([]string, totalParts)
		for idx, tempPath := range f.PartTempFiles {
			partNum := idx + 1
			partFilename := fmt.Sprintf("%s.part%dof%d", basePath, partNum, totalParts)
			partCaption := fmt.Sprintf("%s (part %d/%d)", basePath, partNum, totalParts)

			partData, readErr := os.ReadFile(tempPath)
			if readErr != nil {
				return readErr
			}

			if err := f.sendContent(partData, partFilename, partCaption, true); err != nil {
				return err
			}

			if err := os.Remove(tempPath); err != nil {
				f.Fs.Logger.Warn("telegram remove temp part failed", "err", err, "tempFilePath", tempPath)
			}

			partFilenames[idx] = partFilename
		}

		f.sendJoinInstructions(basePath, partFilenames)
	}

	if f.Fs.fakeFs.stat(f.Path) == nil {
		f.Fs.fakeFs.create(f.Path)
	}
	f.Fs.fakeFs.setSize(f.Path, f.TotalWritten)

	f.Content = []byte{}
	f.PartNumber = 0
	f.PartTempFiles = nil
	f.PartTempDir = ""
	f.TotalWritten = 0
	f.At = 0

	return nil
}

// sendContent sends a chunk of content to the telegram chat with the given caption
// filename is used as the file name for document uploads; caption is the display text.
func (f *File) sendContent(data []byte, filename string, caption string, forceDocument bool) error {
	chat := tele.Chat{ID: f.Fs.ChatID}

	err := f.sendWithRetry(func() error {
		if !forceDocument && isExtension(f.Path, imageExtensions) {
			photo := tele.Photo{File: tele.FromReader(bytes.NewReader(data)), Caption: caption}
			_, sendErr := f.Fs.Bot.Send(&chat, &photo)
			return sendErr
		}
		if !forceDocument && isExtension(f.Path, videoExtensions) {
			video := tele.Video{File: tele.FromReader(bytes.NewReader(data)), Caption: caption}
			_, sendErr := f.Fs.Bot.Send(&chat, &video)
			return sendErr
		}
		if !forceDocument && isExtension(f.Path, audioExtensions) {
			audio := tele.Audio{File: tele.FromReader(bytes.NewReader(data)), Caption: caption}
			_, sendErr := f.Fs.Bot.Send(&chat, &audio)
			return sendErr
		}
		if !forceDocument && isExtension(f.Path, textExtensions) && len(data) < 4096 {
			if isExtension(f.Path, []string{".md"}) {
				_, sendErr := f.Fs.Bot.Send(&chat, string(data), tele.ModeMarkdown)
				return sendErr
			}
			_, sendErr := f.Fs.Bot.Send(&chat, string(data))
			return sendErr
		}

		document := tele.Document{File: tele.FromReader(bytes.NewReader(data)), Caption: caption}
		document.FileName = filename
		_, sendErr := f.Fs.Bot.Send(&chat, &document)
		return sendErr
	})

	f.Fs.Logger.Info("telegram Bot.Send()", "path", f.Path, "caption", caption)

	if err != nil {
		f.Fs.Logger.Error("telegram Bot.Send()", "err", err)
		return err
	}

	return nil
}

func isTransientTelegramError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	transientHints := []string{
		"too many requests",
		"timeout",
		"deadline exceeded",
		"temporarily unavailable",
		"bad gateway",
		"internal server error",
		"gateway timeout",
		"connection reset",
		"eof",
		"i/o timeout",
	}

	for _, hint := range transientHints {
		if strings.Contains(msg, hint) {
			return true
		}
	}

	return false
}

func (f *File) sendWithRetry(send func() error) error {
	var err error

	for attempt := 1; attempt <= sendRetryAttempts; attempt++ {
		err = send()
		if err == nil {
			return nil
		}

		if !isTransientTelegramError(err) || attempt == sendRetryAttempts {
			return err
		}

		f.Fs.Logger.Warn("telegram send retry", "attempt", attempt, "maxAttempts", sendRetryAttempts, "err", err)
		time.Sleep(sendRetryDelay)
	}

	return err
}

func (f *File) spillNextPartToDisk(data []byte) error {
	if f.PartTempDir == "" {
		tempDir, err := os.MkdirTemp(f.Fs.TempDir, "ftpserver-telegram-*")
		if err != nil {
			f.Fs.Logger.Error("telegram spillNextPartToDisk MkdirTemp", "err", err, "tempDir", f.Fs.TempDir)
			return err
		}
		f.PartTempDir = tempDir
	}

	f.PartNumber++
	tempFilePath := filepath.Join(f.PartTempDir, fmt.Sprintf("part-%05d.bin", f.PartNumber))

	if err := os.WriteFile(tempFilePath, data, 0o600); err != nil {
		f.Fs.Logger.Error("telegram spillNextPartToDisk WriteFile", "err", err, "tempFilePath", tempFilePath, "bytes", len(data))
		return err
	}

	f.PartTempFiles = append(f.PartTempFiles, tempFilePath)

	return nil
}

func (f *File) cleanupTempParts() {
	if f.PartTempDir != "" {
		if err := os.RemoveAll(f.PartTempDir); err != nil {
			f.Fs.Logger.Error("telegram cleanupTempParts", "err", err)
		}
	}
}

// sendJoinInstructions sends a text message with instructions on how to join downloaded parts
func (f *File) sendJoinInstructions(originalName string, partFilenames []string) {
	if len(partFilenames) == 0 {
		return
	}

	chat := tele.Chat{ID: f.Fs.ChatID}
	totalParts := len(partFilenames)

	linuxCmd := fmt.Sprintf("for i in $(seq 1 %d); do cat '%s.part${i}of%d' >> '%s'; done", totalParts, originalName, totalParts, originalName)
	pwshCmd := fmt.Sprintf("1..%d | %% { Get-Content -AsByteStream '%s.part${_}of%d' } | Set-Content -AsByteStream '%s'", totalParts, originalName, totalParts, originalName)

	msg := fmt.Sprintf(
		"📦 File: %s\n📊 Total parts: %d\n\n"+
			"To join parts after downloading:\n\n"+
			"Linux/Mac:\n"+
			"```\n%s\n```\n\n"+
			"Windows (PowerShell):\n"+
			"```\n%s\n```",
		originalName, totalParts, linuxCmd, pwshCmd,
	)

	if len(msg) > 3800 {
		msg = fmt.Sprintf(
			"📦 File: %s\n📊 Total parts: %d\n\n"+
				"Message shortened to avoid Telegram length limits.\n"+
				"Use this pattern to reassemble:\n\n"+
				"Linux/Mac:\n"+
				"```\ncat %s.part*of%d > %s\n```\n\n"+
				"Windows (PowerShell):\n"+
				"```\n1..%d | %% { Get-Content -AsByteStream '%s.part${_}of%d' } | Set-Content -AsByteStream '%s'\n```",
			originalName, totalParts,
			originalName, totalParts, originalName,
			totalParts, originalName, totalParts, originalName,
		)
	}

	if _, err := f.Fs.Bot.Send(&chat, msg, tele.ModeMarkdown); err != nil {
		f.Fs.Logger.Error("telegram sendJoinInstructions", "err", err)
	}
}

// Read stores the received file content into the local buffer
func (f *File) Read(b []byte) (int, error) {
	n := 0

	if len(b) > 0 && int(f.At) == len(f.Content) {
		return 0, io.EOF
	}

	if len(f.Content)-int(f.At) >= len(b) {
		n = len(b)
	} else {
		n = len(f.Content) - int(f.At)
	}

	copy(b, f.Content[f.At:f.At+int64(n)])
	atomic.AddInt64(&f.At, int64(n))

	return n, nil
}

// ReadAt is not implemented
func (f *File) ReadAt(_ []byte, _ int64) (int, error) {
	return 0, ErrNotImplemented
}

// Truncate is not implemented
func (f *File) Truncate(_ int64) error {
	return nil
}

// Readdir is not implemented
func (f *File) Readdir(_ int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

// Readdirnames is not implemented
func (f *File) Readdirnames(_ int) ([]string, error) {
	return []string{}, nil
}

// Seek is not implemented
func (f *File) Seek(_ int64, _ int) (int64, error) {
	return 0, nil
}

// Stat for the file relies on the fake filesystem
func (f *File) Stat() (os.FileInfo, error) {
	fileInfo := f.Fs.fakeFs.stat(f.Path)

	if fileInfo == nil {
		return nil, &os.PathError{Op: "stat", Path: f.Path, Err: nil}
	}
	return fileInfo, nil
}

// Sync is not implemented
func (f *File) Sync() error {
	return nil
}

// WriteString is not implemented
func (f *File) WriteString(s string) (int, error) {
	return 0, ErrNotImplemented
}

// WriteAt is not implemented
func (f *File) WriteAt(b []byte, off int64) (int, error) {
	return 0, ErrNotImplemented
}

func (f *File) Write(b []byte) (int, error) {
	f.Content = append(f.Content, b...)
	f.TotalWritten += int64(len(b))

	partSize := int(f.Fs.MaxPartSize)
	if partSize <= 0 {
		f.Fs.Logger.Error("telegram Write invalid MaxPartSize", "maxPartSize", f.Fs.MaxPartSize, "path", f.Path)
		return 0, ErrInvalidParameter
	}

	// Keep max one part in memory and stream full parts immediately.
	for len(f.Content) > partSize {
		part := make([]byte, partSize)
		copy(part, f.Content[:partSize])

		if err := f.spillNextPartToDisk(part); err != nil {
			f.Fs.Logger.Error("telegram Write spillNextPartToDisk failed", "err", err, "path", f.Path, "partSize", partSize)
			return 0, err
		}

		f.Content = f.Content[partSize:]
	}

	return len(b), nil
}

// Name of the filesystem
func (m *Fs) Name() string {
	return "telegram"
}

// Chtimes is not implemented
func (m *Fs) Chtimes(name string, atime, mtime time.Time) error {
	return nil
}

// Chmod is not implemented
func (m *Fs) Chmod(name string, mode os.FileMode) error {
	return nil
}

// Rename is not implemented
func (m *Fs) Rename(name string, newname string) error {
	return nil
}

// Chown is not implemented
func (m *Fs) Chown(string, int, int) error {
	return nil
}

// RemoveAll is not implemented
func (m *Fs) RemoveAll(name string) error {
	return nil
}

// Remove is not implemented
func (m *Fs) Remove(name string) error {
	return nil
}

// Mkdir
func (m *Fs) Mkdir(name string, mode os.FileMode) error {
	m.fakeFs.mkdir(name, mode)
	return nil
}

// MkdirAll creates full path of directories
// like mkdir -p
func (m *Fs) MkdirAll(name string, mode os.FileMode) error {
	path := strings.Split(name, "/")
	for i := 0; i < len(path); i++ {
		dir := strings.Join(path[:i+1], "/")
		m.fakeFs.mkdir(dir, mode)
	}
	return nil
}

// Open opens a file buffer
func (m *Fs) Open(name string) (afero.File, error) {
	return &File{Path: name, Fs: m}, nil
}

// Create creates a file buffer
func (m *Fs) Create(name string) (afero.File, error) {
	m.fakeFs.create(name)
	return &File{Path: name, Fs: m}, nil
}

// OpenFile opens a file buffer
func (m *Fs) OpenFile(name string, flag int, mode os.FileMode) (afero.File, error) {
	return &File{Path: name, Fs: m}, nil
}

// Stat() fake implementation
func (m *Fs) Stat(name string) (os.FileInfo, error) {
	fileInfo := m.fakeFs.stat(name)

	if fileInfo == nil {
		return nil, &os.PathError{Op: "stat", Path: name, Err: nil}
	}
	return fileInfo, nil
}

// LstatIfPossible is not implemented
func (m *Fs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	return nil, false, &os.PathError{Op: "lstat", Path: name, Err: nil}
}

func isExtension(filename string, extensions []string) bool {
	extension := strings.ToLower(filepath.Ext(filename))
	for _, ext := range extensions {
		if extension == ext {
			return true
		}
	}
	return false
}

const readMeURL = "https://github.com/slayer/ftpserver"

// /start command handler
func startHandler(c tele.Context) error {
	err := helpHandler(c)
	if err != nil {
		return err
	}
	var chatID int64
	chat := c.Chat()
	if chat != nil {
		chatID = chat.ID
	}

	err = c.Send(fmt.Sprintf("Current `ChatID` is `%d`", chatID), tele.ModeMarkdown)
	return err
}

func helpHandler(c tele.Context) error {
	firstName := "<unknown>"
	if c.Sender() != nil {
		firstName = c.Sender().FirstName
	}
	message := fmt.Sprintf("Hello %s!, you can read more about me at %s", firstName, readMeURL)
	return c.Send(message)
}
