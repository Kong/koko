local typedefs = require "kong.db.schema.typedefs"

local function validate_shared_dict()
  -- remove shared dictionary validation
  return true
end


return {
  name = "prometheus",
  fields = {
    { protocols = typedefs.protocols },
    { config = {
        type = "record",
        fields = {
          { per_consumer = { type = "boolean", default = false }, },
          { status_code_metrics = { type = "boolean", default = false }, },
          { latency_metrics = { type = "boolean", default = false }, },
          { bandwidth_metrics = { type = "boolean", default = false }, },
          { upstream_health_metrics = { type = "boolean", default = false }, },
        },
        custom_validator = validate_shared_dict,
    }, },
  },
}
