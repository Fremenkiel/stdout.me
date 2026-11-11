package dotenv

import (
	"bytes"
	"io"
	"os"
	"strings"
)

func Load() error {
	var filenames []string
	switch os.Getenv("APP_ENV") {
	case "development":
		filenames = []string{".env", ".env.development"}
	default:
		return nil
	}

	for _, filename := range filenames {
		if err := loadFile(filename, true); err != nil {
			return err
		}
	}

	return nil
}

func unmarshalBytes(src []byte) (map[string]string, error) {
	out := make(map[string]string)
	err := parseBytes(src, out)

	return out, err
}

func loadFile(filename string, overload bool) error {
	envMap, err := readFile(filename)
	if err != nil {
		return err
	}

	currentEnv := map[string]bool{}
	rawEnv := os.Environ()
	for _, rawEnvLine := range rawEnv {
		key, _, _ := strings.Cut(rawEnvLine, "=")
		currentEnv[key] = true
	}

	for key, value := range envMap {
		if !currentEnv[key] || overload {
			os.Setenv(key, value)
		}
	}

	return nil
}

func readFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	return parse(file)
}

func parse(r io.Reader) (map[string]string, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}

	return unmarshalBytes(buf.Bytes())
}

