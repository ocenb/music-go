package cloudinary

import (
	"context"
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type Client struct {
	client *cloudinary.Cloudinary
}

func New(cloudName, apiKey, apiSecret string) (*Client, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	return &Client{client: cld}, nil
}

func (c *Client) Upload(ctx context.Context, filePath, fileName, resourceType, folder string) error {
	_, err := c.client.Upload.Upload(ctx, filePath, uploader.UploadParams{
		PublicID:     fileName,
		Folder:       folder,
		ResourceType: resourceType,
	})
	return err
}

func (c *Client) Delete(ctx context.Context, publicID, resourceType string) error {
	_, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: resourceType,
	})
	return err
}
