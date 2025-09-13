package main

import (
	"context"
	"fmt"
	"github.com/abdddev/go-magistr-lesson1-tpl.git/pkg/models"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL      = "http://srv.msk01.gigacorp.local"
	statsAPIPath   = "/_stats"
	requestTimeout = 5 * time.Second
	tickerTime     = 3 * time.Second
	errMsg         = "Unable to fetch server statistic"
)

var (
	errorCount int
	httpClient = &http.Client{Timeout: requestTimeout}
)

func getMetrics(ctx context.Context) (*models.Metrics, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s"+statsAPIPath, serverURL),
		nil,
	)
	if err != nil {
		return handleError()
	}

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return handleError()
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return handleError()
	}

	metrics, err := parseMetrics(string(body))
	if err != nil {
		return handleError()
	}

	errorCount = 0
	return metrics, nil
}

func handleError() (*models.Metrics, error) {
	errorCount++
	if errorCount >= 3 {
		errorCount = 0
		return nil, fmt.Errorf(errMsg)
	}
	return nil, nil
}

func parseMetrics(line string) (*models.Metrics, error) {
	line = strings.TrimSpace(line)
	fields := strings.Split(line, ",")
	if len(fields) != 7 {
		return nil, fmt.Errorf("unexpected fields count: got %d", len(fields))
	}

	toInt := func(s string) (int, error) {
		v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}

	var m models.Metrics
	var err error
	if m.LoadAvg, err = toInt(fields[0]); err != nil {
		return nil, err
	}
	if m.MemTotalBytes, err = toInt(fields[1]); err != nil {
		return nil, err
	}
	if m.MemUsedBytes, err = toInt(fields[2]); err != nil {
		return nil, err
	}
	if m.DiskTotalBytes, err = toInt(fields[3]); err != nil {
		return nil, err
	}
	if m.DiskUsedBytes, err = toInt(fields[4]); err != nil {
		return nil, err
	}
	if m.NetBandwidthBps, err = toInt(fields[5]); err != nil {
		return nil, err
	}
	if m.NetLoadBps, err = toInt(fields[6]); err != nil {
		return nil, err
	}

	return &m, nil
}

func analyzeMetrics(m *models.Metrics) {
	if m.LoadAvg > 30 {
		fmt.Printf("Load Average is too high: %d\n", m.LoadAvg)
	}

	if m.MemTotalBytes > 0 {
		memPct := float64(m.MemUsedBytes) / float64(m.MemTotalBytes) * 100.0
		if memPct > 80.0 {
			fmt.Printf("Memory usage too high: %.0f%%\n", memPct)
		}
	}

	if m.DiskTotalBytes > 0 {
		usedPct := float64(m.DiskUsedBytes) / float64(m.DiskTotalBytes) * 100.0
		if usedPct > 90.0 {
			freeBytes := m.DiskTotalBytes - m.DiskUsedBytes
			if freeBytes < 0 {
				freeBytes = 0
			}
			freeMB := freeBytes / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeMB)
		}
	}

	if m.NetBandwidthBps > 0 {
		loadPct := float64(m.NetLoadBps) / float64(m.NetBandwidthBps) * 100.0
		if loadPct > 90.0 {
			freeBps := m.NetBandwidthBps - m.NetLoadBps
			if freeBps < 0 {
				freeBps = 0
			}
			freeMbit := float64(freeBps) * 8.0 / 1_000_000.0
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeMbit)
		}
	}
}

func main() {
	ctx := context.Background()

	ticker := time.NewTicker(tickerTime)
	defer ticker.Stop()

	for range ticker.C {
		metrics, err := getMetrics(ctx)
		if err != nil {
			fmt.Println(errMsg)
			continue
		}

		if metrics == nil {
			continue
		}

		analyzeMetrics(metrics)
	}
}
