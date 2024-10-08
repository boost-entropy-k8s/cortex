package tsdb

import (
	"flag"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alecthomas/units"
	"github.com/go-kit/log"
	"github.com/golang/snappy"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thanos-io/objstore"
	"github.com/thanos-io/thanos/pkg/block"
	"github.com/thanos-io/thanos/pkg/block/metadata"
	"github.com/thanos-io/thanos/pkg/cache"
	"github.com/thanos-io/thanos/pkg/cacheutil"
	"github.com/thanos-io/thanos/pkg/model"
	storecache "github.com/thanos-io/thanos/pkg/store/cache"
)

var (
	supportedMetadataCacheBackends = []string{CacheBackendMemcached, CacheBackendRedis}

	errUnsupportedChunkCacheBackend = errors.New("unsupported chunk cache backend")
)

const (
	CacheBackendMemcached = "memcached"
	CacheBackendRedis     = "redis"
	CacheBackendInMemory  = "inmemory"
)

type MetadataCacheBackend struct {
	Backend   string                `yaml:"backend"`
	Memcached MemcachedClientConfig `yaml:"memcached"`
	Redis     RedisClientConfig     `yaml:"redis"`
}

// Validate the config.
func (cfg *MetadataCacheBackend) Validate() error {
	switch cfg.Backend {
	case CacheBackendMemcached:
		return cfg.Memcached.Validate()
	case CacheBackendRedis:
		return cfg.Redis.Validate()
	case "":
	default:
		return fmt.Errorf("unsupported cache backend: %s", cfg.Backend)
	}
	return nil
}

type ChunkCacheBackend struct {
	Backend   string                   `yaml:"backend"`
	InMemory  InMemoryChunkCacheConfig `yaml:"inmemory"`
	Memcached MemcachedClientConfig    `yaml:"memcached"`
	Redis     RedisClientConfig        `yaml:"redis"`
}

// Validate the config.
func (cfg *ChunkCacheBackend) Validate() error {
	switch cfg.Backend {
	case CacheBackendMemcached:
		return cfg.Memcached.Validate()
	case CacheBackendRedis:
		return cfg.Redis.Validate()
	case CacheBackendInMemory, "":
	default:
		return errUnsupportedChunkCacheBackend
	}
	return nil
}

type ChunksCacheConfig struct {
	ChunkCacheBackend `yaml:",inline"`

	SubrangeSize        int64         `yaml:"subrange_size"`
	MaxGetRangeRequests int           `yaml:"max_get_range_requests"`
	AttributesTTL       time.Duration `yaml:"attributes_ttl"`
	SubrangeTTL         time.Duration `yaml:"subrange_ttl"`
}

func (cfg *ChunksCacheConfig) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&cfg.Backend, prefix+"backend", "", fmt.Sprintf("Backend for chunks cache, if not empty. Supported values: %s, %s, %s, and '' (disable).", CacheBackendMemcached, CacheBackendRedis, CacheBackendInMemory))

	cfg.Memcached.RegisterFlagsWithPrefix(f, prefix+"memcached.")
	cfg.Redis.RegisterFlagsWithPrefix(f, prefix+"redis.")
	cfg.InMemory.RegisterFlagsWithPrefix(f, prefix+"inmemory.")

	f.Int64Var(&cfg.SubrangeSize, prefix+"subrange-size", 16000, "Size of each subrange that bucket object is split into for better caching.")
	f.IntVar(&cfg.MaxGetRangeRequests, prefix+"max-get-range-requests", 3, "Maximum number of sub-GetRange requests that a single GetRange request can be split into when fetching chunks. Zero or negative value = unlimited number of sub-requests.")
	f.DurationVar(&cfg.AttributesTTL, prefix+"attributes-ttl", 168*time.Hour, "TTL for caching object attributes for chunks.")
	f.DurationVar(&cfg.SubrangeTTL, prefix+"subrange-ttl", 24*time.Hour, "TTL for caching individual chunks subranges.")
}

func (cfg *ChunksCacheConfig) Validate() error {
	return cfg.ChunkCacheBackend.Validate()
}

type InMemoryChunkCacheConfig struct {
	MaxSizeBytes uint64 `yaml:"max_size_bytes"`
}

func (cfg *InMemoryChunkCacheConfig) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.Uint64Var(&cfg.MaxSizeBytes, prefix+"max-size-bytes", uint64(1*units.Gibibyte), "Maximum size in bytes of in-memory chunk cache used to speed up chunk lookups (shared between all tenants).")
}

func (cfg *InMemoryChunkCacheConfig) toInMemoryChunkCacheConfig() cache.InMemoryCacheConfig {
	maxCacheSize := model.Bytes(cfg.MaxSizeBytes)

	// Calculate the max item size.
	maxItemSize := defaultMaxItemSize
	if maxItemSize > maxCacheSize {
		maxItemSize = maxCacheSize
	}

	return cache.InMemoryCacheConfig{
		MaxSize:     maxCacheSize,
		MaxItemSize: maxItemSize,
	}
}

type MetadataCacheConfig struct {
	MetadataCacheBackend `yaml:",inline"`

	TenantsListTTL          time.Duration `yaml:"tenants_list_ttl"`
	TenantBlocksListTTL     time.Duration `yaml:"tenant_blocks_list_ttl"`
	ChunksListTTL           time.Duration `yaml:"chunks_list_ttl"`
	MetafileExistsTTL       time.Duration `yaml:"metafile_exists_ttl"`
	MetafileDoesntExistTTL  time.Duration `yaml:"metafile_doesnt_exist_ttl"`
	MetafileContentTTL      time.Duration `yaml:"metafile_content_ttl"`
	MetafileMaxSize         int           `yaml:"metafile_max_size_bytes"`
	MetafileAttributesTTL   time.Duration `yaml:"metafile_attributes_ttl"`
	BlockIndexAttributesTTL time.Duration `yaml:"block_index_attributes_ttl"`
	BucketIndexContentTTL   time.Duration `yaml:"bucket_index_content_ttl"`
	BucketIndexMaxSize      int           `yaml:"bucket_index_max_size_bytes"`
}

func (cfg *MetadataCacheConfig) RegisterFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	f.StringVar(&cfg.Backend, prefix+"backend", "", fmt.Sprintf("Backend for metadata cache, if not empty. Supported values: %s, and '' (disable).", strings.Join(supportedMetadataCacheBackends, ", ")))

	cfg.Memcached.RegisterFlagsWithPrefix(f, prefix+"memcached.")
	cfg.Redis.RegisterFlagsWithPrefix(f, prefix+"redis.")

	f.DurationVar(&cfg.TenantsListTTL, prefix+"tenants-list-ttl", 15*time.Minute, "How long to cache list of tenants in the bucket.")
	f.DurationVar(&cfg.TenantBlocksListTTL, prefix+"tenant-blocks-list-ttl", 5*time.Minute, "How long to cache list of blocks for each tenant.")
	f.DurationVar(&cfg.ChunksListTTL, prefix+"chunks-list-ttl", 24*time.Hour, "How long to cache list of chunks for a block.")
	f.DurationVar(&cfg.MetafileExistsTTL, prefix+"metafile-exists-ttl", 2*time.Hour, "How long to cache information that block metafile exists. Also used for user deletion mark file.")
	f.DurationVar(&cfg.MetafileDoesntExistTTL, prefix+"metafile-doesnt-exist-ttl", 5*time.Minute, "How long to cache information that block metafile doesn't exist. Also used for user deletion mark file.")
	f.DurationVar(&cfg.MetafileContentTTL, prefix+"metafile-content-ttl", 24*time.Hour, "How long to cache content of the metafile.")
	f.IntVar(&cfg.MetafileMaxSize, prefix+"metafile-max-size-bytes", 1*1024*1024, "Maximum size of metafile content to cache in bytes. Caching will be skipped if the content exceeds this size. This is useful to avoid network round trip for large content if the configured caching backend has an hard limit on cached items size (in this case, you should set this limit to the same limit in the caching backend).")
	f.DurationVar(&cfg.MetafileAttributesTTL, prefix+"metafile-attributes-ttl", 168*time.Hour, "How long to cache attributes of the block metafile.")
	f.DurationVar(&cfg.BlockIndexAttributesTTL, prefix+"block-index-attributes-ttl", 168*time.Hour, "How long to cache attributes of the block index.")
	f.DurationVar(&cfg.BucketIndexContentTTL, prefix+"bucket-index-content-ttl", 5*time.Minute, "How long to cache content of the bucket index.")
	f.IntVar(&cfg.BucketIndexMaxSize, prefix+"bucket-index-max-size-bytes", 1*1024*1024, "Maximum size of bucket index content to cache in bytes. Caching will be skipped if the content exceeds this size. This is useful to avoid network round trip for large content if the configured caching backend has an hard limit on cached items size (in this case, you should set this limit to the same limit in the caching backend).")
}

func (cfg *MetadataCacheConfig) Validate() error {
	return cfg.MetadataCacheBackend.Validate()
}

func CreateCachingBucket(chunksConfig ChunksCacheConfig, metadataConfig MetadataCacheConfig, matchers Matchers, bkt objstore.InstrumentedBucket, logger log.Logger, reg prometheus.Registerer) (objstore.InstrumentedBucket, error) {
	cfg := cache.NewCachingBucketConfig()
	cachingConfigured := false

	chunksCache, err := createChunkCache("chunks-cache", &chunksConfig.ChunkCacheBackend, logger, reg)
	if err != nil {
		return nil, errors.Wrapf(err, "chunks-cache")
	}
	if chunksCache != nil {
		cachingConfigured = true
		chunksCache = cache.NewTracingCache(chunksCache)
		cfg.CacheGetRange("chunks", chunksCache, matchers.GetChunksMatcher(), chunksConfig.SubrangeSize, chunksConfig.AttributesTTL, chunksConfig.SubrangeTTL, chunksConfig.MaxGetRangeRequests)
	}

	metadataCache, err := createMetadataCache("metadata-cache", &metadataConfig.MetadataCacheBackend, logger, reg)
	if err != nil {
		return nil, errors.Wrapf(err, "metadata-cache")
	}
	if metadataCache != nil {
		cachingConfigured = true
		metadataCache = cache.NewTracingCache(metadataCache)

		cfg.CacheExists("metafile", metadataCache, matchers.GetMetafileMatcher(), metadataConfig.MetafileExistsTTL, metadataConfig.MetafileDoesntExistTTL)
		cfg.CacheGet("metafile", metadataCache, matchers.GetMetafileMatcher(), metadataConfig.MetafileMaxSize, metadataConfig.MetafileContentTTL, metadataConfig.MetafileExistsTTL, metadataConfig.MetafileDoesntExistTTL)
		cfg.CacheAttributes("metafile", metadataCache, matchers.GetMetafileMatcher(), metadataConfig.MetafileAttributesTTL)
		cfg.CacheAttributes("block-index", metadataCache, matchers.GetBlockIndexMatcher(), metadataConfig.BlockIndexAttributesTTL)
		cfg.CacheGet("bucket-index", metadataCache, matchers.GetBucketIndexMatcher(), metadataConfig.BucketIndexMaxSize, metadataConfig.BucketIndexContentTTL /* do not cache exist / not exist: */, 0, 0)

		codec := snappyIterCodec{storecache.JSONIterCodec{}}
		cfg.CacheIter("tenants-iter", metadataCache, matchers.GetTenantsIterMatcher(), metadataConfig.TenantsListTTL, codec, "")
		cfg.CacheIter("tenant-blocks-iter", metadataCache, matchers.GetTenantBlocksIterMatcher(), metadataConfig.TenantBlocksListTTL, codec, "")
		cfg.CacheIter("chunks-iter", metadataCache, matchers.GetChunksIterMatcher(), metadataConfig.ChunksListTTL, codec, "")
	}

	if !cachingConfigured {
		// No caching is configured.
		return bkt, nil
	}

	return storecache.NewCachingBucket(bkt, cfg, logger, reg)
}

func createMetadataCache(cacheName string, cacheBackend *MetadataCacheBackend, logger log.Logger, reg prometheus.Registerer) (cache.Cache, error) {
	switch cacheBackend.Backend {
	case "":
		// No caching.
		return nil, nil
	case CacheBackendMemcached:
		var client cacheutil.MemcachedClient
		client, err := cacheutil.NewMemcachedClientWithConfig(logger, cacheName, cacheBackend.Memcached.ToMemcachedClientConfig(), reg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create memcached client")
		}
		return cache.NewMemcachedCache(cacheName, logger, client, reg), nil

	case CacheBackendRedis:
		redisCache, err := cacheutil.NewRedisClientWithConfig(logger, cacheName, cacheBackend.Redis.ToRedisClientConfig(), reg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create redis client")
		}
		return cache.NewRedisCache(cacheName, logger, redisCache, reg), nil

	default:
		return nil, errors.Errorf("unsupported cache type for cache %s: %s", cacheName, cacheBackend.Backend)
	}
}

func createChunkCache(cacheName string, cacheBackend *ChunkCacheBackend, logger log.Logger, reg prometheus.Registerer) (cache.Cache, error) {
	switch cacheBackend.Backend {
	case "":
		// No caching.
		return nil, nil
	case CacheBackendInMemory:
		inMemoryCache, err := cache.NewInMemoryCacheWithConfig(cacheName, logger, reg, cacheBackend.InMemory.toInMemoryChunkCacheConfig())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create in-memory chunk cache")
		}
		return inMemoryCache, nil
	case CacheBackendMemcached:
		var client cacheutil.MemcachedClient
		client, err := cacheutil.NewMemcachedClientWithConfig(logger, cacheName, cacheBackend.Memcached.ToMemcachedClientConfig(), reg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create memcached client")
		}
		return cache.NewMemcachedCache(cacheName, logger, client, reg), nil

	case CacheBackendRedis:
		redisCache, err := cacheutil.NewRedisClientWithConfig(logger, cacheName, cacheBackend.Redis.ToRedisClientConfig(), reg)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create redis client")
		}
		return cache.NewRedisCache(cacheName, logger, redisCache, reg), nil

	default:
		return nil, errors.Errorf("unsupported cache type for cache %s: %s", cacheName, cacheBackend.Backend)
	}
}

type Matchers struct {
	matcherMap map[string]func(string) bool
}

func NewMatchers() Matchers {
	matcherMap := make(map[string]func(string) bool)
	matcherMap["chunks"] = isTSDBChunkFile
	matcherMap["metafile"] = isMetaFile
	matcherMap["block-index"] = isBlockIndexFile
	matcherMap["bucket-index"] = isBucketIndexFiles
	matcherMap["tenants-iter"] = isTenantsDir
	matcherMap["tenant-blocks-iter"] = isTenantBlocksDir
	matcherMap["chunks-iter"] = isChunksDir
	return Matchers{
		matcherMap: matcherMap,
	}
}

func (m *Matchers) SetMetaFileMatcher(f func(string) bool) {
	m.matcherMap["metafile"] = f
}

func (m *Matchers) SetChunksMatcher(f func(string) bool) {
	m.matcherMap["chunks"] = f
}

func (m *Matchers) SetBlockIndexMatcher(f func(string) bool) {
	m.matcherMap["block-index"] = f
}

func (m *Matchers) SetBucketIndexMatcher(f func(string) bool) {
	m.matcherMap["bucket-index"] = f
}

func (m *Matchers) SetTenantsIterMatcher(f func(string) bool) {
	m.matcherMap["tenants-iter"] = f
}

func (m *Matchers) SetTenantBlocksIterMatcher(f func(string) bool) {
	m.matcherMap["tenant-blocks-iter"] = f
}

func (m *Matchers) SetChunksIterMatcher(f func(string) bool) {
	m.matcherMap["chunks-iter"] = f
}

func (m *Matchers) GetChunksMatcher() func(string) bool {
	return m.matcherMap["chunks"]
}

func (m *Matchers) GetMetafileMatcher() func(string) bool {
	return m.matcherMap["metafile"]
}

func (m *Matchers) GetBlockIndexMatcher() func(string) bool {
	return m.matcherMap["block-index"]
}

func (m *Matchers) GetBucketIndexMatcher() func(string) bool {
	return m.matcherMap["bucket-index"]
}

func (m *Matchers) GetTenantsIterMatcher() func(string) bool {
	return m.matcherMap["tenants-iter"]
}

func (m *Matchers) GetTenantBlocksIterMatcher() func(string) bool {
	return m.matcherMap["tenant-blocks-iter"]
}

func (m *Matchers) GetChunksIterMatcher() func(string) bool {
	return m.matcherMap["chunks-iter"]
}

var chunksMatcher = regexp.MustCompile(`^.*/chunks/\d+$`)

func isTSDBChunkFile(name string) bool { return chunksMatcher.MatchString(name) }

func isMetaFile(name string) bool {
	return strings.HasSuffix(name, "/"+metadata.MetaFilename) || strings.HasSuffix(name, "/"+metadata.DeletionMarkFilename) || strings.HasSuffix(name, "/"+TenantDeletionMarkFile)
}

func isBlockIndexFile(name string) bool {
	// Ensure the path ends with "<block id>/<index filename>".
	if !strings.HasSuffix(name, "/"+block.IndexFilename) {
		return false
	}

	_, err := ulid.Parse(filepath.Base(filepath.Dir(name)))
	return err == nil
}

func isBucketIndexFiles(name string) bool {
	// TODO can't reference bucketindex because of a circular dependency. To be fixed.
	return strings.HasSuffix(name, "/bucket-index.json.gz") || strings.HasSuffix(name, "/bucket-index-sync-status.json")
}

func isTenantsDir(name string) bool {
	return name == ""
}

var tenantDirMatcher = regexp.MustCompile("^[^/]+/?$")

func isTenantBlocksDir(name string) bool {
	return tenantDirMatcher.MatchString(name)
}

func isChunksDir(name string) bool {
	return strings.HasSuffix(name, "/chunks")
}

type snappyIterCodec struct {
	cache.IterCodec
}

func (i snappyIterCodec) Encode(files []string) ([]byte, error) {
	b, err := i.IterCodec.Encode(files)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

func (i snappyIterCodec) Decode(cachedData []byte) ([]string, error) {
	b, err := snappy.Decode(nil, cachedData)
	if err != nil {
		return nil, errors.Wrap(err, "snappyIterCodec")
	}
	return i.IterCodec.Decode(b)
}
