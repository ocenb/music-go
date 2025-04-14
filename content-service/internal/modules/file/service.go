package file

import (
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"github.com/ocenb/music-go/content-service/internal/clients/cloudinaryclient"
	"github.com/ocenb/music-go/content-service/internal/config"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type FileCategory string

const (
	AudioCategory  FileCategory = "audio"
	ImagesCategory FileCategory = "images"
	tempDir                     = "/app/temp"
)

type AudioResult struct {
	FileName string
	Duration int
}

type FileServiceInterface interface {
	SaveAudio(ctx context.Context, file *multipart.FileHeader) (*AudioResult, error)
	SaveImage(ctx context.Context, file *multipart.FileHeader) (string, error)
	DeleteFile(ctx context.Context, fileName string, category FileCategory) error
}

type FileService struct {
	cloudinary cloudinaryclient.CloudinaryClientInterface
	log        *slog.Logger
	cfg        *config.Config
}

func NewFileService(
	cloudinary cloudinaryclient.CloudinaryClientInterface,
	log *slog.Logger,
	cfg *config.Config,
) *FileService {
	return &FileService{
		cloudinary: cloudinary,
		log:        log,
		cfg:        cfg,
	}
}

func (s *FileService) SaveAudio(ctx context.Context, file *multipart.FileHeader) (*AudioResult, error) {
	if file.Size > s.cfg.AudioFileLimit {
		return nil, ErrAudioFileTooLarge
	}

	fileName := uuid.New().String()
	fileExt := strings.ToLower(filepath.Ext(file.Filename))

	tempFileName := fmt.Sprintf("%s_temp%s", fileName, fileExt)
	tempFilePath := filepath.Join(tempDir, tempFileName)

	outputFileName := fmt.Sprintf("%s.webm", fileName)
	outputFilePath := filepath.Join(tempDir, outputFileName)

	if err := s.saveMultipartFile(file, tempFilePath); err != nil {
		return nil, fmt.Errorf("failed to save temporary file: %w", err)
	}
	defer func() {
		err := os.Remove(tempFilePath)
		if err != nil {
			s.log.Error("Failed to remove file", "error", err)
		}
	}()

	if fileExt != ".webm" {
		if err := s.convertAudioToWebm(tempFilePath, outputFilePath); err != nil {
			return nil, fmt.Errorf("failed to convert to webm: %w", err)
		}
	} else {
		if err := s.normalizeAudio(tempFilePath, outputFilePath); err != nil {
			return nil, fmt.Errorf("failed to normalize audio: %w", err)
		}
	}
	defer func() {
		err := os.Remove(outputFilePath)
		if err != nil {
			s.log.Error("Failed to remove file", "error", err)
		}
	}()

	duration, err := s.getAudioDuration(outputFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %w", err)
	}

	if err := s.cloudinary.Upload(ctx, outputFilePath, fileName, "video", "audio"); err != nil {
		return nil, fmt.Errorf("failed to upload to cloudinary: %w", err)
	}

	return &AudioResult{
		FileName: fileName,
		Duration: duration,
	}, nil
}

func (s *FileService) SaveImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	if file.Size > s.cfg.ImageFileLimit {
		return "", ErrImageFileTooLarge
	}

	fileName := uuid.New().String()
	fileName250 := fmt.Sprintf("%s_250x250", fileName)
	fileName50 := fmt.Sprintf("%s_50x50", fileName)

	filePath250 := filepath.Join(tempDir, fmt.Sprintf("%s.jpg", fileName250))
	filePath50 := filepath.Join(tempDir, fmt.Sprintf("%s.jpg", fileName50))

	src250, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		err := src250.Close()
		if err != nil {
			s.log.Error("Failed to close file", "error", err)
		}
	}()

	out250, err := os.Create(filePath250)
	if err != nil {
		return "", fmt.Errorf("failed to create 250x250 file: %w", err)
	}
	defer func() {
		err := out250.Close()
		if err != nil {
			s.log.Error("Failed to close file", "error", err)
		}
	}()

	if err := s.resize(src250, out250, 250, 250); err != nil {
		return "", fmt.Errorf("failed to process 250x250 image: %w", err)
	}
	defer func() {
		err := os.Remove(filePath250)
		if err != nil {
			s.log.Error("Failed to remove file", "error", err)
		}
	}()

	src50, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		err := src50.Close()
		if err != nil {
			s.log.Error("Failed to close file", "error", err)
		}
	}()

	out50, err := os.Create(filePath50)
	if err != nil {
		return "", fmt.Errorf("failed to create 50x50 file: %w", err)
	}
	defer func() {
		err := out50.Close()
		if err != nil {
			s.log.Error("Failed to close file", "error", err)
		}
	}()

	if err := s.resize(src50, out50, 50, 50); err != nil {
		return "", fmt.Errorf("failed to process 50x50 image: %w", err)
	}
	defer func() {
		err := os.Remove(filePath50)
		if err != nil {
			s.log.Error("Failed to remove file", "error", err)
		}
	}()

	if err := s.cloudinary.Upload(ctx, filePath250, fileName250, "image", "images"); err != nil {
		return "", fmt.Errorf("failed to upload 250x250 image: %w", err)
	}

	if err := s.cloudinary.Upload(ctx, filePath50, fileName50, "image", "images"); err != nil {
		return "", fmt.Errorf("failed to upload 50x50 image: %w", err)
	}

	return fileName, nil
}

func (s *FileService) DeleteFile(ctx context.Context, fileName string, category FileCategory) error {
	if category == ImagesCategory {
		s.log.Info("Deleting 250x250 image", "fileName", fileName)
		if err := s.cloudinary.Delete(ctx, fmt.Sprintf("images/%s_250x250", fileName), "image"); err != nil {
			return fmt.Errorf("failed to delete 250x250 image: %w", err)
		}
		s.log.Info("Deleting 50x50 image", "fileName", fileName)
		if err := s.cloudinary.Delete(ctx, fmt.Sprintf("images/%s_50x50", fileName), "image"); err != nil {
			return fmt.Errorf("failed to delete 50x50 image: %w", err)
		}
	} else {
		s.log.Info("Deleting audio", "fileName", fileName)
		if err := s.cloudinary.Delete(ctx, fmt.Sprintf("audio/%s", fileName), "video"); err != nil {
			return fmt.Errorf("failed to delete audio file: %w", err)
		}
	}
	return nil
}

func (s *FileService) saveMultipartFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		err := src.Close()
		if err != nil {
			s.log.Error("Failed to close file", "error", err)
		}
	}()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		err := out.Close()
		if err != nil {
			s.log.Error("Failed to close file", "error", err)
		}
	}()

	_, err = io.Copy(out, src)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

func (s *FileService) resize(input io.Reader, output io.Writer, width, height int) error {
	img, err := jpeg.Decode(input)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	resized := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	if err := jpeg.Encode(output, resized, &jpeg.Options{Quality: 90}); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

func (s *FileService) convertAudioToWebm(inputPath, outputPath string) error {
	err := ffmpeg.Input(inputPath).
		Output(outputPath, ffmpeg.KwArgs{
			"c:a": "libvorbis",
			"af":  "dynaudnorm",
			"f":   "webm",
		}).
		OverWriteOutput().
		Run()

	if err != nil {
		return fmt.Errorf("failed to convert to webm: %w", err)
	}

	return nil
}

func (s *FileService) normalizeAudio(inputPath, outputPath string) error {
	err := ffmpeg.Input(inputPath).
		Output(outputPath, ffmpeg.KwArgs{
			"af": "dynaudnorm",
			"f":  "webm",
		}).
		OverWriteOutput().
		Run()

	if err != nil {
		return fmt.Errorf("failed to normalize audio: %w", err)
	}

	return nil
}

func (s *FileService) getAudioDuration(filePath string) (int, error) {
	probe, err := ffmpeg.Probe(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to probe audio file: %w", err)
	}

	var metadata struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	if err := json.Unmarshal([]byte(probe), &metadata); err != nil {
		return 0, fmt.Errorf("failed to parse metadata: %w", err)
	}

	duration, err := strconv.ParseFloat(metadata.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return int(duration), nil
}
