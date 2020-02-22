/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Monday 13 January 2020 - 20:00:58
** @Filename:				resizer.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Friday 21 February 2020 - 17:39:45
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
	_ = imgConfig
	_ = fileType
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
func	storePicture(blob []byte, contentType, size string) (string, error) {
	filePath := storeDecryptedThumbnail(blob, contentType, size)

	return filePath, nil
}