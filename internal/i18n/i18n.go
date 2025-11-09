package i18n

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultDir = "internal/i18n"

// Translator resolves localized strings using dot-separated keys.
type Translator interface {
	T(key string) string
	Lang() string
}

// Manager stores all available translations.
type Manager struct {
	translations map[string]map[string]string
	defaultLang  string
}

// Load loads translations from the default directory.
func Load(defaultLang string) (*Manager, error) {
	return LoadFromDir(defaultDir, defaultLang)
}

// LoadFromDir loads translations from a directory containing YAML files.
func LoadFromDir(dir, defaultLang string) (*Manager, error) {
	catalog, err := parseDir(dir)
	if err != nil {
		return nil, err
	}

	if defaultLang == "" {
		defaultLang = "en"
	}

	if _, ok := catalog[defaultLang]; !ok {
		return nil, fmt.Errorf("i18n: default language %q is missing", defaultLang)
	}

	return &Manager{translations: catalog, defaultLang: defaultLang}, nil
}

// Translator returns a translator for the requested language.
func (m *Manager) Translator(lang string) Translator {
	if m == nil {
		return translator{}
	}

	norm := strings.ToLower(strings.TrimSpace(lang))
	if norm == "" || m.translations[norm] == nil {
		norm = m.defaultLang
	}

	return translator{
		lang:         norm,
		fallback:     m.defaultLang,
		translations: m.translations,
	}
}

// Languages returns all loaded languages.
func (m *Manager) Languages() []string {
	if m == nil {
		return nil
	}

	languages := make([]string, 0, len(m.translations))
	for lang := range m.translations {
		languages = append(languages, lang)
	}
	return languages
}

type translator struct {
	lang         string
	fallback     string
	translations map[string]map[string]string
}

func (t translator) Lang() string {
	return t.lang
}

func (t translator) T(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}

	if value := t.lookup(t.lang, key); value != "" {
		return value
	}

	if value := t.lookup(t.fallback, key); value != "" {
		return value
	}

	return key
}

func (t translator) lookup(lang, key string) string {
	if lang == "" || t.translations == nil {
		return ""
	}

	if entries := t.translations[lang]; entries != nil {
		if value, ok := entries[key]; ok {
			return value
		}
	}

	return ""
}

func parseDir(dir string) (map[string]map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("i18n: read dir %s: %w", dir, err)
	}

	catalog := make(map[string]map[string]string)
	var processed bool

	for _, entry := range entries {
		if entry.IsDir() || !isYAML(entry) {
			continue
		}

		processed = true

		path := filepath.Join(dir, entry.Name())
		fileCatalog, err := parseFile(path)
		if err != nil {
			return nil, err
		}

		for lang, translations := range fileCatalog {
			if _, ok := catalog[lang]; !ok {
				catalog[lang] = make(map[string]string)
			}
			for key, value := range translations {
				catalog[lang][key] = value
			}
		}
	}

	if !processed {
		return nil, fmt.Errorf("i18n: no yaml files found in %s", dir)
	}

	return catalog, nil
}

func isYAML(entry fs.DirEntry) bool {
	name := strings.ToLower(entry.Name())
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}

func parseFile(path string) (map[string]map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("i18n: read file %s: %w", path, err)
	}

	if strings.TrimSpace(string(data)) == "" {
		return map[string]map[string]string{}, nil
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("i18n: parse file %s: %w", path, err)
	}

	catalog := make(map[string]map[string]string)
	for lang, value := range raw {
		langKey := strings.ToLower(strings.TrimSpace(lang))
		if langKey == "" {
			continue
		}

		normalized := toStringMap(value)
		if len(normalized) == 0 {
			continue
		}

		flattened := make(map[string]string)
		flatten("", normalized, flattened)
		if len(flattened) == 0 {
			continue
		}

		catalog[langKey] = flattened
	}

	return catalog, nil
}

func toStringMap(value any) map[string]any {
	switch v := value.(type) {
	case map[string]any:
		return v
	case map[interface{}]any:
		converted := make(map[string]any, len(v))
		for key, item := range v {
			keyStr, ok := key.(string)
			if !ok {
				continue
			}
			converted[keyStr] = item
		}
		return converted
	default:
		return nil
	}
}

func flatten(prefix string, in map[string]any, out map[string]string) {
	for key, value := range in {
		if key == "" {
			continue
		}

		nextKey := key
		if prefix != "" {
			nextKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			out[nextKey] = v
		case map[string]any:
			flatten(nextKey, v, out)
		case map[interface{}]any:
			child := toStringMap(v)
			if len(child) == 0 {
				continue
			}
			flatten(nextKey, child, out)
		}
	}
}
