package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/model"
)

const contentTypeSniffBytes = 64 * 1024

var ErrFileRequired = errors.New("file is required")
var ErrEmptyFile = errors.New("empty file")
var ErrUnsupportedFileType = errors.New("unsupported file type")
var ErrFileTooLarge = errors.New("file too large")

type UploadRequest struct {
	FileHeader *multipart.FileHeader
	BaseURL    string
	SceneCode  string
}

type UploadService struct {
	cfg config.UploadConfig
}

type uploadCategory struct {
	subdir     string
	publicPath string
	maxSize    int64
}

func NewUploadService(cfg config.UploadConfig) *UploadService {
	return &UploadService{cfg: cfg}
}

// =========================
// 图片上传：仍然保持“扩展名 + 内容类型”双重校验
// =========================

func (s *UploadService) UploadImage(ctx context.Context, req *UploadRequest) (*model.UploadResponse, error) {
	return s.uploadImage(ctx, req)
}

func (s *UploadService) uploadImage(_ context.Context, req *UploadRequest) (*model.UploadResponse, error) {
	sceneCode, err := uploadSceneCode(req)
	if err != nil {
		return nil, err
	}
	category := uploadCategory{
		subdir:     "images",
		publicPath: "/static/" + sceneCode + "/images/",
		maxSize:    s.cfg.ImageMaxSize,
	}

	if req == nil || req.FileHeader == nil {
		return nil, ErrFileRequired
	}
	if req.FileHeader.Size <= 0 {
		return nil, ErrEmptyFile
	}
	if req.FileHeader.Size > category.maxSize {
		return nil, ErrFileTooLarge
	}

	filename := sanitizeFilename(req.FileHeader.Filename)
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil, unsupportedFileTypeError(ext, req.FileHeader.Header.Get("Content-Type"), "")
	}

	allowedImageTypes := allowedImageContentTypesByExt()
	allowedContentTypes, ok := allowedImageTypes[ext]
	if !ok {
		return nil, unsupportedFileTypeError(ext, req.FileHeader.Header.Get("Content-Type"), "")
	}

	contentType, err := sniffContentType(req.FileHeader, ext, allowedContentTypes)
	if err != nil {
		return nil, err
	}

	targetDir := filepath.Join(s.cfg.RootDir, sceneCode, category.subdir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}

	storedName, targetPath, err := createTargetFile(targetDir, ext)
	if err != nil {
		return nil, fmt.Errorf("create target file: %w", err)
	}

	if err := copyUploadedFile(req.FileHeader, targetPath); err != nil {
		return nil, fmt.Errorf("save uploaded file: %w", err)
	}

	return &model.UploadResponse{
		URL:         joinPublicURL(req.BaseURL, category.publicPath, storedName),
		Filename:    storedName,
		Size:        req.FileHeader.Size,
		ContentType: contentType,
	}, nil
}

// =========================
// 音频上传：改为“内容优先”
// 1. 先识别真实内容类型
// 2. 再映射标准扩展名
// 3. 保存时使用标准扩展名
// =========================

func (s *UploadService) UploadAudio(ctx context.Context, req *UploadRequest) (*model.UploadResponse, error) {
	return s.uploadAudio(ctx, req)
}

func (s *UploadService) uploadAudio(_ context.Context, req *UploadRequest) (*model.UploadResponse, error) {
	sceneCode, err := uploadSceneCode(req)
	if err != nil {
		return nil, err
	}
	category := uploadCategory{
		subdir:     "audio",
		publicPath: "/static/" + sceneCode + "/audio/",
		maxSize:    s.cfg.AudioMaxSize,
	}

	if req == nil || req.FileHeader == nil {
		return nil, ErrFileRequired
	}
	if req.FileHeader.Size <= 0 {
		return nil, ErrEmptyFile
	}
	if req.FileHeader.Size > category.maxSize {
		return nil, ErrFileTooLarge
	}

	// 只把原始扩展名作为日志/辅助信息，不再作为最终准入依据
	originalName := sanitizeFilename(req.FileHeader.Filename)
	originalExt := strings.ToLower(filepath.Ext(originalName))

	contentType, err := sniffActualAudioContentType(req.FileHeader)
	if err != nil {
		return nil, err
	}

	standardExt, ok := audioExtByContentType(contentType)
	if !ok {
		return nil, unsupportedFileTypeError(
			originalExt,
			req.FileHeader.Header.Get("Content-Type"),
			contentType,
		)
	}

	targetDir := filepath.Join(s.cfg.RootDir, sceneCode, category.subdir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}

	// 注意：这里不再使用 originalExt，而是使用识别后的标准扩展名
	storedName, targetPath, err := createTargetFile(targetDir, standardExt)
	if err != nil {
		return nil, fmt.Errorf("create target file: %w", err)
	}

	if err := copyUploadedFile(req.FileHeader, targetPath); err != nil {
		return nil, fmt.Errorf("save uploaded file: %w", err)
	}

	return &model.UploadResponse{
		URL:         joinPublicURL(req.BaseURL, category.publicPath, storedName),
		Filename:    storedName,
		Size:        req.FileHeader.Size,
		ContentType: contentType,
	}, nil
}

// =========================
// 通用 sniff 逻辑
// =========================

func sniffContentType(fileHeader *multipart.FileHeader, ext string, allowed map[string]struct{}) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer file.Close()

	buffer, err := io.ReadAll(io.LimitReader(file, contentTypeSniffBytes))
	if err != nil {
		return "", fmt.Errorf("read uploaded file header: %w", err)
	}
	if len(buffer) == 0 {
		return "", ErrEmptyFile
	}

	detected := normalizeContentType(http.DetectContentType(buffer[:min(len(buffer), 512)]))
	declared := normalizeContentType(fileHeader.Header.Get("Content-Type"))

	if _, ok := allowed[detected]; ok {
		return detected, nil
	}

	if fallback := detectAudioContentType(buffer, allowed); fallback != "" {
		return fallback, nil
	}

	if detected == "application/octet-stream" {
		if _, ok := allowed[declared]; ok {
			return declared, nil
		}
	}

	return "", unsupportedFileTypeError(ext, declared, detected)
}

// 专门给音频上传使用：
// 不依赖原始扩展名，直接从真实内容判断出音频类型。
func sniffActualAudioContentType(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer file.Close()

	buffer, err := io.ReadAll(io.LimitReader(file, contentTypeSniffBytes))
	if err != nil {
		return "", fmt.Errorf("read uploaded file header: %w", err)
	}
	if len(buffer) == 0 {
		return "", ErrEmptyFile
	}

	declared := normalizeContentType(fileHeader.Header.Get("Content-Type"))
	detected := normalizeContentType(http.DetectContentType(buffer[:min(len(buffer), 512)]))

	// 1) 先信标准库可直接识别出的结果
	if isSupportedAudioContentType(detected) {
		return normalizeAudioContentType(detected), nil
	}

	// 2) 对 MP3 做增强检测：很多 MP3 前面有较长 ID3，512 字节检测不到
	if looksLikeMP3(buffer) {
		return "audio/mpeg", nil
	}

	// 3) 对 WAV 做头部特征识别
	if looksLikeWAV(buffer) {
		return "audio/wav", nil
	}

	// 4) 对 OGG 做头部特征识别
	if looksLikeOGG(buffer) {
		return "audio/ogg", nil
	}

	// 5) 对 MP4/M4A 容器做头部特征识别
	if looksLikeMP4Audio(buffer) {
		// 保持和你原先兼容：允许 video/mp4 这一类识别结果
		return "audio/mp4", nil
	}

	// 6) 如果标准库识别成 octet-stream，再尝试使用前端声明类型兜底
	if detected == "application/octet-stream" && isSupportedAudioContentType(declared) {
		return normalizeAudioContentType(declared), nil
	}

	return "", unsupportedFileTypeError("", declared, detected)
}

// =========================
// 音频类型映射
// =========================

func audioExtByContentType(contentType string) (string, bool) {
	switch normalizeAudioContentType(contentType) {
	case "audio/mpeg":
		return ".mp3", true
	case "audio/wav":
		return ".wav", true
	case "audio/ogg":
		return ".ogg", true
	case "audio/mp4":
		return ".m4a", true
	default:
		return "", false
	}
}

func isSupportedAudioContentType(contentType string) bool {
	switch normalizeAudioContentType(contentType) {
	case "audio/mpeg", "audio/wav", "audio/ogg", "audio/mp4":
		return true
	default:
		return false
	}
}

func normalizeAudioContentType(contentType string) string {
	switch normalizeContentType(contentType) {
	case "audio/mp3":
		return "audio/mpeg"
	case "audio/wave", "audio/x-wav":
		return "audio/wav"
	case "application/ogg":
		return "audio/ogg"
	case "video/mp4":
		return "audio/mp4"
	default:
		return normalizeContentType(contentType)
	}
}

func allowedImageContentTypesByExt() map[string]map[string]struct{} {
	return map[string]map[string]struct{}{
		".jpg":  contentTypeSet("image/jpeg"),
		".jpeg": contentTypeSet("image/jpeg"),
		".png":  contentTypeSet("image/png"),
		".webp": contentTypeSet("image/webp"),
	}
}

// =========================
// 文件写入
// =========================

func createTargetFile(targetDir, ext string) (string, string, error) {
	for range 5 {
		filename, err := generateStoredFilename(ext)
		if err != nil {
			return "", "", err
		}

		targetPath := filepath.Join(targetDir, filename)
		file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			if closeErr := file.Close(); closeErr != nil {
				return "", "", closeErr
			}
			return filename, targetPath, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", "", err
		}
	}

	return "", "", errors.New("could not allocate unique filename")
}

func copyUploadedFile(fileHeader *multipart.FileHeader, targetPath string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(targetPath, os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		_ = os.Remove(targetPath)
		return fmt.Errorf("copy uploaded file: %w", err)
	}

	return nil
}

func generateStoredFilename(ext string) (string, error) {
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random suffix: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102150405")
	suffix := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s-%s%s", timestamp, suffix, ext), nil
}

// =========================
// 文件名 / URL / Content-Type 工具
// =========================

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\x00", "")
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == string(filepath.Separator) {
		return ""
	}
	return name
}

func joinPublicURL(baseURL, publicPath, filename string) string {
	return strings.TrimRight(baseURL, "/") + publicPath + filename
}

func uploadSceneCode(req *UploadRequest) (string, error) {
	if req == nil {
		return "default", nil
	}

	sceneCode := req.SceneCode
	if strings.TrimSpace(sceneCode) == "" {
		sceneCode = "default"
	}

	normalized, err := model.ValidateSceneCode(sceneCode)
	if err != nil {
		return "", ErrInvalidContentParams
	}

	return normalized, nil
}

func normalizeContentType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	if idx := strings.Index(value, ";"); idx >= 0 {
		value = value[:idx]
	}

	return strings.TrimSpace(value)
}

func contentTypeSet(values ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[normalizeContentType(value)] = struct{}{}
	}
	return set
}

// =========================
// 音频内容识别增强
// =========================

func detectAudioContentType(buffer []byte, allowed map[string]struct{}) string {
	if len(buffer) < 4 {
		return ""
	}

	if _, ok := allowed["audio/mpeg"]; ok && looksLikeMP3(buffer) {
		return "audio/mpeg"
	}
	if _, ok := allowed["audio/wav"]; ok && looksLikeWAV(buffer) {
		return "audio/wav"
	}
	if _, ok := allowed["audio/ogg"]; ok && looksLikeOGG(buffer) {
		return "audio/ogg"
	}
	if _, ok := allowed["audio/mp4"]; ok {
		if looksLikeMP4Audio(buffer) {
			return "audio/mp4"
		}
	}

	if _, ok := allowed["video/mp4"]; ok {
		if looksLikeMP4Audio(buffer) {
			return "audio/mp4"
		}
	}

	return ""
}

func looksLikeMP3(buffer []byte) bool {
	if tagSize, ok := parseID3v2Tag(buffer); ok {
		if tagSize > len(buffer) {
			return true
		}
		if tagSize == len(buffer) {
			return false
		}
		return findMP3FrameHeader(buffer[tagSize:]) >= 0
	}

	return findMP3FrameHeader(buffer) >= 0
}

func parseID3v2Tag(buffer []byte) (int, bool) {
	if len(buffer) < 10 || !bytes.HasPrefix(buffer, []byte("ID3")) {
		return 0, false
	}

	version := buffer[3]
	if version < 2 || version > 4 {
		return 0, false
	}

	size := 0
	for _, value := range buffer[6:10] {
		if value&0x80 != 0 {
			return 0, false
		}
		size = (size << 7) | int(value)
	}

	total := 10 + size
	if buffer[5]&0x10 != 0 {
		total += 10
	}

	return total, true
}

func findMP3FrameHeader(buffer []byte) int {
	for i := 0; i+4 <= len(buffer); i++ {
		if isMP3FrameHeader(buffer[i:]) {
			return i
		}
	}

	return -1
}

func isMP3FrameHeader(buffer []byte) bool {
	if len(buffer) < 4 {
		return false
	}
	if buffer[0] != 0xff || buffer[1]&0xe0 != 0xe0 {
		return false
	}

	versionID := (buffer[1] >> 3) & 0x03
	layer := (buffer[1] >> 1) & 0x03
	bitrateIndex := (buffer[2] >> 4) & 0x0f
	sampleRateIndex := (buffer[2] >> 2) & 0x03
	paddingBit := (buffer[2] >> 1) & 0x01

	if versionID == 0x01 || layer != 0x01 {
		return false
	}
	if bitrateIndex == 0x00 || bitrateIndex == 0x0f {
		return false
	}
	if sampleRateIndex == 0x03 {
		return false
	}

	bitrate := mp3Bitrate(versionID, bitrateIndex)
	sampleRate := mp3SampleRate(versionID, sampleRateIndex)
	if bitrate == 0 || sampleRate == 0 {
		return false
	}

	frameLength := mp3FrameLength(versionID, bitrate, sampleRate, int(paddingBit))
	if frameLength <= 4 {
		return false
	}
	if frameLength+4 > len(buffer) {
		return true
	}

	return buffer[frameLength] == 0xff && buffer[frameLength+1]&0xe0 == 0xe0
}

func mp3Bitrate(versionID, bitrateIndex byte) int {
	v1Layer3 := [...]int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0}
	v2Layer3 := [...]int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}

	if versionID == 0x03 {
		return v1Layer3[bitrateIndex]
	}

	return v2Layer3[bitrateIndex]
}

func mp3SampleRate(versionID, sampleRateIndex byte) int {
	switch versionID {
	case 0x03:
		return [...]int{44100, 48000, 32000, 0}[sampleRateIndex]
	case 0x02:
		return [...]int{22050, 24000, 16000, 0}[sampleRateIndex]
	case 0x00:
		return [...]int{11025, 12000, 8000, 0}[sampleRateIndex]
	default:
		return 0
	}
}

func mp3FrameLength(versionID byte, bitrate, sampleRate, padding int) int {
	if versionID == 0x03 {
		return (144*bitrate*1000)/sampleRate + padding
	}
	return (72*bitrate*1000)/sampleRate + padding
}

func looksLikeWAV(buffer []byte) bool {
	// RIFF xxxx WAVE
	return len(buffer) >= 12 &&
		bytes.Equal(buffer[0:4], []byte("RIFF")) &&
		bytes.Equal(buffer[8:12], []byte("WAVE"))
}

func looksLikeOGG(buffer []byte) bool {
	return len(buffer) >= 4 && bytes.Equal(buffer[0:4], []byte("OggS"))
}

func looksLikeMP4Audio(buffer []byte) bool {
	// ISO BMFF 常见特征：前若干字节中有 ftyp
	// 你的那个误后缀 mp3 样本就是典型这种头
	if len(buffer) < 12 {
		return false
	}

	searchLimit := min(len(buffer), 64)
	for i := 4; i+4 <= searchLimit; i++ {
		if bytes.Equal(buffer[i:i+4], []byte("ftyp")) {
			return true
		}
	}

	return false
}

// =========================
// 错误 / 辅助
// =========================

func unsupportedFileTypeError(ext, declared, detected string) error {
	return fmt.Errorf(
		"%w: ext=%q declared_content_type=%q detected_content_type=%q",
		ErrUnsupportedFileType,
		ext,
		normalizeContentType(declared),
		normalizeContentType(detected),
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
