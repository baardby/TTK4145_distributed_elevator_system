package elevalgo

/*import (
	"fmt"
	"strings"
	"os"
)

type En int
const (
	En1 En = iota
	En2
	En3
)

func (e *En) UnmarshalText(text string) error {
	switch strings.ToLower(text) {
	case "en1":
		*e = En1
	case "en2":
		*e = En2
	case "en3":
		*e = En3
	default:
		return fmt.Errorf("unkown enum value: %q", text)
	}
	return nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "--") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0][2:])
		value := strings.Join(parts[1:], " ")

		switch key {
			case
		}
	}
}*/
