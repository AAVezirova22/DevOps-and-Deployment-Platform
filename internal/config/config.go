package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Config is the project-level deployment contract consumed by deployctl.
type Config struct {
	Name           string
	Namespace      string
	Registry       string
	Image          string
	Tag            string
	Domain         string
	Port           int
	Replicas       int
	Context        string
	Dockerfile     string
	IngressClass   string
	ClusterIssuer  string
	HealthPath     string
	ServiceAccount string
	SecretRefs     []string
	Env            map[string]string
	Resources      ResourceConfig
	Rollback       RollbackConfig
}

type ResourceConfig struct {
	CPURequest    string
	MemoryRequest string
	CPULimit      string
	MemoryLimit   string
}

type RollbackConfig struct {
	Enabled     bool
	FailureText []string
}

// LoadFile reads the small deploykit YAML subset used by this project.
func LoadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	return ParseYAML(string(data))
}

// ParseYAML parses top-level scalar keys plus env and rollback.failureText lists.
func ParseYAML(input string) (Config, error) {
	cfg := Config{Env: map[string]string{}}
	scanner := bufio.NewScanner(strings.NewReader(input))
	section := ""
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		trimmed := stripComment(raw)
		if strings.TrimSpace(trimmed) == "" {
			continue
		}

		indent := len(trimmed) - len(strings.TrimLeft(trimmed, " "))
		line := strings.TrimSpace(trimmed)
		if indent == 0 {
			key, value, ok := splitKeyValue(line)
			if !ok {
				return Config{}, fmt.Errorf("line %d: expected key: value", lineNo)
			}
			section = ""
			switch key {
			case "name":
				cfg.Name = clean(value)
			case "namespace":
				cfg.Namespace = clean(value)
			case "registry":
				cfg.Registry = clean(value)
			case "image":
				cfg.Image = clean(value)
			case "tag":
				cfg.Tag = clean(value)
			case "domain":
				cfg.Domain = clean(value)
			case "port":
				n, err := atoiField(key, value)
				if err != nil {
					return Config{}, err
				}
				cfg.Port = n
			case "replicas":
				n, err := atoiField(key, value)
				if err != nil {
					return Config{}, err
				}
				cfg.Replicas = n
			case "context":
				cfg.Context = clean(value)
			case "dockerfile":
				cfg.Dockerfile = clean(value)
			case "ingressClass":
				cfg.IngressClass = clean(value)
			case "clusterIssuer":
				cfg.ClusterIssuer = clean(value)
			case "healthPath":
				cfg.HealthPath = clean(value)
			case "serviceAccount":
				cfg.ServiceAccount = clean(value)
			case "secretRefs":
				cfg.SecretRefs = parseInlineList(value)
			case "env", "resources", "rollback":
				section = key
			default:
				return Config{}, fmt.Errorf("line %d: unknown config key %q", lineNo, key)
			}
			continue
		}

		switch section {
		case "env":
			key, value, ok := splitKeyValue(line)
			if !ok {
				return Config{}, fmt.Errorf("line %d: expected env key: value", lineNo)
			}
			cfg.Env[key] = clean(value)
		case "resources":
			key, value, ok := splitKeyValue(line)
			if !ok {
				return Config{}, fmt.Errorf("line %d: expected resources key: value", lineNo)
			}
			switch key {
			case "cpuRequest":
				cfg.Resources.CPURequest = clean(value)
			case "memoryRequest":
				cfg.Resources.MemoryRequest = clean(value)
			case "cpuLimit":
				cfg.Resources.CPULimit = clean(value)
			case "memoryLimit":
				cfg.Resources.MemoryLimit = clean(value)
			default:
				return Config{}, fmt.Errorf("line %d: unknown resources key %q", lineNo, key)
			}
		case "rollback":
			key, value, ok := splitKeyValue(line)
			if !ok {
				return Config{}, fmt.Errorf("line %d: expected rollback key: value", lineNo)
			}
			switch key {
			case "enabled":
				b, err := strconv.ParseBool(clean(value))
				if err != nil {
					return Config{}, fmt.Errorf("line %d: rollback.enabled must be true or false", lineNo)
				}
				cfg.Rollback.Enabled = b
			case "failureText":
				cfg.Rollback.FailureText = parseInlineList(value)
			default:
				return Config{}, fmt.Errorf("line %d: unknown rollback key %q", lineNo, key)
			}
		default:
			return Config{}, fmt.Errorf("line %d: indented value without a parent section", lineNo)
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) ApplyDefaults() {
	if c.Namespace == "" {
		c.Namespace = c.Name
	}
	if c.Tag == "" {
		c.Tag = "latest"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.Replicas == 0 {
		c.Replicas = 2
	}
	if c.Context == "" {
		c.Context = "."
	}
	if c.Dockerfile == "" {
		c.Dockerfile = "Dockerfile"
	}
	if c.IngressClass == "" {
		c.IngressClass = "nginx"
	}
	if c.ClusterIssuer == "" {
		c.ClusterIssuer = "letsencrypt-prod"
	}
	if c.HealthPath == "" {
		c.HealthPath = "/"
	}
	if c.ServiceAccount == "" {
		c.ServiceAccount = c.Name
	}
	if c.Resources.CPURequest == "" {
		c.Resources.CPURequest = "100m"
	}
	if c.Resources.MemoryRequest == "" {
		c.Resources.MemoryRequest = "128Mi"
	}
	if c.Resources.CPULimit == "" {
		c.Resources.CPULimit = "500m"
	}
	if c.Resources.MemoryLimit == "" {
		c.Resources.MemoryLimit = "512Mi"
	}
	if len(c.Rollback.FailureText) == 0 {
		c.Rollback.FailureText = []string{"panic:", "fatal", "exception", "crashloopbackoff", "imagepullbackoff"}
	}
}

func (c Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !isDNSLabel(c.Name) {
		return fmt.Errorf("name must be a lowercase DNS label")
	}
	if c.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if !isDNSSubdomain(c.Namespace) {
		return fmt.Errorf("namespace must be a lowercase DNS subdomain")
	}
	if !isDNSLabel(c.ServiceAccount) {
		return fmt.Errorf("serviceAccount must be a lowercase DNS label")
	}
	for _, secretRef := range c.SecretRefs {
		if !isDNSSubdomain(secretRef) {
			return fmt.Errorf("secretRefs contains invalid secret name %q", secretRef)
		}
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.Replicas < 1 {
		return fmt.Errorf("replicas must be at least 1")
	}
	return nil
}

var (
	dnsLabelRE     = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	dnsSubdomainRE = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
)

func isDNSLabel(value string) bool {
	return len(value) <= 63 && dnsLabelRE.MatchString(value)
}

func isDNSSubdomain(value string) bool {
	return len(value) <= 253 && dnsSubdomainRE.MatchString(value)
}

func (c Config) ImageRef() string {
	if c.Image != "" {
		return normalizeImageRef(c.Image)
	}
	base := c.Name
	if c.Registry != "" {
		base = strings.TrimSuffix(c.Registry, "/") + "/" + c.Name
	}
	return normalizeImageRef(base + ":" + c.Tag)
}

func normalizeImageRef(ref string) string {
	name := ref
	suffix := ""
	if idx := strings.Index(name, "@"); idx >= 0 {
		suffix = name[idx:]
		name = name[:idx]
	} else if idx := strings.LastIndex(name, ":"); idx > strings.LastIndex(name, "/") {
		suffix = name[idx:]
		name = name[:idx]
	}
	return strings.ToLower(name) + suffix
}

func StarterYAML(projectName string) string {
	if projectName == "" {
		projectName = "my-service"
	}
	return fmt.Sprintf(`name: %s
namespace: %s-prod
registry: ghcr.io/YOUR_GITHUB_ORG
tag: latest
domain: %s.example.com
port: 8080
replicas: 2
context: .
dockerfile: Dockerfile
ingressClass: nginx
clusterIssuer: letsencrypt-prod
healthPath: /
serviceAccount: %s
secretRefs: []
env:
  APP_ENV: production
resources:
  cpuRequest: 100m
  memoryRequest: 128Mi
  cpuLimit: 500m
  memoryLimit: 512Mi
rollback:
  enabled: true
  failureText: [panic:, fatal, exception, crashloopbackoff, imagepullbackoff]
`, projectName, projectName, projectName, projectName)
}

func splitKeyValue(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
}

func stripComment(line string) string {
	inQuote := rune(0)
	for i, r := range line {
		switch r {
		case '\'', '"':
			if inQuote == 0 {
				inQuote = r
			} else if inQuote == r {
				inQuote = 0
			}
		case '#':
			if inQuote == 0 {
				return line[:i]
			}
		}
	}
	return line
}

func clean(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}

func atoiField(key, value string) (int, error) {
	n, err := strconv.Atoi(clean(value))
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return n, nil
}

func parseInlineList(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = clean(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
