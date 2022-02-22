create extension if not exists pg_trgm; 
create table if not exists store(key text PRIMARY KEY, value jsonb);
create index if not exists store_tags_idx ON store USING gin ((value->'tags') jsonb_path_ops);