create extension if not exists pg_trgm;
create index if not exists store_value_tags_idx on store USING gin ((value->'tags') jsonb_path_ops);
