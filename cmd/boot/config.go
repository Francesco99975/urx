package boot

import (
	"fmt"
	"net"
	"os"

	"regexp"

	"github.com/Francesco99975/urx/internal/enums"
	"github.com/sixafter/nanoid"
)

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

var dsnRegex = regexp.MustCompile(`^postgresql:\/\/([a-zA-Z0-9._%+-]+):([^@]+)@([a-zA-Z0-9.-]+):(\d+)\/([a-zA-Z0-9._-]+)\?sslmode=(disable|require|verify-ca|verify-full)$`)

func isValidDSN(dsn string) bool {
	return dsnRegex.MatchString(dsn)
}

type Config struct {
	Port  string
	Host  string
	GoEnv enums.Environment

	DSN string

	NTFY         string
	NTFYToken    string
	URL          string
	MetricSecret string
	Prometheus   string
}

var Environment = &Config{}

const alphabet = "ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz-2356789"

var Slugerr nanoid.Interface

func LoadEnvVariables() error {

	if !enums.IsEnvironmentValid(os.Getenv("GO_ENV")) {
		return fmt.Errorf("invalid environment variable: %s", os.Getenv("GO_ENV"))
	}

	Environment.Port = os.Getenv("PORT")
	Environment.Host = os.Getenv("HOST")
	Environment.GoEnv = enums.GetEnvironmentFromString(os.Getenv("GO_ENV"))

	Environment.DSN = os.Getenv("DSN")
	if !isValidDSN(Environment.DSN) {
		return fmt.Errorf("invalid DSN: %s", Environment.DSN)
	}

	Environment.NTFY = os.Getenv("NTFY")
	Environment.NTFYToken = os.Getenv("NTFY_TOKEN")
	Environment.MetricSecret = os.Getenv("METRIC_SECRET")
	Environment.Prometheus = os.Getenv("PROMETHEUS")
	if Environment.GoEnv == enums.Environments.DEVELOPMENT {
		localIP := getLocalIP()
		Environment.URL = fmt.Sprintf("http://%s:%s", localIP, Environment.Port)
	} else {
		Environment.URL = fmt.Sprintf("https://%s", Environment.Host)
	}

	var err error
	Slugerr, err = nanoid.NewGenerator(nanoid.WithAlphabet(alphabet), nanoid.WithLengthHint(7))
	if err != nil {
		return err
	}

	return nil
}
