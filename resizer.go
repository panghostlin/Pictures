/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Monday 13 January 2020 - 20:00:58
** @Filename:				resizer.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Sunday 09 February 2020 - 17:29:05
*******************************************************************************/

package			main

import			"bytes"
import			"image"
import			"image/jpeg"
import			"image/png"
import			"image/gif"
import			"golang.org/x/image/bmp"
import			"golang.org/x/image/tiff"
import			"github.com/chai2010/webp"
import			"github.com/nfnt/resize"
import			"github.com/microgolang/logs"
import			"mime"
import			"errors"
import			"context"
import			"gitlab.com/betterpiwigo/sdk/Keys"

func	getExtFromMime(contentType string) string {
	ext, err := mime.ExtensionsByType(contentType)
	if (len(ext) == 0 || err != nil) {
		return ``
	}
	return ext[0]
}
func	getImageSize(img image.Image) (uint, uint) {
	b := img.Bounds()
	imgWidth := b.Max.X
	imgHeight := b.Max.Y

	return uint(imgWidth), uint(imgHeight)
}
func	getContentType(blob []byte) {
	imgConfig, fileType, err := image.DecodeConfig(bytes.NewReader(blob))
	if err != nil {
		logs.Error(err)
		return
	}
	logs.Pretty(imgConfig, fileType)
}
func	getBlobSize(blob []byte, contentType string) (uint, uint, error) {
	var	imgConfig image.Config
	var	err	error

	if (contentType == `image/bmp` || contentType == `image/x-windows-bmp`) {
		imgConfig, err = bmp.DecodeConfig(bytes.NewReader(blob))
		if err != nil {
			logs.Error(err)
			return 0, 0, err
		}
		return uint(imgConfig.Width), uint(imgConfig.Height), nil
	} else if (contentType == `image/gif`) {
		imgConfig, err = gif.DecodeConfig(bytes.NewReader(blob))
		if err != nil {
			logs.Error(err)
			return 0, 0, err
		}

		return uint(imgConfig.Width), uint(imgConfig.Height), nil
	} else if (contentType == `image/jpeg` || contentType == `image/pjpeg`) {
		imgConfig, err = jpeg.DecodeConfig(bytes.NewReader(blob))
		if err != nil {
			logs.Error(err)
			return 0, 0, err
		}

		return uint(imgConfig.Width), uint(imgConfig.Height), nil
	} else if (contentType == `image/png`) {
		imgConfig, err = png.DecodeConfig(bytes.NewReader(blob))
		if err != nil {
			logs.Error(err)
			return 0, 0, err
		}
		return uint(imgConfig.Width), uint(imgConfig.Height), nil
	} else if (contentType == `image/webp`) {
		imgConfig, err = webp.DecodeConfig(bytes.NewReader(blob))
		if err != nil {
			logs.Error(err)
			return 0, 0, err
		}

		return uint(imgConfig.Width), uint(imgConfig.Height), nil
	} else if (contentType == `image/tiff`) {
		imgConfig, err = tiff.DecodeConfig(bytes.NewReader(blob))
		if err != nil {
			logs.Error(err)
			return 0, 0, err
		}

		return uint(imgConfig.Width), uint(imgConfig.Height), nil
	}
	return 0, 0, errors.New(`Format not supported`)
}

func	generateThumbnail(encryptedData []byte, contentType string, width, height uint) ([]byte, uint, uint, error) {
	var	img image.Image
	var	err	error

	if (contentType == `image/bmp` || contentType == `image/x-windows-bmp`) {
		img, err = bmp.Decode(bytes.NewReader(encryptedData))
		if err != nil {
			logs.Error(err)
			return nil, 0, 0, err
		}

		thumbnail := resize.Resize(width, height, img, resize.Lanczos3)
		imageWidth, imageHeight := getImageSize(thumbnail)
		
		buf := new(bytes.Buffer)
		err = bmp.Encode(buf, thumbnail)
		if (err != nil) {
			logs.Error(err)
			return nil, 0, 0, err
		}

		buffer := buf.Bytes()
		return buffer, imageWidth, imageHeight, nil
	} else if (contentType == `image/gif`) {
		img, err = gif.Decode(bytes.NewReader(encryptedData))
		if err != nil {
			logs.Error(err)
			return nil, 0, 0, err
		}

		thumbnail := resize.Resize(width, height, img, resize.Lanczos3)
		imageWidth, imageHeight := getImageSize(thumbnail)

		buf := new(bytes.Buffer)
		err = gif.Encode(buf, thumbnail, nil)
		if (err != nil) {
			logs.Error(err)
			return nil, 0, 0, err
		}

		buffer := buf.Bytes()
		return buffer, imageWidth, imageHeight, nil
	} else if (contentType == `image/jpeg` || contentType == `image/pjpeg`) {
		img, err = jpeg.Decode(bytes.NewReader(encryptedData))
		if err != nil {
			logs.Error(err)
			return nil, 0, 0, err
		}

		thumbnail := resize.Resize(width, height, img, resize.Lanczos3)
		imageWidth, imageHeight := getImageSize(thumbnail)

		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, thumbnail, nil)
		if (err != nil) {
			logs.Error(err)
			return nil, 0, 0, err
		}

		buffer := buf.Bytes()
		return buffer, imageWidth, imageHeight, nil
	} else if (contentType == `image/png`) {
		img, err = png.Decode(bytes.NewReader(encryptedData))
		if err != nil {
			logs.Error(err)
			return nil, 0, 0, err
		}

		thumbnail := resize.Resize(width, height, img, resize.Lanczos3)
		imageWidth, imageHeight := getImageSize(thumbnail)

		buf := new(bytes.Buffer)
		err = png.Encode(buf, thumbnail)
		if (err != nil) {
			logs.Error(err)
			return nil, 0, 0, err
		}

		buffer := buf.Bytes()
		return buffer, imageWidth, imageHeight, nil
	} else if (contentType == `image/webp`) {
		img, err = webp.Decode(bytes.NewReader(encryptedData))
		if err != nil {
			logs.Error(err)
			return nil, 0, 0, err
		}

		thumbnail := resize.Resize(width, height, img, resize.Lanczos3)
		imageWidth, imageHeight := getImageSize(thumbnail)

		buf := new(bytes.Buffer)
		err = webp.Encode(buf, thumbnail, &webp.Options{Lossless: true, Quality: 90, Exact: true})
		if (err != nil) {
			logs.Error(err)
			return nil, 0, 0, err
		}

		buffer := buf.Bytes()
		return buffer, imageWidth, imageHeight, nil
	} else if (contentType == `image/tiff`) {
		img, err = tiff.Decode(bytes.NewReader(encryptedData))
		if err != nil {
			logs.Error(err)
			return nil, 0, 0, err
		}

		thumbnail := resize.Resize(width, height, img, resize.Lanczos3)
		imageWidth, imageHeight := getImageSize(thumbnail)

		buf := new(bytes.Buffer)
		err = tiff.Encode(buf, thumbnail, nil)
		if (err != nil) {
			logs.Error(err)
			return nil, 0, 0, err
		}

		buffer := buf.Bytes()
		return buffer, imageWidth, imageHeight, nil
	}
	return nil, 0, 0, errors.New(`Format not supported`)
}
func	createThumbnails(blob []byte, contentType, memberID string, size string, width, height uint) (int, int, string, string, error) {
	var stream keys.KeysService_EncryptPictureClient

	stream, err := clients.keys.EncryptPicture(context.Background())
	if (err != nil) {
		logs.Error("Fail to init stream", err)
		return -1, -1, ``, ``, err
	}
	defer stream.Context().Done()
	
	thumbnail, thumbnailWidth, thumbnailHeight, err := generateThumbnail(blob, contentType, width, height)
	if (err != nil) {
		logs.Error(err)
		return -1, -1, ``, ``, err
	}

	isSuccess, err := encryptPictureSender(stream, thumbnail, memberID)
	if (err != nil || !isSuccess) {
		logs.Error(`Impossible to encrypt image`, err)
		return -1, -1, ``, ``, err
	}

	response, err := encryptPictureReceiver(stream)
	if (err != nil || !response.GetSuccess()) {
		logs.Error(`Impossible to encrypt image`, err)
		return -1, -1, ``, ``, err
	}

	filePath := storeDecryptedThumbnail(response.GetChunk(), contentType, size)

	return int(thumbnailWidth), int(thumbnailHeight), filePath, response.GetKey(), nil
}
func	createOriginal(blob []byte, contentType, memberID string) (int, int, string, string, error) {
	var stream keys.KeysService_EncryptPictureClient

	stream, err := clients.keys.EncryptPicture(context.Background())
	if (err != nil) {
		logs.Error("Fail to init stream", err)
		return -1, -1, ``, ``, err
	}
	defer stream.Context().Done()

	originalWidth, originalHeight, err := getBlobSize(blob, contentType)
	if (err != nil) {
		logs.Error(err)
		return -1, -1, ``, ``, err
	}
	isSuccess, err := encryptPictureSender(stream, blob, memberID)
	if (err != nil || !isSuccess) {
		logs.Error(`Impossible to encrypt image`, err)
		return -1, -1, ``, ``, err
	}

	response, err := encryptPictureReceiver(stream)
	if (err != nil || !response.GetSuccess()) {
		logs.Error(`Impossible to encrypt image`, err)
		return -1, -1, ``, ``, err
	}

	filePath := storeDecryptedThumbnail(response.GetChunk(), contentType, `original`)

	return int(originalWidth), int(originalHeight), filePath, response.GetKey(), nil
}