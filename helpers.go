/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Wednesday 08 January 2020 - 23:56:35
** @Filename:				Helpers.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Saturday 22 February 2020 - 11:42:28
*******************************************************************************/

package			main

import			"os"
import			"io"
import			"fmt"
import			"mime"
import			"bytes"
import			"crypto/rand"
import			"github.com/microgolang/logs"

func	generateUUID(n uint32) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if (err != nil) {
		return ``, err
	}

	uuid := fmt.Sprintf(
		"%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x-%x",
		b[0:2], b[2:4], b[4:6], b[6:8], b[8:10], b[10:12], b[12:14], b[14:16], b[16:18], b[18:20], b[20:22], b[22:24], b[24:26], b[26:28], b[28:30], b[30:32], 
	)
	return uuid, nil
}

func	getExtFromMime(contentType string) string {
	ext, err := mime.ExtensionsByType(contentType)
	if (len(ext) == 0 || err != nil) {
		return ``
	}
	return ext[0]
}

func	storePicture(blob []byte, contentType, size string) string {
	imageName, err := generateUUID(32)
	if (err != nil) {
		logs.Error(`Impossible to generate UUID`, err)
		return ``
	}

	os.MkdirAll(`/pictures/` + size + `/`, os.ModePerm)
	filePath := `/pictures/` + size + `/` + imageName + getExtFromMime(contentType)
	f, err := os.Create(filePath)
	if (err != nil) {
		logs.Error(`Impossible to create file`, err)
		return ``
	}
	defer f.Close()
    _, err = io.Copy(f, bytes.NewReader(blob))
	if (err != nil) {
		logs.Error(`Impossible to copy data to file`, err)
		return ``
	}
	return (filePath)
}

func	removePicture(path string) error {
	err := os.RemoveAll(path)
	return err
}