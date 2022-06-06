package store

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/mspraggs/hoard/internal/store/models"
)

// Delete erases the specified file upload from the relevant backend storage
// bucket.
func (c *AWSClient) Delete(ctx context.Context, fileUpload *models.FileUpload) error {
	input := fileUpload.ToDeleteObjectInput()

	if _, err := c.bc.DeleteObject(ctx, (*s3.DeleteObjectInput)(input)); err != nil {
		return fmt.Errorf("unable to delete object: %w", err)
	}

	return nil
}
