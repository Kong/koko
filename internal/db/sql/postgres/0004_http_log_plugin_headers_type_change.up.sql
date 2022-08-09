-- In Kong Gateway v3.0, the `headers` config field on the `http-log` plugin was altered from an array of strings to a
-- single string value in the following PR: https://github.com/Kong/kong/pull/6992
--
-- As such, Koko did the same & made the following updates: https://github.com/Kong/koko/pull/329
--
-- In the aforementioned Koko PR, the change kept backwards compatibility on the DB persistence store. This DB migration
-- is to handle converting all `[]string` values on the `http-log` plugin `headers` field to a single string, so that we
-- can remove the code allowing both to be supported at the same time.
--
-- We did not choose for this DB migration to be a part of that PR, as if it were, there could be a brief impact to
-- users as this DB migration is dependent on those code changes to be deployed first (as the previous logic does not
-- know how to support reading both header value types at the same time).
--
-- The below logic looks at all headers on `http-log` plugins and handles the following conversions:
--    Input                                                  Output
-- 1. `{"headers": []}`                                      `{"headers": null}`
-- 2. `{"headers": {"header-1": "value-1"}}`                 No-op
-- 3. `{"headers": {"header-1": ["value-1"]}}`               `{"headers": {"header-1": "value-1"}}`
-- 3. `{"headers": {"header-1": ["value-1", "value-2"]}}`    `{"headers": {"header-1": "value-1, "value-2"}}`
UPDATE store SET value = jsonb_set(
    value,
    '{object, config, headers}',
    (
        CASE
            -- Header values have been provided, so we'll iterate upon all the keys & process the values.
            -- e.g.:
            --   DP <= 2.8: `{"headers": {"header-1": ["value-1", "value-2"]}}`
            --   DP >= 3.0: `{"headers": {"header-1": "value-1, value-2"}}`
            WHEN jsonb_typeof(value -> 'object' -> 'config' -> 'headers') = 'object' THEN (
                SELECT jsonb_object_agg(key, (
                    CASE WHEN jsonb_typeof(value) = 'array' THEN (
                        -- The header values consist of a JSON array, so we'll re-write the values
                        -- from a list to a single string value. This matches the behavior of
                        -- the 2.8 -> 3.0 migration happening on the data plane.
                        --
                        -- Read more: https://github.com/Kong/kong/pull/9162
                        SELECT to_jsonb(string_agg(vals, ', '))
                        FROM jsonb_array_elements_text(value) as vals
                    )
                    ELSE
                        -- The header value is already a string, so nothing to do here.
                        value
                    END
                ))
                FROM jsonb_each(value -> 'object' -> 'config' -> 'headers')
            )

            -- When an empty object (`{}`) has been provided for the `headers` during creation, Koko is storing that
            -- as an empty array (`"headers": []`). As such, we'll need to convert this to a literal JSON null value,
            -- as arrays are no longer supported with the backwards incompatible schema changes that were done.
            WHEN jsonb_typeof(value -> 'object' -> 'config' -> 'headers') = 'array' THEN 'null'::jsonb
        END
    )
)
WHERE
    key LIKE 'c/%/o/plugin/%' AND
    value -> 'object' ->> 'name' = 'http-log' AND
    jsonb_typeof(value -> 'object' -> 'config' -> 'headers') != 'null';
