// SPDX-License-Identifier: GPL-3.0-or-later

package postgres

import (
	"fmt"
	"strings"

	"github.com/netdata/go.d.plugin/agent/module"
)

const (
	prioConnectionsUtilization = module.Priority + iota
	prioConnectionsUsage
	prioCheckpoints
	prioCheckpointTime
	prioBGWriterBuffersAllocated
	prioBGWriterBuffersWritten
	prioBGWriterMaxWrittenClean
	prioBGWriterBackedFsync
	prioWALWrites
	prioWALFiles
	prioWALArchive
	prioAutovacuumWorkers
	prioAutovacuumPercentTowards
	prioTXIDWraparoundPercentTowards
	prioTXIDWraparoundOldestTXID
	prioCatalogRelationCount
	prioCatalogRelationSize
	prioUptime
	prioDBTransactions
	prioDBConnectionsUtilization
	prioDBConnections
	prioDBBufferCache
	prioDBReadOperations
	prioDBWriteOperations
	prioDBConflicts
	prioDBConflictsStat
	prioDBDeadlocks
	prioDBLocksHeld
	prioDBLocksAwaited
	prioDBTempFiles
	prioDBTempFilesData
	prioDBSize
)

var baseCharts = module.Charts{
	serverConnectionsUtilizationChart.Copy(),
	serverConnectionsUsageChart.Copy(),
	checkpointsChart.Copy(),
	checkpointWriteChart.Copy(),
	bgWriterBuffersWrittenChart.Copy(),
	bgWriterBuffersAllocChart.Copy(),
	bgWriterMaxWrittenCleanChart.Copy(),
	bgWriterBuffersBackendFsyncChart.Copy(),
	walWritesChart.Copy(),
	walFilesChart.Copy(),
	walArchiveFilesChart.Copy(),
	autovacuumWorkersChart.Copy(),
	percentTowardsEmergencyAutovacuumChart.Copy(),
	percentTowardTXIDWraparoundChart.Copy(),
	oldestTXIDChart.Copy(),

	catalogRelationCountChart.Copy(),
	catalogRelationSizeChart.Copy(),
	serverUptimeChart.Copy(),
}

var (
	serverConnectionsUtilizationChart = module.Chart{
		ID:       "connections_utilization",
		Title:    "Connections utilization",
		Units:    "percentage",
		Fam:      "connections",
		Ctx:      "postgres.connections_utilization",
		Priority: prioConnectionsUtilization,
		Dims: module.Dims{
			{ID: "server_connections_utilization", Name: "used"},
		},
	}
	serverConnectionsUsageChart = module.Chart{
		ID:       "connections_usage",
		Title:    "Connections usage",
		Units:    "connections",
		Fam:      "connections",
		Ctx:      "postgres.connections_usage",
		Priority: prioConnectionsUsage,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "server_connections_available", Name: "available"},
			{ID: "server_connections_used", Name: "used"},
		},
	}

	checkpointsChart = module.Chart{
		ID:       "checkpoints",
		Title:    "Checkpoints",
		Units:    "checkpoints/s",
		Fam:      "checkpointer",
		Ctx:      "postgres.checkpoints",
		Priority: prioCheckpoints,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "checkpoints_timed", Name: "scheduled", Algo: module.Incremental},
			{ID: "checkpoints_req", Name: "requested", Algo: module.Incremental},
		},
	}
	// TODO: should be seconds, also it is units/s when using incremental...
	checkpointWriteChart = module.Chart{
		ID:       "checkpoint_time",
		Title:    "Checkpoint time",
		Units:    "milliseconds",
		Fam:      "checkpointer",
		Ctx:      "postgres.checkpoint_time",
		Priority: prioCheckpointTime,
		Dims: module.Dims{
			{ID: "checkpoint_write_time", Name: "write", Algo: module.Incremental},
			{ID: "checkpoint_sync_time", Name: "sync", Algo: module.Incremental},
		},
	}

	bgWriterBuffersAllocChart = module.Chart{
		ID:       "bgwriter_buffers_alloc",
		Title:    "Background writer buffers allocated",
		Units:    "B/s",
		Fam:      "background writer",
		Ctx:      "postgres.bgwriter_buffers_alloc",
		Priority: prioBGWriterBuffersAllocated,
		Dims: module.Dims{
			{ID: "buffers_alloc", Name: "allocated", Algo: module.Incremental},
		},
	}
	bgWriterBuffersWrittenChart = module.Chart{
		ID:       "bgwriter_buffers_written",
		Title:    "Background writer buffers written",
		Units:    "B/s",
		Fam:      "background writer",
		Ctx:      "postgres.bgwriter_buffers_written",
		Priority: prioBGWriterBuffersWritten,
		Type:     module.Area,
		Dims: module.Dims{
			{ID: "buffers_checkpoint", Name: "checkpoint", Algo: module.Incremental},
			{ID: "buffers_backend", Name: "backend", Algo: module.Incremental},
			{ID: "buffers_clean", Name: "clean", Algo: module.Incremental},
		},
	}
	bgWriterMaxWrittenCleanChart = module.Chart{
		ID:       "bgwriter_maxwritten_clean",
		Title:    "Background writer cleaning scan stops",
		Units:    "events/s",
		Fam:      "background writer",
		Ctx:      "postgres.bgwriter_maxwritten_clean",
		Priority: prioBGWriterMaxWrittenClean,
		Dims: module.Dims{
			{ID: "maxwritten_clean", Name: "maxwritten", Algo: module.Incremental},
		},
	}
	bgWriterBuffersBackendFsyncChart = module.Chart{
		ID:       "bgwriter_buffers_backend_fsync",
		Title:    "Backend fsync",
		Units:    "operations/s",
		Fam:      "background writer",
		Ctx:      "postgres.bgwriter_buffers_backend_fsync",
		Priority: prioBGWriterBackedFsync,
		Dims: module.Dims{
			{ID: "buffers_backend_fsync", Name: "fsync", Algo: module.Incremental},
		},
	}

	walWritesChart = module.Chart{
		ID:       "wal_writes",
		Title:    "Write-Ahead Log",
		Units:    "B/s",
		Fam:      "wal",
		Ctx:      "postgres.wal_writes",
		Priority: prioWALWrites,
		Dims: module.Dims{
			{ID: "wal_writes", Name: "writes", Algo: module.Incremental},
		},
	}
	walFilesChart = module.Chart{
		ID:       "wal_files",
		Title:    "Write-Ahead Log files",
		Units:    "files",
		Fam:      "wal",
		Ctx:      "postgres.wal_files",
		Priority: prioWALFiles,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "wal_written_files", Name: "written"},
			{ID: "wal_recycled_files", Name: "recycled"},
		},
	}

	walArchiveFilesChart = module.Chart{
		ID:       "wal_archive_files",
		Title:    "Write-Ahead Log archive files",
		Units:    "files/s",
		Fam:      "wal archive",
		Ctx:      "postgres.wal_archive_files",
		Priority: prioWALArchive,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "wal_archive_files_ready_count", Name: "ready", Algo: module.Incremental},
			{ID: "wal_archive_files_done_count", Name: "done", Algo: module.Incremental},
		},
	}

	autovacuumWorkersChart = module.Chart{
		ID:       "autovacuum_workers",
		Title:    "Autovacuum workers",
		Units:    "workers",
		Fam:      "autovacuum",
		Ctx:      "postgres.autovacuum_workers",
		Priority: prioAutovacuumWorkers,
		Dims: module.Dims{
			{ID: "autovacuum_analyze", Name: "analyze"},
			{ID: "autovacuum_vacuum_analyze", Name: "vacuum_analyze"},
			{ID: "autovacuum_vacuum", Name: "vacuum"},
			{ID: "autovacuum_vacuum_freeze", Name: "vacuum_freeze"},
			{ID: "autovacuum_brin_summarize", Name: "brin_summarize"},
		},
	}
	percentTowardsEmergencyAutovacuumChart = module.Chart{
		ID:       "percent_towards_emergency_autovacuum",
		Title:    "Percent towards emergency autovacuum",
		Units:    "percentage",
		Fam:      "autovacuum",
		Ctx:      "postgres.percent_towards_emergency_autovacuum",
		Priority: prioAutovacuumPercentTowards,
		Dims: module.Dims{
			{ID: "percent_towards_emergency_autovacuum", Name: "emergency_autovacuum"},
		},
	}

	percentTowardTXIDWraparoundChart = module.Chart{
		ID:       "percent_towards_txid_wraparound",
		Title:    "Percent towards transaction ID wraparound",
		Units:    "percentage",
		Fam:      "txid wraparound",
		Ctx:      "postgres.percent_towards_txid_wraparound",
		Priority: prioTXIDWraparoundPercentTowards,
		Dims: module.Dims{
			{ID: "percent_towards_wraparound", Name: "txid_wraparound"},
		},
	}
	oldestTXIDChart = module.Chart{
		ID:       "oldest_transaction_xid",
		Title:    "Oldest transaction XID",
		Units:    "xid",
		Fam:      "txid wraparound",
		Ctx:      "postgres.oldest_transaction_xid",
		Priority: prioTXIDWraparoundOldestTXID,
		Dims: module.Dims{
			{ID: "oldest_current_xid", Name: "xid"},
		},
	}

	catalogRelationCountChart = module.Chart{
		ID:       "catalog_relation_count",
		Title:    "Relation count",
		Units:    "relations",
		Fam:      "catalog",
		Ctx:      "postgres.catalog_relation_count",
		Priority: prioCatalogRelationCount,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "catalog_relkind_r_count", Name: "ordinary_table"},
			{ID: "catalog_relkind_i_count", Name: "index"},
			{ID: "catalog_relkind_S_count", Name: "sequence"},
			{ID: "catalog_relkind_t_count", Name: "toast_table"},
			{ID: "catalog_relkind_v_count", Name: "view"},
			{ID: "catalog_relkind_m_count", Name: "materialized_view"},
			{ID: "catalog_relkind_c_count", Name: "composite_type"},
			{ID: "catalog_relkind_f_count", Name: "foreign_table"},
			{ID: "catalog_relkind_p_count", Name: "partitioned_table"},
			{ID: "catalog_relkind_I_count", Name: "partitioned_index"},
		},
	}
	catalogRelationSizeChart = module.Chart{
		ID:       "catalog_relation_size",
		Title:    "Relation size",
		Units:    "B",
		Fam:      "catalog",
		Ctx:      "postgres.catalog_relation_size",
		Priority: prioCatalogRelationSize,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "catalog_relkind_r_size", Name: "ordinary_table"},
			{ID: "catalog_relkind_i_size", Name: "index"},
			{ID: "catalog_relkind_S_size", Name: "sequence"},
			{ID: "catalog_relkind_t_size", Name: "toast_table"},
			{ID: "catalog_relkind_v_size", Name: "view"},
			{ID: "catalog_relkind_m_size", Name: "materialized_view"},
			{ID: "catalog_relkind_c_size", Name: "composite_type"},
			{ID: "catalog_relkind_f_size", Name: "foreign_table"},
			{ID: "catalog_relkind_p_size", Name: "partitioned_table"},
			{ID: "catalog_relkind_I_size", Name: "partitioned_index"},
		},
	}

	serverUptimeChart = module.Chart{
		ID:       "server_uptime",
		Title:    "Uptime",
		Units:    "seconds",
		Fam:      "uptime",
		Ctx:      "postgres.uptime",
		Priority: prioUptime,
		Dims: module.Dims{
			{ID: "server_uptime", Name: "uptime"},
		},
	}
)

var (
	dbChartsTmpl = module.Charts{
		dbTransactionsChartTmpl.Copy(),
		dbConnectionsUtilizationChartTmpl.Copy(),
		dbConnectionsChartTmpl.Copy(),
		dbBufferCacheChartTmpl.Copy(),
		dbReadOpsChartTmpl.Copy(),
		dbWriteOpsChartTmpl.Copy(),
		dbConflictsChartTmpl.Copy(),
		dbConflictsStatChartTmpl.Copy(),
		dbDeadlocksChartTmpl.Copy(),
		dbLocksHeldChartTmpl.Copy(),
		dbLocksAwaitedChartTmpl.Copy(),
		dbTempFilesChartTmpl.Copy(),
		dbTempFilesDataChartTmpl.Copy(),
		dbSizeChartTmpl.Copy(),
	}
	dbTransactionsChartTmpl = module.Chart{
		ID:       "db_%s_transactions",
		Title:    "Database transactions",
		Units:    "transactions/s",
		Fam:      "db transactions",
		Ctx:      "postgres.db_transactions",
		Priority: prioDBTransactions,
		Dims: module.Dims{
			{ID: "db_%s_xact_commit", Name: "committed", Algo: module.Incremental},
			{ID: "db_%s_xact_rollback", Name: "rollback", Algo: module.Incremental},
		},
	}
	dbConnectionsUtilizationChartTmpl = module.Chart{
		ID:       "db_%s_connections_utilization",
		Title:    "Database connections utilization withing limits",
		Units:    "percentage",
		Fam:      "db connections",
		Ctx:      "postgres.db_connections_utilization",
		Priority: prioDBConnectionsUtilization,
		Dims: module.Dims{
			{ID: "db_%s_numbackends_utilization", Name: "used"},
		},
	}
	dbConnectionsChartTmpl = module.Chart{
		ID:       "db_%s_connections",
		Title:    "Database connections",
		Units:    "connections",
		Fam:      "db connections",
		Ctx:      "postgres.db_connections",
		Priority: prioDBConnections,
		Dims: module.Dims{
			{ID: "db_%s_numbackends", Name: "connections"},
		},
	}
	dbBufferCacheChartTmpl = module.Chart{
		ID:       "db_%s_buffer_cache",
		Title:    "Database buffer cache",
		Units:    "blocks/s",
		Fam:      "db buffer cache",
		Ctx:      "postgres.db_buffer_cache",
		Priority: prioDBBufferCache,
		Type:     module.Area,
		Dims: module.Dims{
			{ID: "db_%s_blks_hit", Name: "hit", Algo: module.Incremental},
			{ID: "db_%s_blks_read", Name: "miss", Algo: module.Incremental},
		},
	}
	dbReadOpsChartTmpl = module.Chart{
		ID:       "db_%s_read_operations",
		Title:    "Database read operations",
		Units:    "rows/s",
		Fam:      "db operations",
		Ctx:      "postgres.db_read_operations",
		Priority: prioDBReadOperations,
		Dims: module.Dims{
			{ID: "db_%s_tup_returned", Name: "returned", Algo: module.Incremental},
			{ID: "db_%s_tup_fetched", Name: "fetched", Algo: module.Incremental},
		},
	}
	dbWriteOpsChartTmpl = module.Chart{
		ID:       "db_%s_write_operations",
		Title:    "Database write operations",
		Units:    "rows/s",
		Fam:      "db operations",
		Ctx:      "postgres.db_write_operations",
		Priority: prioDBWriteOperations,
		Dims: module.Dims{
			{ID: "db_%s_tup_inserted", Name: "inserted", Algo: module.Incremental},
			{ID: "db_%s_tup_deleted", Name: "deleted", Algo: module.Incremental},
			{ID: "db_%s_tup_updated", Name: "updated", Algo: module.Incremental},
		},
	}
	dbConflictsChartTmpl = module.Chart{
		ID:       "db_%s_conflicts",
		Title:    "Database canceled queries",
		Units:    "queries/s",
		Fam:      "db operations",
		Ctx:      "postgres.db_conflicts",
		Priority: prioDBConflicts,
		Dims: module.Dims{
			{ID: "db_%s_conflicts", Name: "conflicts", Algo: module.Incremental},
		},
	}
	dbConflictsStatChartTmpl = module.Chart{
		ID:       "db_%s_conflicts_stat",
		Title:    "Database canceled queries by reason",
		Units:    "queries/s",
		Fam:      "db operations",
		Ctx:      "postgres.db_conflicts_stat",
		Priority: prioDBConflictsStat,
		Dims: module.Dims{
			{ID: "db_%s_confl_tablespace", Name: "tablespace", Algo: module.Incremental},
			{ID: "db_%s_confl_lock", Name: "lock", Algo: module.Incremental},
			{ID: "db_%s_confl_snapshot", Name: "snapshot", Algo: module.Incremental},
			{ID: "db_%s_confl_bufferpin", Name: "bufferpin", Algo: module.Incremental},
			{ID: "db_%s_confl_deadlock", Name: "deadlock", Algo: module.Incremental},
		},
	}
	dbDeadlocksChartTmpl = module.Chart{
		ID:       "db_%s_deadlocks",
		Title:    "Database deadlocks",
		Units:    "deadlocks/s",
		Fam:      "db deadlocks",
		Ctx:      "postgres.db_deadlocks",
		Priority: prioDBDeadlocks,
		Dims: module.Dims{
			{ID: "db_%s_deadlocks", Name: "deadlocks", Algo: module.Incremental},
		},
	}
	dbLocksHeldChartTmpl = module.Chart{
		ID:       "db_%s_locks_held",
		Title:    "Database locks held",
		Units:    "locks",
		Fam:      "db locks",
		Ctx:      "postgres.db_locks_held",
		Priority: prioDBLocksHeld,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "db_%s_lock_mode_AccessShareLock_held", Name: "access_share"},
			{ID: "db_%s_lock_mode_RowShareLock_held", Name: "row_share"},
			{ID: "db_%s_lock_mode_RowExclusiveLock_held", Name: "row_exclusive"},
			{ID: "db_%s_lock_mode_ShareUpdateExclusiveLock_held", Name: "share_update"},
			{ID: "db_%s_lock_mode_ShareLock_held", Name: "share"},
			{ID: "db_%s_lock_mode_ShareRowExclusiveLock_held", Name: "share_row_exclusive"},
			{ID: "db_%s_lock_mode_ExclusiveLock_held", Name: "exclusive"},
			{ID: "db_%s_lock_mode_AccessExclusiveLock_held", Name: "access_exclusive"},
		},
	}
	dbLocksAwaitedChartTmpl = module.Chart{
		ID:       "db_%s_locks_awaited",
		Title:    "Database locks awaited",
		Units:    "locks",
		Fam:      "db locks",
		Ctx:      "postgres.db_locks_awaited",
		Priority: prioDBLocksAwaited,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "db_%s_lock_mode_AccessShareLock_awaited", Name: "access_share"},
			{ID: "db_%s_lock_mode_RowShareLock_awaited", Name: "row_share"},
			{ID: "db_%s_lock_mode_RowExclusiveLock_awaited", Name: "row_exclusive"},
			{ID: "db_%s_lock_mode_ShareUpdateExclusiveLock_awaited", Name: "share_update"},
			{ID: "db_%s_lock_mode_ShareLock_awaited", Name: "share"},
			{ID: "db_%s_lock_mode_ShareRowExclusiveLock_awaited", Name: "share_row_exclusive"},
			{ID: "db_%s_lock_mode_ExclusiveLock_awaited", Name: "exclusive"},
			{ID: "db_%s_lock_mode_AccessExclusiveLock_awaited", Name: "access_exclusive"},
		},
	}
	dbTempFilesChartTmpl = module.Chart{
		ID:       "db_%s_temp_files",
		Title:    "Database temporary files written to disk",
		Units:    "files/s",
		Fam:      "db temp files",
		Ctx:      "postgres.db_temp_files",
		Priority: prioDBTempFiles,
		Dims: module.Dims{
			{ID: "db_%s_temp_files", Name: "written", Algo: module.Incremental},
		},
	}
	dbTempFilesDataChartTmpl = module.Chart{
		ID:       "db_%s_temp_files_data",
		Title:    "Database temporary files data written to disk",
		Units:    "B/s",
		Fam:      "db temp files",
		Ctx:      "postgres.db_temp_files_data",
		Priority: prioDBTempFilesData,
		Dims: module.Dims{
			{ID: "db_%s_temp_bytes", Name: "written", Algo: module.Incremental},
		},
	}
	dbSizeChartTmpl = module.Chart{
		ID:       "db_%s_size",
		Title:    "Database size",
		Units:    "B",
		Fam:      "db size",
		Ctx:      "postgres.db_size",
		Priority: prioDBSize,
		Dims: module.Dims{
			{ID: "db_%s_size", Name: "size"},
		},
	}
)

func newDatabaseCharts(dbname string) *module.Charts {
	charts := dbChartsTmpl.Copy()
	for _, c := range *charts {
		c.ID = fmt.Sprintf(c.ID, dbname)
		c.Labels = []module.Label{
			{Key: "database", Value: dbname},
		}
		for _, d := range c.Dims {
			d.ID = fmt.Sprintf(d.ID, dbname)
		}
	}
	return charts
}

func (p *Postgres) addNewDatabaseCharts(dbname string) {
	charts := newDatabaseCharts(dbname)
	if err := p.Charts().Add(*charts...); err != nil {
		p.Warning(err)
	}
}

func (p *Postgres) removeDatabaseCharts(dbname string) {
	prefix := fmt.Sprintf("db_%s_", dbname)
	for _, c := range *p.Charts() {
		if strings.HasPrefix(c.ID, prefix) {
			c.MarkRemove()
			c.MarkNotCreated()
		}
	}
}