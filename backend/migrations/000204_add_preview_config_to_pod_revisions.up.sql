ALTER TABLE pod_config_revisions
  ADD COLUMN preview_port INTEGER NOT NULL DEFAULT 0;

ALTER TABLE pod_config_revisions
  ADD COLUMN preview_path VARCHAR(255) NOT NULL DEFAULT '/';

-- Migration 000183 accepted unrestricted legacy preview metadata. Invalid
-- ports are disabled and unsafe paths reset to root before constraints land.
WITH normalized AS (
  SELECT
    id,
    CASE
      WHEN preview_port <> 0
        AND (preview_port < 1024 OR preview_port > 65535)
      THEN 0
      ELSE preview_port
    END AS preview_port,
    CASE
      WHEN preview_path = ''
        OR left(preview_path, 1) <> '/'
        OR position('?' IN preview_path) > 0
        OR position('#' IN preview_path) > 0
        OR preview_path ~ '(^|/)\.\.(/|$)'
        OR preview_path ~* '%2e|%2f'
        OR position(
          '%' IN regexp_replace(preview_path, '%[0-9A-Fa-f]{2}', '', 'g')
        ) > 0
      THEN '/'
      ELSE preview_path
    END AS preview_path
  FROM pods
),
canonical AS (
  SELECT
    id,
    preview_port,
    COALESCE(
      NULLIF(
        regexp_replace(
          regexp_replace(
            regexp_replace(preview_path, '/(\./)+', '/', 'g'),
            '/\.$',
            ''
          ),
          '/+',
          '/',
          'g'
        ),
        ''
      ),
      '/'
    ) AS preview_path
  FROM normalized
)
UPDATE pods
SET
  preview_port = canonical.preview_port,
  preview_path = CASE
    WHEN canonical.preview_path = '/' THEN '/'
    ELSE regexp_replace(canonical.preview_path, '/+$', '')
  END
FROM canonical
WHERE pods.id = canonical.id;

-- Revision-level preview metadata did not exist before this migration, so the
-- current canonical Pod value is the only deterministic snapshot available.
UPDATE pod_config_revisions
SET
  preview_port = pods.preview_port,
  preview_path = pods.preview_path
FROM pods
WHERE pod_config_revisions.pod_id = pods.id;

-- These checks guard the canonical storage form expressible in PostgreSQL.
-- NormalizePreviewConfig remains authoritative for decoded URL semantics.
ALTER TABLE pods
  ADD CONSTRAINT pods_preview_port_check
  CHECK (preview_port = 0 OR preview_port BETWEEN 1024 AND 65535);

ALTER TABLE pods
  ADD CONSTRAINT pods_preview_path_check
  CHECK (
    preview_path <> ''
    AND left(preview_path, 1) = '/'
    AND position('?' IN preview_path) = 0
    AND position('#' IN preview_path) = 0
    AND position('//' IN preview_path) = 0
    AND (preview_path = '/' OR right(preview_path, 1) <> '/')
    AND preview_path !~ '(^|/)\.{1,2}(/|$)'
    AND preview_path !~* '%2e|%2f'
    AND position(
      '%' IN regexp_replace(preview_path, '%[0-9A-Fa-f]{2}', '', 'g')
    ) = 0
  );

ALTER TABLE pod_config_revisions
  ADD CONSTRAINT pod_config_revisions_preview_port_check
  CHECK (preview_port = 0 OR preview_port BETWEEN 1024 AND 65535);

ALTER TABLE pod_config_revisions
  ADD CONSTRAINT pod_config_revisions_preview_path_check
  CHECK (
    preview_path <> ''
    AND left(preview_path, 1) = '/'
    AND position('?' IN preview_path) = 0
    AND position('#' IN preview_path) = 0
    AND position('//' IN preview_path) = 0
    AND (preview_path = '/' OR right(preview_path, 1) <> '/')
    AND preview_path !~ '(^|/)\.{1,2}(/|$)'
    AND preview_path !~* '%2e|%2f'
    AND position(
      '%' IN regexp_replace(preview_path, '%[0-9A-Fa-f]{2}', '', 'g')
    ) = 0
  );
