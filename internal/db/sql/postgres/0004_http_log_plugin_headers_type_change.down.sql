-- See the relevant up migration (`0004_http_log_plugin_headers_type_change.up.sql`) for more information.
--
-- This migration converts all `http-log` header values back to an array of strings.
UPDATE store SET value = jsonb_set(
    value,
    '{object, config, headers}',
    (
        SELECT jsonb_object_agg(key, (
            -- Convert the header value from a string to a string within an array. `#>> '{}'` is a shortcut
            -- to convert a JSONB string value to a textual string, without the surrounding double quotes.
            CASE WHEN jsonb_typeof(value) = 'string' THEN array_to_json(ARRAY[value #>> '{}'])::jsonb
            ELSE
                -- The header value is already an array, so nothing to do here.
                value
            END
        ))
        FROM jsonb_each(value -> 'object' -> 'config' -> 'headers')
    )
)
WHERE
    key LIKE 'c/%/o/plugin/%' AND
    value -> 'object' ->> 'name' = 'http-log' AND
    jsonb_typeof(value -> 'object' -> 'config' -> 'headers') != 'null';
