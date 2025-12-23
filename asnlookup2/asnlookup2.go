package asnlookup2

import (
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LaunchPad-Network/NetPeek/internal/logger"
	"github.com/cespare/xxhash/v2"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"golang.org/x/sync/semaphore"
)

var (
	ErrASNNotFound = errors.New("asn not found")
	ErrInvalidASN  = errors.New("invalid asn format")
)

// ASNRecord 表示ASN记录
type ASNRecord struct {
	ASN   uint32 `json:"asn"`
	Name  string `json:"name"`
	Class string `json:"class"`
	CC    string `json:"cc"`
}

// MetaInfo 元数据信息
type MetaInfo struct {
	Timestamp int64             `json:"timestamp"`
	Version   string            `json:"version"`
	Stats     MetaStats         `json:"stats"`
	HashList  map[string]string `json:"hash_list"`
}

type MetaStats struct {
	GeneratedAt string `json:"generated_at"`
}

// Config 配置
type Config struct {
	DataDir           string         // 数据目录
	MaxMemoryItems    int            // 内存最大缓存项
	UpdateInterval    time.Duration  // 更新间隔
	HTTPTimeout       time.Duration  // HTTP超时
	MaxConcurrent     int            // 最大并发数
	EnableCompression bool           // 是否启用压缩
	Logger            *logrus.Logger // 日志记录器
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		DataDir:           "./asn_data",
		MaxMemoryItems:    500, // 500条内存缓存
		UpdateInterval:    24 * time.Hour,
		HTTPTimeout:       30 * time.Second,
		MaxConcurrent:     10,
		EnableCompression: true,
		Logger:            logger.New("ASN Lookup2"),
	}
}

// Lookup 高性能ASN查询接口
type Lookup interface {
	// Query 查询ASN信息
	Query(asn uint32) (*ASNRecord, error)
	// BatchQuery 批量查询
	BatchQuery(asns []uint32) (map[uint32]*ASNRecord, error)
	// Start 启动服务
	Start(ctx context.Context) error
	// Stop 停止服务
	Stop() error
	// Stats 获取统计信息
	Stats() Stats
	// WaitForReady 等待服务准备就绪
	WaitForReady(ctx context.Context) error
}

// Stats 统计信息
type Stats struct {
	MemoryHits     uint64
	DiskHits       uint64
	Misses         uint64
	MemorySize     int
	DiskSize       int64
	LastUpdateTime time.Time
	UpdateCount    uint64
	IsUpdating     bool
}

// 实现类
type asnLookupImpl struct {
	config Config
	stats  Stats

	// 内存缓存（LRU策略）
	memoryCache map[uint32]*ASNRecord
	memoryMutex sync.RWMutex
	memoryOrder []uint32 // 用于LRU

	// 磁盘缓存（LevelDB）
	diskDB    *leveldb.DB
	diskMutex sync.RWMutex

	// 更新相关
	updateMutex sync.RWMutex
	isUpdating  atomic.Bool
	lastUpdate  atomic.Value // time.Time

	// 就绪状态
	isReady   atomic.Bool
	readyOnce sync.Once
	readyChan chan struct{}

	// 索引文件
	indexFile string
	dataFile  string
	metaFile  string

	// 信号量控制并发
	semaphore *semaphore.Weighted

	// 控制循环
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New 创建新的ASN查询实例
func New(cfg Config) (Lookup, error) {
	if cfg.MaxMemoryItems <= 0 {
		cfg.MaxMemoryItems = 500
	}

	// 创建数据目录
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	impl := &asnLookupImpl{
		config:      cfg,
		memoryCache: make(map[uint32]*ASNRecord),
		memoryOrder: make([]uint32, 0, cfg.MaxMemoryItems),
		indexFile:   filepath.Join(cfg.DataDir, "index-meta.json"),
		dataFile:    filepath.Join(cfg.DataDir, "asns.csv"),
		metaFile:    filepath.Join(cfg.DataDir, "metadata.json"),
		semaphore:   semaphore.NewWeighted(int64(cfg.MaxConcurrent)),
		readyChan:   make(chan struct{}),
	}

	// 初始化磁盘数据库
	if err := impl.initDiskDB(); err != nil {
		return nil, err
	}

	// 加载现有数据
	if err := impl.loadExistingData(); err != nil {
		return nil, err
	}

	return impl, nil
}

// initDiskDB 初始化LevelDB
func (a *asnLookupImpl) initDiskDB() error {
	dbPath := filepath.Join(a.config.DataDir, "leveldb")
	opts := &opt.Options{
		Compression:        opt.SnappyCompression,
		BlockCacheCapacity: 8 * opt.MiB,
		WriteBuffer:        4 * opt.MiB,
	}

	db, err := leveldb.OpenFile(dbPath, opts)
	if err != nil {
		return fmt.Errorf("failed to open leveldb: %w", err)
	}
	a.diskDB = db
	return nil
}

// loadExistingData 加载现有数据
func (a *asnLookupImpl) loadExistingData() error {
	// 检查元数据文件
	if _, err := os.Stat(a.metaFile); os.IsNotExist(err) {
		return nil // 没有数据文件是正常的
	}

	// 加载元数据
	data, err := os.ReadFile(a.metaFile)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var meta MetaInfo
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	a.lastUpdate.Store(time.Unix(meta.Timestamp, 0))
	return nil
}

// Query 查询ASN信息
func (a *asnLookupImpl) Query(asn uint32) (*ASNRecord, error) {
	// 1. 先查内存缓存
	if record := a.queryMemory(asn); record != nil {
		atomic.AddUint64(&a.stats.MemoryHits, 1)
		return record, nil
	}

	// 2. 再查磁盘缓存
	if record, err := a.queryDisk(asn); err == nil {
		atomic.AddUint64(&a.stats.DiskHits, 1)
		// 放入内存缓存
		a.updateMemoryCache(asn, record)
		return record, nil
	}

	atomic.AddUint64(&a.stats.Misses, 1)
	return nil, ErrASNNotFound
}

// queryMemory 查询内存缓存
func (a *asnLookupImpl) queryMemory(asn uint32) *ASNRecord {
	a.memoryMutex.RLock()
	defer a.memoryMutex.RUnlock()

	if record, exists := a.memoryCache[asn]; exists {
		// 更新LRU顺序
		a.updateLRUOrder(asn)
		return record
	}
	return nil
}

// queryDisk 查询磁盘缓存
func (a *asnLookupImpl) queryDisk(asn uint32) (*ASNRecord, error) {
	a.diskMutex.RLock()
	defer a.diskMutex.RUnlock()

	key := fmt.Sprintf("asn:%d", asn)
	data, err := a.diskDB.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}

	var record ASNRecord
	if a.config.EnableCompression {
		// 如果需要解压
		decompressed, err := decompress(data)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(decompressed, &record); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, err
		}
	}

	return &record, nil
}

// updateMemoryCache 更新内存缓存
func (a *asnLookupImpl) updateMemoryCache(asn uint32, record *ASNRecord) {
	a.memoryMutex.Lock()
	defer a.memoryMutex.Unlock()

	// 如果缓存已满，移除最久未使用的项
	if len(a.memoryCache) >= a.config.MaxMemoryItems && len(a.memoryOrder) > 0 {
		oldest := a.memoryOrder[0]
		delete(a.memoryCache, oldest)
		a.memoryOrder = a.memoryOrder[1:]
	}

	// 添加新记录
	a.memoryCache[asn] = record
	a.memoryOrder = append(a.memoryOrder, asn)
	a.stats.MemorySize = len(a.memoryCache)
}

// updateLRUOrder 更新LRU顺序
func (a *asnLookupImpl) updateLRUOrder(asn uint32) {
	for i, val := range a.memoryOrder {
		if val == asn {
			// 移动到末尾
			a.memoryOrder = append(a.memoryOrder[:i], a.memoryOrder[i+1:]...)
			a.memoryOrder = append(a.memoryOrder, asn)
			break
		}
	}
}

// BatchQuery 批量查询
func (a *asnLookupImpl) BatchQuery(asns []uint32) (map[uint32]*ASNRecord, error) {
	results := make(map[uint32]*ASNRecord)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(asns))

	for _, asn := range asns {
		wg.Add(1)
		go func(asn uint32) {
			defer wg.Done()

			// 使用信号量控制并发
			if err := a.semaphore.Acquire(a.ctx, 1); err != nil {
				errChan <- err
				return
			}
			defer a.semaphore.Release(1)

			record, err := a.Query(asn)
			if err != nil && !errors.Is(err, ErrASNNotFound) {
				errChan <- err
				return
			}

			if record != nil {
				mu.Lock()
				results[asn] = record
				mu.Unlock()
			}
		}(asn)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// Start 启动服务
func (a *asnLookupImpl) Start(ctx context.Context) error {
	a.ctx, a.cancel = context.WithCancel(ctx)

	// 启动初始化协程
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		// 立即检查更新
		a.config.Logger.Info("Fetching initial data...")
		err := a.checkAndUpdate()
		if err != nil {
			a.config.Logger.Errorf("Failed to fetch initial data: %v\n", err)
		} else {
			a.config.Logger.Info("Initial data fetched successfully.")
		}

		// 无论初始化成功还是失败，都标记就绪
		a.readyOnce.Do(func() {
			a.isReady.Store(true)
			close(a.readyChan)
		})
	}()

	// 启动定期更新循环
	a.wg.Add(1)
	go a.updateLoop()

	return nil
}

// updateLoop 更新循环
func (a *asnLookupImpl) updateLoop() {
	defer a.wg.Done()

	ticker := time.NewTicker(a.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.config.Logger.Info("Checking for updates...")
			if err := a.checkAndUpdate(); err != nil {
				a.config.Logger.Warnf("Update failed: %v\n", err)
			}
			a.config.Logger.Info("Update completed.")
		}
	}
}

// WaitForReady 等待服务准备就绪
func (a *asnLookupImpl) WaitForReady(ctx context.Context) error {
	// 如果已经就绪，直接返回
	if a.isReady.Load() {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-a.readyChan:
		a.isReady.Store(true)
		return nil
	}
}

// checkAndUpdate 检查并更新数据
func (a *asnLookupImpl) checkAndUpdate() error {
	if a.isUpdating.Load() {
		return errors.New("update already in progress")
	}

	a.isUpdating.Store(true)
	defer a.isUpdating.Store(false)

	// 下载元数据
	meta, err := a.downloadMeta()
	if err != nil {
		return fmt.Errorf("failed to download meta: %w", err)
	}

	// 检查是否需要更新
	if lastUpdate, ok := a.lastUpdate.Load().(time.Time); ok {
		if meta.Timestamp <= lastUpdate.Unix() {
			return nil // 不需要更新
		}
	}

	// 下载并处理数据
	if err := a.downloadAndProcess(meta); err != nil {
		return fmt.Errorf("failed to process data: %w", err)
	}

	a.lastUpdate.Store(time.Unix(meta.Timestamp, 0))
	atomic.AddUint64(&a.stats.UpdateCount, 1)
	a.stats.LastUpdateTime = time.Unix(meta.Timestamp, 0)

	return nil
}

// downloadMeta 下载元数据
func (a *asnLookupImpl) downloadMeta() (*MetaInfo, error) {
	client := &http.Client{
		Timeout: a.config.HTTPTimeout,
	}

	resp, err := client.Get("https://cdn.akaere.online/https://raw.githubusercontent.com/Alice39s/BGP.Tools-OpenDB/refs/heads/auto-update/asns/index-meta.json?t=" + strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var meta MetaInfo
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// downloadAndProcess 下载并处理数据
func (a *asnLookupImpl) downloadAndProcess(meta *MetaInfo) error {
	// 下载压缩文件
	gzFile := filepath.Join(a.config.DataDir, "asns.csv.gz")
	if err := a.downloadFile("https://cdn.akaere.online/https://raw.githubusercontent.com/Alice39s/BGP.Tools-OpenDB/refs/heads/auto-update/asns/asns.csv.gz?t="+strconv.FormatInt(time.Now().Unix(), 10), gzFile); err != nil {
		return err
	}

	// 验证文件哈希
	if err := a.verifyFileHash(gzFile, meta.HashList["asns.csv.gz"]); err != nil {
		return err
	}

	// 解压文件
	csvFile := filepath.Join(a.config.DataDir, "asns.csv")
	if err := a.decompressFile(gzFile, csvFile); err != nil {
		return err
	}

	// 验证CSV哈希
	if err := a.verifyFileHash(csvFile, meta.HashList["asns.csv"]); err != nil {
		return err
	}

	// 导入数据
	if err := a.importCSV(csvFile); err != nil {
		return err
	}

	// 保存元数据
	return a.saveMetadata(meta)
}

// downloadFile 下载文件
func (a *asnLookupImpl) downloadFile(url, dest string) error {
	client := &http.Client{
		Timeout: a.config.HTTPTimeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// verifyFileHash 验证文件哈希
func (a *asnLookupImpl) verifyFileHash(filepath, expectedHash string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(h.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("hash mismatch for %s: expected %s, got %s",
			filepath, expectedHash, actualHash)
	}

	return nil
}

// decompressFile 解压文件
func (a *asnLookupImpl) decompressFile(gzFile, dest string) error {
	gz, err := os.Open(gzFile)
	if err != nil {
		return err
	}
	defer gz.Close()

	reader, err := gzip.NewReader(gz)
	if err != nil {
		return err
	}
	defer reader.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	return err
}

// importCSV 导入CSV数据
func (a *asnLookupImpl) importCSV(csvFile string) error {
	file, err := os.Open(csvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ','
	reader.LazyQuotes = true

	// 跳过标题行
	if _, err := reader.Read(); err != nil {
		return err
	}

	batchSize := 1000
	batch := make([]*ASNRecord, 0, batchSize)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if len(record) < 4 {
			continue // 跳过格式错误的行
		}

		// 解析ASN
		asnStr := strings.TrimPrefix(record[0], "AS")
		asn, err := strconv.ParseUint(asnStr, 10, 32)
		if err != nil {
			continue // 跳过无效的ASN
		}

		asnRecord := &ASNRecord{
			ASN:   uint32(asn),
			Name:  record[1],
			Class: record[2],
			CC:    record[3],
		}

		batch = append(batch, asnRecord)

		// 批量写入
		if len(batch) >= batchSize {
			if err := a.writeBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	// 写入剩余记录
	if len(batch) > 0 {
		return a.writeBatch(batch)
	}

	return nil
}

// writeBatch 批量写入数据
func (a *asnLookupImpl) writeBatch(records []*ASNRecord) error {
	a.diskMutex.Lock()
	defer a.diskMutex.Unlock()

	batch := new(leveldb.Batch)

	for _, record := range records {
		key := fmt.Sprintf("asn:%d", record.ASN)

		data, err := json.Marshal(record)
		if err != nil {
			return err
		}

		if a.config.EnableCompression {
			compressed, err := compress(data)
			if err != nil {
				return err
			}
			batch.Put([]byte(key), compressed)
		} else {
			batch.Put([]byte(key), data)
		}

		// 同时更新内存中的索引（用于快速查找）
		indexKey := fmt.Sprintf("index:hash:%d", xxhash.Sum64String(key))
		batch.Put([]byte(indexKey), []byte(key))
	}

	return a.diskDB.Write(batch, nil)
}

// saveMetadata 保存元数据
func (a *asnLookupImpl) saveMetadata(meta *MetaInfo) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.metaFile, data, 0644)
}

// Stats 获取统计信息
func (a *asnLookupImpl) Stats() Stats {
	stats := Stats{
		MemoryHits:     atomic.LoadUint64(&a.stats.MemoryHits),
		DiskHits:       atomic.LoadUint64(&a.stats.DiskHits),
		Misses:         atomic.LoadUint64(&a.stats.Misses),
		MemorySize:     a.stats.MemorySize,
		LastUpdateTime: a.stats.LastUpdateTime,
		UpdateCount:    atomic.LoadUint64(&a.stats.UpdateCount),
		IsUpdating:     a.isUpdating.Load(),
	}

	// 获取磁盘大小
	if info, err := a.diskDB.GetProperty("leveldb.stats"); err == nil {
		stats.DiskSize = int64(len(info))
	}

	return stats
}

// Stop 停止服务
func (a *asnLookupImpl) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	a.wg.Wait()

	if a.diskDB != nil {
		return a.diskDB.Close()
	}
	return nil
}

// compress 压缩数据
func compress(data []byte) ([]byte, error) {
	// 简单的压缩实现，可以使用更高效的算法
	// 这里使用gzip进行演示
	var buf strings.Builder
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

// decompress 解压数据
func decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// Helper function to parse ASN string
func ParseASN(asnStr string) (uint32, error) {
	asnStr = strings.TrimSpace(strings.ToUpper(asnStr))
	if !strings.HasPrefix(asnStr, "AS") {
		return 0, ErrInvalidASN
	}

	asn, err := strconv.ParseUint(asnStr[2:], 10, 32)
	if err != nil {
		return 0, ErrInvalidASN
	}

	return uint32(asn), nil
}
