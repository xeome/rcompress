package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	dbpath      string
	compressdir string
	maxconcur   int
	human       bool
}

func (c *Config) setDefaults() {
	c.dbpath = "./oxipng.db"
	c.compressdir = "."
	c.maxconcur = 8
	c.human = false
}

func (c *Config) Parse() {
	dbPath := flag.String("dbPath", "./oxipng.db", "Path to the database")
	compressDir := flag.String("compressDir", ".", "Directory to compress")
	maxConcur := flag.Int("maxConcur", 8, "Maximum concurrency")
	human := flag.Bool("human", false, "Human readable format")

	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error finding executable path: %q", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.toml")

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("Error reading config file: %q", err)
		}

		err = toml.Unmarshal(data, &c)
		if err != nil {
			log.Fatalf("Error parsing config file: %q", err)
		}
	}
	flag.Parse()

	if *dbPath != "./oxipng.db" {
		c.dbpath = *dbPath
	}
	if *compressDir != "." {
		c.compressdir = *compressDir
	}
	if *maxConcur != 8 {
		c.maxconcur = *maxConcur
	}
	if *human {
		c.human = *human
	}
}
