package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Listen    string    `yaml:"listen"`
	Domain    string    `yaml:"domain"`
	Bookmarks Bookmarks `yaml:"bookmarks"`
	Log       *Log      `yaml:"log"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	if c.Listen == "" {
		c.Listen = "0.0.0.0:8080"
	}
	if c.Domain == "" {
		return fmt.Errorf("A domain must be set")
	}
	return nil
}

type Bookmarks []*Bookmark

type Bookmark struct {
	Name               string                 `yaml:"name"`
	Url                string                 `yaml:"url"`
	InsecureSkipVerify bool                   `yaml:"insecure_skip_verify"`
	Proxify            bool                   `yaml:"proxify"`
	LinkerConfig       map[string]interface{} `yaml:"linker_config"`
}

func (c *Bookmark) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Bookmark
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	if c.Name == "" {
		return fmt.Errorf("name cannot be empty for bookmark")
	}
	if c.Url == "" {
		return fmt.Errorf("url cannot be empty for bookmark")
	}
	c.Name = strings.Replace(c.Name, ".", "-", -1)
	c.Name = strings.Replace(c.Name, "_", "-", -1)
	c.Name = strings.Replace(c.Name, " ", "-", -1)
	c.Name = strings.Replace(c.Name, "/", "-", -1)
	return nil
}

type Log struct {
	Level   string `yaml:"level"`
	NoColor bool   `yaml:"no_color"`
	InJson  bool   `yaml:"in_json"`
}

func (c *Log) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Log
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: c.NoColor,
	})
	if c.Level != "" {
		lvl, err := log.ParseLevel(c.Level)
		if err != nil {
			return err
		}
		log.SetLevel(lvl)
	}
	if c.InJson {
		log.SetFormatter(&log.JSONFormatter{})
	}

	return nil
}

func LoadConfig(path string) (Config, error) {
	var cnf Config
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(b, &cnf)
	if err != nil {
		return Config{}, err
	}
	return cnf, nil
}
