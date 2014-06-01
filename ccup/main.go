// Copyright 2014 Tamás Gulácsi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main of ccup is a simple uploading client for cloduconvert.org
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tgulacsi/cloudconvert"
	"gopkg.in/inconshreveable/log15.v2"
)

const ccAPIkeyEnvName = "CLOUDCONVERT_APIKEY"

func main() {
	flagVerbose := flag.Bool("v", false, "verbose logging")
	flagAPIKey := flag.String("apikey", "", "api key (this, or "+ccAPIkeyEnvName+" env needed)")
	flagFromFormat := flag.String("fromfmt", "", "from format (optional, will from input file name)")
	flagToFormat := flag.String("tofmt", "", "to format - this or a second arg (destination filename) is needed")
	flagWaitDur := flag.Duration("wait", time.Second, "wait duration")
	flag.Parse()

	log15.Root().SetHandler(log15.StderrHandler)
	if !*flagVerbose {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlInfo, log15.StderrHandler))
	}
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "A filename to upload is needed.\n")
		os.Exit(1)
	}
	fromFile := flag.Arg(0)
	toFormat := *flagToFormat
	toFile := flag.Arg(1)
	if toFile == "" {
		if toFormat == "" {
			log15.Crit("-tofmt or a destination filename (second arg) is needed!")
			os.Exit(2)
		}
		ext := filepath.Ext(fromFile)
		if ext == "" {
			toFile = fromFile + "." + toFormat
		} else {
			toFile = fromFile[:len(fromFile)-len(ext)] + "." + toFormat
		}
	}
	apiKey := *flagAPIKey
	if apiKey == "" {
		apiKey = os.Getenv(ccAPIkeyEnvName)
	}
	if apiKey == "" {
		log15.Crit("API key is needed!")
		os.Exit(3)
	}

	if err := convert(apiKey, fromFile, toFile, *flagFromFormat, toFormat, *flagWaitDur); err != nil {
		log15.Crit("ERROR", "error", err)
		os.Exit(4)
	}
}

func convert(apiKey, fromFile, toFile, fromFormat, toFormat string, wait time.Duration) error {
	c, err := cloudconvert.NewConversion(apiKey, fromFile, toFile, fromFormat, toFormat)
	if err != nil {
		return fmt.Errorf("NewConversion: %v", err)
	}
	log15.Info("process", "URL", c.Process.URL)
	if err = c.Start(); err != nil {
		return fmt.Errorf("Start: %v", err)
	}
	log15.Info("Uploaded.")
	if err = c.Wait(wait); err != nil {
		return err
	}
	log15.Info("Done.")
	return c.Save()
}
