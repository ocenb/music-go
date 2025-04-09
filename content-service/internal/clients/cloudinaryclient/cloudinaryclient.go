package cloudinaryclient

import (
	"context"
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryClientInterface interface {
	Upload(ctx context.Context, filePath, fileName, resourceType, folder string) error
	Delete(ctx context.Context, publicID, resourceType string) error
}

type CloudinaryClient struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryClient(cloudName, apiKey, apiSecret string) (CloudinaryClientInterface, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	return &CloudinaryClient{
		client: cld,
	}, nil
}

func (s *CloudinaryClient) Upload(ctx context.Context, filePath, fileName, resourceType, folder string) error {
	_, err := s.client.Upload.Upload(ctx, filePath, uploader.UploadParams{
		PublicID:     fileName,
		Folder:       folder,
		ResourceType: resourceType,
	})
	return err
}

func (s *CloudinaryClient) Delete(ctx context.Context, publicID, resourceType string) error {
	_, err := s.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: resourceType,
	})
	return err
}
