package helpers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Francesco99975/urx/cmd/boot"
	"github.com/labstack/gommon/log"
)

func Notify(topic string, message string) {
	resp, err := http.Post(fmt.Sprintf("%s/%s", boot.Environment.NTFY, topic), "text/plain",
		strings.NewReader(message))

	if err != nil {
		log.Warnf("Failed to send notification: %v", err)
	}

	if resp != nil {
		log.Debugf("Notification sent with status code: %d", resp.StatusCode)
		defer func() { _ = resp.Body.Close() }()
	}
}
