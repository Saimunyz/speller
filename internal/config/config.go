package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SpellerConfig - contains all parametrs for speller configuration
type SpellerConfig struct {
	SentencesPath string  `yaml:"sentences_path"`
	DictPath      string  `yaml:"dict_path"`
	MinWordFreq   int     `yaml:"min_word_freq"`
	MinWordLength int     `yaml:"min_word_length"`
	UnigramWeight float64 `yaml:"unigram_weight"`
	BigramWeight  float64 `yaml:"bigram_weight"`
	TrigramWeight float64 `yaml:"trigram_weight"`
	AutoTrainMode bool    `yaml:"auto_train_mode"`
}

// Config - contains all configuration parameters in config package
type Config struct {
	SpellerConfig SpellerConfig `yaml:"speller_config"`
}

func (o *Config) Validate() error {
	// if o.SpellerConfig.Addr == "" {
	// 	return fmt.Errorf("provide non empty SC_ADDR")
	// }
	if o.SpellerConfig.SentencesPath == "" {
		return fmt.Errorf("you need to set 'sentences_path'")
	}
	if o.SpellerConfig.DictPath == "" {
		return fmt.Errorf("you need to set 'dict_path'")
	}
	if o.SpellerConfig.MinWordLength == 0 {
		return fmt.Errorf("you need to set non zero 'min_word_length'")
	}
	if o.SpellerConfig.MinWordFreq == 0 {
		return fmt.Errorf("you need to set non zero 'min_word_freq'")
	}
	if o.SpellerConfig.UnigramWeight == 0 {
		return fmt.Errorf("you need to set non zero 'unigram_weight'")
	}
	if o.SpellerConfig.BigramWeight == 0 {
		return fmt.Errorf("you need to set non zero 'bigram_weight'")
	}
	if o.SpellerConfig.TrigramWeight == 0 {
		return fmt.Errorf("you need to set non zero 'trigram_weight'")
	}
	return nil
}

// ReadConfigYML - read configurations from file and init Config instance
func ReadConfigYML(filePath string) (cfg *Config, err error) {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return cfg, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return cfg, err
	}

	setDefault(cfg)

	return cfg, nil
}

func setDefault(cfg *Config) (*Config, error) {
	// if cfg.Addr == "" {
	// 	cfg.Addr = ":10000"
	// }
	if cfg.SpellerConfig.MinWordLength == 0 {
		cfg.SpellerConfig.MinWordLength = 3
	}
	if cfg.SpellerConfig.MinWordFreq == 0 {
		cfg.SpellerConfig.MinWordFreq = 5
	}
	if cfg.SpellerConfig.UnigramWeight == 0 {
		cfg.SpellerConfig.UnigramWeight = 100
	}
	if cfg.SpellerConfig.BigramWeight == 0 {
		cfg.SpellerConfig.BigramWeight = 50
	}
	if cfg.SpellerConfig.TrigramWeight == 0 {
		cfg.SpellerConfig.TrigramWeight = 80
	}

	return cfg, cfg.Validate()
}
