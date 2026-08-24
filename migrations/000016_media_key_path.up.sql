-- file_path kini menyimpan object key (bukan URL penuh).
-- URL lama: http(s)://<host>/<bucket>/<key>  ->  <key>
UPDATE media
SET file_path = regexp_replace(file_path, '^https?://[^/]+/[^/]+/', '')
WHERE file_path LIKE 'http%://%';
