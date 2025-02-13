// Copyright 2015 The Nho Luong Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !nointerrupts
// +build !nointerrupts

package collector

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	interruptLabelNames = []string{"cpu", "type", "info", "devices"}
)

func (c *interruptsCollector) Update(ch chan<- prometheus.Metric) (err error) {
	interrupts, err := getInterrupts()
	if err != nil {
		return fmt.Errorf("couldn't get interrupts: %w", err)
	}
	for name, interrupt := range interrupts {
		for cpuNo, value := range interrupt.values {
			filterName := name + ";" + interrupt.info + ";" + interrupt.devices
			if c.nameFilter.ignored(filterName) {
				c.logger.Debug("ignoring interrupt name", "filter_name", filterName)
				continue
			}
			fv, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value %s in interrupts: %w", value, err)
			}
			if !c.includeZeros && fv == 0.0 {
				c.logger.Debug("ignoring interrupt with zero value", "filter_name", filterName, "cpu", cpuNo)
				continue
			}
			ch <- c.desc.mustNewConstMetric(fv, strconv.Itoa(cpuNo), name, interrupt.info, interrupt.devices)
		}
	}
	return err
}

type interrupt struct {
	info    string
	devices string
	values  []string
}

func getInterrupts() (map[string]interrupt, error) {
	file, err := os.Open(procFilePath("interrupts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseInterrupts(file)
}

func parseInterrupts(r io.Reader) (map[string]interrupt, error) {
	var (
		interrupts = map[string]interrupt{}
		scanner    = bufio.NewScanner(r)
	)

	if !scanner.Scan() {
		return nil, errors.New("interrupts empty")
	}
	cpuNum := len(strings.Fields(scanner.Text())) // one header per cpu

	for scanner.Scan() {
		// On aarch64 there can be zero space between the name/label
		// and the values, so we need to split on `:` before using
		// strings.Fields() to split on fields.
		group := strings.SplitN(scanner.Text(), ":", 2)
		if len(group) > 1 {
			parts := strings.Fields(group[1])

			if len(parts) < cpuNum+1 { // irq + one column per cpu + details,
				continue // we ignore ERR and MIS for now
			}
			intName := strings.TrimLeft(group[0], " ")
			intr := interrupt{
				values: parts[0:cpuNum],
			}

			if _, err := strconv.Atoi(intName); err == nil { // numeral interrupt
				intr.info = parts[cpuNum]
				intr.devices = strings.Join(parts[cpuNum+1:], " ")
			} else {
				intr.info = strings.Join(parts[cpuNum:], " ")
			}
			interrupts[intName] = intr
		}
	}

	return interrupts, scanner.Err()
}
