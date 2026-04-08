package helpers

import (
	"encoding/base64"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCodeBase64(content string) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	base64Str := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + base64Str, nil
}
