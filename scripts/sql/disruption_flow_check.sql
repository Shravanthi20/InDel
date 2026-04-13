-- Disruption flow verification (backend DB -> worker frontend feed)
-- Run against indel_demo or your target DB.

-- 1) Latest disruptions (source table used by platform /disruptions API)
SELECT id,
       zone_id,
       type,
       severity,
       status,
       confidence,
       signal_timestamp,
       confirmed_at,
       created_at
FROM disruptions
ORDER BY created_at DESC
LIMIT 20;

-- 2) Latest disruption notifications (source table used by worker /notifications API)
SELECT id,
       worker_id,
       type,
       message,
       created_at
FROM notifications
WHERE type = 'disruption_alert'
ORDER BY created_at DESC
LIMIT 50;

-- 3) Zone-to-worker mapping sanity
SELECT wp.zone_id,
       COUNT(DISTINCT wp.worker_id) AS workers_in_zone
FROM worker_profiles wp
GROUP BY wp.zone_id
ORDER BY workers_in_zone DESC, wp.zone_id;

-- 4) For latest disruptions, count matching worker notifications in same zone window
WITH latest_disruptions AS (
    SELECT id, zone_id, created_at
    FROM disruptions
    ORDER BY created_at DESC
    LIMIT 20
)
SELECT ld.id AS disruption_id,
       ld.zone_id,
       ld.created_at AS disruption_created_at,
       COUNT(n.id) AS disruption_alert_notifications
FROM latest_disruptions ld
LEFT JOIN worker_profiles wp ON wp.zone_id = ld.zone_id
LEFT JOIN notifications n
    ON n.worker_id = wp.worker_id
   AND n.type = 'disruption_alert'
   AND n.created_at >= ld.created_at - INTERVAL '2 minutes'
GROUP BY ld.id, ld.zone_id, ld.created_at
ORDER BY ld.created_at DESC;

-- 5) Quick health summary for API-facing flow
SELECT
  (SELECT COUNT(*) FROM disruptions WHERE created_at > NOW() - INTERVAL '1 hour') AS disruptions_last_hour,
  (SELECT COUNT(*) FROM notifications WHERE type = 'disruption_alert' AND created_at > NOW() - INTERVAL '1 hour') AS disruption_alerts_last_hour,
  (SELECT COUNT(*) FROM worker_profiles) AS total_workers,
  (SELECT COUNT(*) FROM zones) AS total_zones;
