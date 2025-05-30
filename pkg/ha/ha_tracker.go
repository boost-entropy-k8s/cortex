package ha

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/prometheus/model/timestamp"

	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/cortexproject/cortex/pkg/ring/kv/codec"
	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/services"
)

const (
	userReplicaGroupUpdateInterval = 30 * time.Second
)

var (
	errNegativeUpdateTimeoutJitterMax = errors.New("HA tracker max update timeout jitter shouldn't be negative")
	errInvalidFailoverTimeout         = "HA Tracker failover timeout (%v) must be at least 1s greater than update timeout - max jitter (%v)"
)

// nolint:revive
type HATrackerLimits interface {
	// MaxHAReplicaGroups returns max number of replica groups that HA tracker should track for a user.
	// Samples from additional replicaGroups are rejected.
	MaxHAReplicaGroups(user string) int
}

// ProtoReplicaDescFactory makes new InstanceDescs
func ProtoReplicaDescFactory() proto.Message {
	return NewReplicaDesc()
}

// NewReplicaDesc returns an empty *ha.ReplicaDesc.
func NewReplicaDesc() *ReplicaDesc {
	return &ReplicaDesc{}
}

// HATrackerConfig contains the configuration require to create a HA Tracker.
// nolint:revive
type HATrackerConfig struct {
	EnableHATracker bool `yaml:"enable_ha_tracker"`
	// We should only update the timestamp if the difference
	// between the stored timestamp and the time we received a sample at
	// is more than this duration.
	UpdateTimeout          time.Duration `yaml:"ha_tracker_update_timeout"`
	UpdateTimeoutJitterMax time.Duration `yaml:"ha_tracker_update_timeout_jitter_max"`
	// We should only failover to accepting samples from a replica
	// other than the replica written in the KVStore if the difference
	// between the stored timestamp and the time we received a sample is
	// more than this duration
	FailoverTimeout time.Duration `yaml:"ha_tracker_failover_timeout"`

	KVStore kv.Config `yaml:"kvstore" doc:"description=Backend storage to use for the ring. Please be aware that memberlist is not supported by the HA tracker since gossip propagation is too slow for HA purposes."`
}

// RegisterFlags adds the flags required to config this to the given FlagSet with a specified prefix
func (cfg *HATrackerConfig) RegisterFlags(f *flag.FlagSet) {
	cfg.RegisterFlagsWithPrefix("", "", f)
}

// RegisterFlags adds the flags required to config this to the given FlagSet.
func (cfg *HATrackerConfig) RegisterFlagsWithPrefix(flagPrefix string, kvPrefix string, f *flag.FlagSet) {
	finalFlagPrefix := ""
	if flagPrefix != "" {
		finalFlagPrefix = flagPrefix
	}

	finalKVPrefix := ""
	if kvPrefix != "" {
		finalKVPrefix = kvPrefix
	}

	f.BoolVar(&cfg.EnableHATracker, finalFlagPrefix+"ha-tracker.enable", false, "Enable the HA tracker so that it can accept data from Prometheus HA replicas gracefully.")
	f.DurationVar(&cfg.UpdateTimeout, finalFlagPrefix+"ha-tracker.update-timeout", 15*time.Second, "Update the timestamp in the KV store for a given cluster/replicaGroup only after this amount of time has passed since the current stored timestamp.")
	f.DurationVar(&cfg.UpdateTimeoutJitterMax, finalFlagPrefix+"ha-tracker.update-timeout-jitter-max", 5*time.Second, "Maximum jitter applied to the update timeout, in order to spread the HA heartbeats over time.")
	f.DurationVar(&cfg.FailoverTimeout, finalFlagPrefix+"ha-tracker.failover-timeout", 30*time.Second, "If we don't receive any data from the accepted replica for a cluster/replicaGroup in this amount of time we will failover to the next replica we receive a sample from. This value must be greater than the update timeout")

	// We want the ability to use different Consul instances for the ring and
	// for HA cluster tracking. We also customize the default keys prefix, in
	// order to not clash with the ring key if they both share the same KVStore
	// backend (ie. run on the same consul cluster).
	cfg.KVStore.RegisterFlagsWithPrefix(finalFlagPrefix+"ha-tracker.", finalKVPrefix+"ha-tracker/", f)
}

// Validate config and returns error on failure
func (cfg *HATrackerConfig) Validate() error {
	if cfg.UpdateTimeoutJitterMax < 0 {
		return errNegativeUpdateTimeoutJitterMax
	}

	minFailureTimeout := cfg.UpdateTimeout + cfg.UpdateTimeoutJitterMax + time.Second
	if cfg.FailoverTimeout < minFailureTimeout {
		return fmt.Errorf(errInvalidFailoverTimeout, cfg.FailoverTimeout, minFailureTimeout)
	}

	// Tracker kv store only supports consul and etcd.
	storeAllowedList := []string{"consul", "etcd"}
	for _, as := range storeAllowedList {
		if cfg.KVStore.Store == as {
			return nil
		}
	}
	return fmt.Errorf("invalid HATracker KV store type: %s", cfg.KVStore.Store)
}

func GetReplicaDescCodec() codec.Proto {
	return codec.NewProtoCodec("replicaDesc", ProtoReplicaDescFactory)
}

// Track the replica we're accepting samples from
// for each HA replica group we know about.
// nolint:revive
type HATracker struct {
	services.Service

	logger              log.Logger
	cfg                 HATrackerConfig
	client              kv.Client
	updateTimeoutJitter time.Duration
	limits              HATrackerLimits

	electedLock   sync.RWMutex
	elected       map[string]ReplicaDesc         // Replicas we are accepting samples from. Key = "user/replicaGroup".
	replicaGroups map[string]map[string]struct{} // Known replica groups with elected replicas per user. First key = user, second key = replica group name (e.g. cluster).

	electedReplicaChanges         *prometheus.CounterVec
	electedReplicaTimestamp       *prometheus.GaugeVec
	electedReplicaPropagationTime prometheus.Histogram
	kvCASCalls                    *prometheus.CounterVec
	userReplicaGroupCount         *prometheus.GaugeVec

	cleanupRuns               prometheus.Counter
	replicasMarkedForDeletion prometheus.Counter
	deletedReplicas           prometheus.Counter
	markingForDeletionsFailed prometheus.Counter

	trackerStatusConfig HATrackerStatusConfig
}

// NewHATracker returns a new HA cluster tracker using either Consul
// or in-memory KV store. Tracker must be started via StartAsync().
func NewHATracker(cfg HATrackerConfig, limits HATrackerLimits, trackerStatusConfig HATrackerStatusConfig, reg prometheus.Registerer, kvNameLabel string, logger log.Logger) (*HATracker, error) {
	var jitter time.Duration
	if cfg.UpdateTimeoutJitterMax > 0 {
		jitter = time.Duration(rand.Int63n(int64(2*cfg.UpdateTimeoutJitterMax))) - cfg.UpdateTimeoutJitterMax
	}

	t := &HATracker{
		logger:              logger,
		cfg:                 cfg,
		updateTimeoutJitter: jitter,
		limits:              limits,
		elected:             map[string]ReplicaDesc{},
		replicaGroups:       map[string]map[string]struct{}{},

		trackerStatusConfig: trackerStatusConfig,

		electedReplicaChanges: promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "ha_tracker_elected_replica_changes_total",
			Help: "The total number of times the elected replica has changed for a user ID/cluster.",
		}, []string{"user", "cluster"}),
		electedReplicaTimestamp: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Name: "ha_tracker_elected_replica_timestamp_seconds",
			Help: "The timestamp stored for the currently elected replica, from the KVStore.",
		}, []string{"user", "cluster"}),
		electedReplicaPropagationTime: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "ha_tracker_elected_replica_change_propagation_time_seconds",
			Help:    "The time it for the distributor to update the replica change.",
			Buckets: prometheus.DefBuckets,
		}),
		kvCASCalls: promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "ha_tracker_kv_store_cas_total",
			Help: "The total number of CAS calls to the KV store for a user ID/cluster.",
		}, []string{"user", "cluster"}),

		userReplicaGroupCount: promauto.With(reg).NewGaugeVec(prometheus.GaugeOpts{
			Name: "ha_tracker_user_replica_group_count",
			Help: "Number of HA replica groups tracked for each user.",
		}, []string{"user"}),

		cleanupRuns: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "ha_tracker_replicas_cleanup_started_total",
			Help: "Number of elected replicas cleanup loops started.",
		}),
		replicasMarkedForDeletion: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "ha_tracker_replicas_cleanup_marked_for_deletion_total",
			Help: "Number of elected replicas marked for deletion.",
		}),
		deletedReplicas: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "ha_tracker_replicas_cleanup_deleted_total",
			Help: "Number of elected replicas deleted from KV store.",
		}),
		markingForDeletionsFailed: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "ha_tracker_replicas_cleanup_delete_failed_total",
			Help: "Number of elected replicas that failed to be marked for deletion, or deleted.",
		}),
	}

	if cfg.EnableHATracker {
		client, err := kv.NewClient(
			cfg.KVStore,
			GetReplicaDescCodec(),
			kv.RegistererWithKVName(reg, kvNameLabel),
			logger,
		)
		if err != nil {
			return nil, err
		}
		t.client = client
	}

	t.Service = services.NewBasicService(nil, t.loop, nil)
	return t, nil
}

// Follows pattern used by ring for WatchKey.
func (c *HATracker) loop(ctx context.Context) error {
	if !c.cfg.EnableHATracker {
		// don't do anything, but wait until asked to stop.
		<-ctx.Done()
		return nil
	}

	// Start cleanup loop. It will stop when context is done.
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.cleanupOldReplicasLoop(ctx)
	}()
	// Start periodic update of user replica group count.
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(userReplicaGroupUpdateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.updateUserReplicaGroupCount()
			case <-ctx.Done():
				return
			}
		}
	}()

	// The KVStore config we gave when creating c should have contained a prefix,
	// which would have given us a prefixed KVStore client. So, we can pass empty string here.
	c.client.WatchPrefix(ctx, "", func(key string, value interface{}) bool {
		replica := value.(*ReplicaDesc)
		user, cluster, keyHasSeparator := strings.Cut(key, "/")

		// Valid key would look like cluster/replica, and a key without a / such as `ring` would be invalid.
		if !keyHasSeparator {
			return true
		}

		c.electedLock.Lock()
		defer c.electedLock.Unlock()

		if replica.DeletedAt > 0 {
			delete(c.elected, key)
			c.electedReplicaChanges.DeleteLabelValues(user, cluster)
			c.electedReplicaTimestamp.DeleteLabelValues(user, cluster)

			userClusters := c.replicaGroups[user]
			if userClusters != nil {
				delete(userClusters, cluster)
				if len(userClusters) == 0 {
					delete(c.replicaGroups, user)
				}
			}
			return true
		}

		elected, exists := c.elected[key]
		if replica.Replica != elected.Replica {
			c.electedReplicaChanges.WithLabelValues(user, cluster).Inc()
		}
		if !exists {
			if c.replicaGroups[user] == nil {
				c.replicaGroups[user] = map[string]struct{}{}
			}
			c.replicaGroups[user][cluster] = struct{}{}
		}
		c.elected[key] = *replica
		c.electedReplicaTimestamp.WithLabelValues(user, cluster).Set(float64(replica.ReceivedAt / 1000))
		c.electedReplicaPropagationTime.Observe(time.Since(timestamp.Time(replica.ReceivedAt)).Seconds())
		return true
	})

	wg.Wait()
	return nil
}

const (
	cleanupCyclePeriod         = 30 * time.Minute
	cleanupCycleJitterVariance = 0.2 // for 30 minutes, this is ±6 min

	// If we have received last sample for given cluster before this timeout, we will mark selected replica for deletion.
	// If selected replica is marked for deletion for this time, it is deleted completely.
	deletionTimeout = 30 * time.Minute
)

func (c *HATracker) cleanupOldReplicasLoop(ctx context.Context) {
	tick := time.NewTicker(util.DurationWithJitter(cleanupCyclePeriod, cleanupCycleJitterVariance))
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-tick.C:
			c.cleanupRuns.Inc()
			c.cleanupOldReplicas(ctx, t.Add(-deletionTimeout))
		}
	}
}

// Replicas marked for deletion before deadline will be deleted.
// Replicas with last-received timestamp before deadline will be marked for deletion.
func (c *HATracker) cleanupOldReplicas(ctx context.Context, deadline time.Time) {
	keys, err := c.client.List(ctx, "")
	if err != nil {
		level.Warn(c.logger).Log("msg", "cleanup: failed to list replica keys", "err", err)
		return
	}

	for _, key := range keys {
		if ctx.Err() != nil {
			return
		}

		val, err := c.client.Get(ctx, key)
		if err != nil {
			level.Warn(c.logger).Log("msg", "cleanup: failed to get replica value", "key", key, "err", err)
			continue
		}

		desc, ok := val.(*ReplicaDesc)
		if !ok {
			level.Error(c.logger).Log("msg", "cleanup: got invalid replica descriptor", "key", key)
			continue
		}

		if desc.DeletedAt > 0 {
			if timestamp.Time(desc.DeletedAt).After(deadline) {
				continue
			}

			// We're blindly deleting a key here. It may happen that value was updated since we have read it few lines above,
			// in which case Distributors will have updated value in memory, but Delete will remove it from KV store anyway.
			// That's not great, but should not be a problem. If KV store sends Watch notification for Delete, distributors will
			// delete it from memory, and recreate on next sample with matching replica.
			//
			// If KV store doesn't send Watch notification for Delete, distributors *with* replica in memory will keep using it,
			// while distributors *without* replica in memory will try to write it to KV store -- which will update *all*
			// watching distributors.
			err = c.client.Delete(ctx, key)
			if err != nil {
				level.Error(c.logger).Log("msg", "cleanup: failed to delete old replica", "key", key, "err", err)
				c.markingForDeletionsFailed.Inc()
			} else {
				level.Info(c.logger).Log("msg", "cleanup: deleted old replica", "key", key)
				c.deletedReplicas.Inc()
			}
			continue
		}

		// Not marked as deleted yet.
		if desc.DeletedAt == 0 && timestamp.Time(desc.ReceivedAt).Before(deadline) {
			err := c.client.CAS(ctx, key, func(in interface{}) (out interface{}, retry bool, err error) {
				d, ok := in.(*ReplicaDesc)
				if !ok || d == nil || d.DeletedAt > 0 || !timestamp.Time(desc.ReceivedAt).Before(deadline) {
					return nil, false, nil
				}

				d.DeletedAt = timestamp.FromTime(time.Now())
				return d, true, nil
			})

			if err != nil {
				c.markingForDeletionsFailed.Inc()
				level.Error(c.logger).Log("msg", "cleanup: failed to mark replica as deleted", "key", key, "err", err)
			} else {
				c.replicasMarkedForDeletion.Inc()
				level.Info(c.logger).Log("msg", "cleanup: marked replica as deleted", "key", key)
			}
		}
	}
}

// CheckReplica checks the cluster and replica against the backing KVStore and local cache in the
// tracker c to see if we should accept the incoming sample. It will return an error if the sample
// should not be accepted. Note that internally this function does checks against the stored values
// and may modify the stored data, for example to failover between replicas after a certain period of time.
// ReplicasNotMatchError is returned (from checkKVStore) if we shouldn't store this sample but are
// accepting samples from another replica for the cluster, so that there isn't a bunch of error's returned
// to customers clients.
func (c *HATracker) CheckReplica(ctx context.Context, userID, replicaGroup, replica string, now time.Time) error {
	// If HA tracking isn't enabled then accept the sample
	if !c.cfg.EnableHATracker {
		return nil
	}
	key := fmt.Sprintf("%s/%s", userID, replicaGroup)

	c.electedLock.RLock()
	entry, ok := c.elected[key]
	replicaGroups := len(c.replicaGroups[userID])
	c.electedLock.RUnlock()

	if ok && now.Sub(timestamp.Time(entry.ReceivedAt)) < c.cfg.UpdateTimeout+c.updateTimeoutJitter {
		if entry.Replica != replica {
			return ReplicasNotMatchError{replica: replica, elected: entry.Replica}
		}
		return nil
	}

	if !ok {
		if c.limits != nil {
			// If we don't know about this replicaGroup yet and we have reached the limit for number of replicaGroups, we error out now.
			if limit := c.limits.MaxHAReplicaGroups(userID); limit > 0 && replicaGroups+1 > limit {
				return TooManyReplicaGroupsError{limit: limit}
			}
		}
	}

	err := c.checkKVStore(ctx, key, replica, now)
	c.kvCASCalls.WithLabelValues(userID, replicaGroup).Inc()
	if err != nil {
		// The callback within checkKVStore will return a ReplicasNotMatchError if the sample is being deduped,
		// otherwise there may have been an actual error CAS'ing that we should log.
		if !errors.Is(err, ReplicasNotMatchError{}) {
			level.Error(c.logger).Log("msg", "rejecting sample", "err", err)
		}
	}
	return err
}

func (c *HATracker) checkKVStore(ctx context.Context, key, replica string, now time.Time) error {
	return c.client.CAS(ctx, key, func(in interface{}) (out interface{}, retry bool, err error) {
		if desc, ok := in.(*ReplicaDesc); ok && desc.DeletedAt == 0 {
			// We don't need to CAS and update the timestamp in the KV store if the timestamp we've received
			// this sample at is less than updateTimeout amount of time since the timestamp in the KV store.
			if desc.Replica == replica && now.Sub(timestamp.Time(desc.ReceivedAt)) < c.cfg.UpdateTimeout+c.updateTimeoutJitter {
				return nil, false, nil
			}

			// We shouldn't failover to accepting a new replica if the timestamp we've received this sample at
			// is less than failover timeout amount of time since the timestamp in the KV store.
			if desc.Replica != replica && now.Sub(timestamp.Time(desc.ReceivedAt)) < c.cfg.FailoverTimeout {
				return nil, false, ReplicasNotMatchError{replica: replica, elected: desc.Replica}
			}
		}

		// There was either invalid or no data for the key, so we now accept samples
		// from this replica. Invalid could mean that the timestamp in the KV store was
		// out of date based on the update and failover timeouts when compared to now.
		return &ReplicaDesc{
			Replica:    replica,
			ReceivedAt: timestamp.FromTime(now),
			DeletedAt:  0,
		}, true, nil
	})
}

func (c *HATracker) Cfg() HATrackerConfig {
	return c.cfg
}

type ReplicasNotMatchError struct {
	replica, elected string
}

func (e ReplicasNotMatchError) Error() string {
	return fmt.Sprintf("replicas did not mach, rejecting sample: replica=%s, elected=%s", e.replica, e.elected)
}

// Needed for errors.Is to work properly.
func (e ReplicasNotMatchError) Is(err error) bool {
	_, ok1 := err.(ReplicasNotMatchError)
	_, ok2 := err.(*ReplicasNotMatchError)
	return ok1 || ok2
}

// IsOperationAborted returns whether the error has been caused by an operation intentionally aborted.
func (e ReplicasNotMatchError) IsOperationAborted() bool {
	return true
}

type TooManyReplicaGroupsError struct {
	limit int
}

func (e TooManyReplicaGroupsError) Error() string {
	return fmt.Sprintf("too many HA replicaGroups (limit: %d)", e.limit)
}

// Needed for errors.Is to work properly.
func (e TooManyReplicaGroupsError) Is(err error) bool {
	_, ok1 := err.(TooManyReplicaGroupsError)
	_, ok2 := err.(*TooManyReplicaGroupsError)
	return ok1 || ok2
}

func (c *HATracker) CleanupHATrackerMetricsForUser(userID string) {
	filter := map[string]string{"user": userID}

	if err := util.DeleteMatchingLabels(c.electedReplicaChanges, filter); err != nil {
		level.Warn(c.logger).Log("msg", "failed to remove cortex_ha_tracker_elected_replica_changes_total metric for user", "user", userID, "err", err)
	}
	if err := util.DeleteMatchingLabels(c.electedReplicaTimestamp, filter); err != nil {
		level.Warn(c.logger).Log("msg", "failed to remove cortex_ha_tracker_elected_replica_timestamp_seconds metric for user", "user", userID, "err", err)
	}
	if err := util.DeleteMatchingLabels(c.kvCASCalls, filter); err != nil {
		level.Warn(c.logger).Log("msg", "failed to remove cortex_ha_tracker_kv_store_cas_total metric for user", "user", userID, "err", err)
	}
	if err := util.DeleteMatchingLabels(c.userReplicaGroupCount, filter); err != nil {
		level.Warn(c.logger).Log("msg", "failed to remove cortex_ha_tracker_user_replica_group_count metric for user", "user", userID, "err", err)
	}
}

// Returns a snapshot of the currently elected replicas.  Useful for status display
func (c *HATracker) SnapshotElectedReplicas() map[string]ReplicaDesc {
	c.electedLock.Lock()
	defer c.electedLock.Unlock()

	electedCopy := make(map[string]ReplicaDesc)
	for key, desc := range c.elected {
		electedCopy[key] = ReplicaDesc{
			Replica:    desc.Replica,
			ReceivedAt: desc.ReceivedAt,
			DeletedAt:  desc.DeletedAt,
		}
	}
	return electedCopy
}

func (t *HATracker) updateUserReplicaGroupCount() {
	t.electedLock.RLock()
	defer t.electedLock.RUnlock()

	for user, groups := range t.replicaGroups {
		t.userReplicaGroupCount.WithLabelValues(user).Set(float64(len(groups)))
	}
}
