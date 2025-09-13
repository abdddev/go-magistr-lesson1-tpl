package models

type Metrics struct {
	LoadAvg         int
	MemTotalBytes   int
	MemUsedBytes    int
	DiskTotalBytes  int
	DiskUsedBytes   int
	NetBandwidthBps int
	NetLoadBps      int
}
