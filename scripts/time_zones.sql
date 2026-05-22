TRUNCATE TABLE time_zones CASCADE;

INSERT INTO time_zones (id, name, timezone_offset, created_at, updated_at) VALUES
	('11111111-2222-3333-4444-555555550001', 'America/New_York', 'UTC-05:00', NOW(), NOW()),
	('11111111-2222-3333-4444-555555550002', 'America/Chicago', 'UTC-06:00', NOW(), NOW());
