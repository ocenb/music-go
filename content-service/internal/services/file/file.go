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
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"github.com/ocenb/music-go/content-service/internal/config"
	"github.com/ocenb/music-go/content-service/internal/errs"
)

type Category string

const (
	AudioCategory  Category = "audio"
	ImagesCategory Category = "images"
	tempDir                 = "/app/temp"
	imageSizeLarge          = 250
	imageSizeSmall          = 50
	jpegQuality             = 90
)

type AudioResult struct {
	FileName string
	Duration int
}

type CloudinaryClient interface {
	Upload(ctx context.Context, filePath, fileName, resourceType, folder string) error
	Delete(ctx context.Context, publicID, resourceType string) error
}

type Service struct {
	cloudinary CloudinaryClient
	log        *slog.Logger
	cfg        *config.Config
}

func New(cloudinary CloudinaryClient, log *slog.Logger, cfg *config.Config) *Service {
	return &Service{
		cloudinary: cloudinary,
		log:        log,
		cfg:        cfg,
	}
}

func (s *Service) SaveAudio(ctx context.Context, file *multipart.FileHeader) (*AudioResult, error) {
	if file.Size > s.cfg.AudioFileLimit {
		return nil, errs.ErrAudioFileTooLarge
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
		if err := os.Remove(tempFilePath); err != nil {
			s.log.ErrorContext(ctx, "Failed to remove file", "error", err)
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
		if err := os.Remove(outputFilePath); err != nil {
			s.log.ErrorContext(ctx, "Failed to remove file", "error", err)
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

func (s *Service) SaveImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	if file.Size > s.cfg.ImageFileLimit {
		return "", errs.ErrImageFileTooLarge
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
		if closeErr := src250.Close(); closeErr != nil {
			s.log.ErrorContext(ctx, "Failed to close file", "error", closeErr)
		}
	}()

	out250, err := os.Create(filePath250)
	if err != nil {
		return "", fmt.Errorf("failed to create 250x250 file: %w", err)
	}
	defer func() {
		if closeErr := out250.Close(); closeErr != nil {
			s.log.ErrorContext(ctx, "Failed to close file", "error", closeErr)
		}
	}()

	if resizeErr := s.resize(src250, out250, imageSizeLarge, imageSizeLarge); resizeErr != nil {
		return "", fmt.Errorf("failed to process 250x250 image: %w", resizeErr)
	}
	defer func() {
		if removeErr := os.Remove(filePath250); removeErr != nil {
			s.log.ErrorContext(ctx, "Failed to remove file", "error", removeErr)
		}
	}()

	src50, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := src50.Close(); closeErr != nil {
			s.log.ErrorContext(ctx, "Failed to close file", "error", closeErr)
		}
	}()

	out50, err := os.Create(filePath50)
	if err != nil {
		return "", fmt.Errorf("failed to create 50x50 file: %w", err)
	}
	defer func() {
		if closeErr := out50.Close(); closeErr != nil {
			s.log.ErrorContext(ctx, "Failed to close file", "error", closeErr)
		}
	}()

	if resizeErr := s.resize(src50, out50, imageSizeSmall, imageSizeSmall); resizeErr != nil {
		return "", fmt.Errorf("failed to process 50x50 image: %w", resizeErr)
	}
	defer func() {
		if err := os.Remove(filePath50); err != nil {
			s.log.ErrorContext(ctx, "Failed to remove file", "error", err)
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

func (s *Service) DeleteFile(ctx context.Context, fileName string, category Category) error {
	if category == ImagesCategory {
		s.log.InfoContext(ctx, "Deleting 250x250 image", "fileName", fileName)
		if err := s.cloudinary.Delete(ctx, fmt.Sprintf("images/%s_250x250", fileName), "image"); err != nil {
			return fmt.Errorf("failed to delete 250x250 image: %w", err)
		}
		s.log.InfoContext(ctx, "Deleting 50x50 image", "fileName", fileName)
		if err := s.cloudinary.Delete(ctx, fmt.Sprintf("images/%s_50x50", fileName), "image"); err != nil {
			return fmt.Errorf("failed to delete 50x50 image: %w", err)
		}
	} else {
		s.log.InfoContext(ctx, "Deleting audio", "fileName", fileName)
		if err := s.cloudinary.Delete(ctx, fmt.Sprintf("audio/%s", fileName), "video"); err != nil {
			return fmt.Errorf("failed to delete audio file: %w", err)
		}
	}
	return nil
}

func (s *Service) saveMultipartFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := src.Close(); closeErr != nil {
			s.log.Error("Failed to close file", "error", closeErr)
		}
	}()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			s.log.Error("Failed to close file", "error", closeErr)
		}
	}()

	if _, err = io.Copy(out, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

func (s *Service) resize(input io.Reader, output io.Writer, width, height int) error {
	img, err := jpeg.Decode(input)
	if err != nil {
		return fmt.Errorf("%w", errs.ErrInvalidImageFormat)
	}

	resized := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	if encodeErr := jpeg.Encode(output, resized, &jpeg.Options{Quality: jpegQuality}); encodeErr != nil {
		return fmt.Errorf("failed to encode image: %w", encodeErr)
	}

	return nil
}

func (s *Service) convertAudioToWebm(inputPath, outputPath string) error {
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

func (s *Service) normalizeAudio(inputPath, outputPath string) error {
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

func (s *Service) getAudioDuration(filePath string) (int, error) {
	probe, err := ffmpeg.Probe(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to probe audio file: %w", err)
	}

	var metadata struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	if unmarshalErr := json.Unmarshal([]byte(probe), &metadata); unmarshalErr != nil {
		return 0, fmt.Errorf("failed to parse metadata: %w", unmarshalErr)
	}

	duration, err := strconv.ParseFloat(metadata.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return int(duration), nil
}
