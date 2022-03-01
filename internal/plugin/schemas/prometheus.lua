local function validate_shared_dict()
  -- remove shared dictionary validation
  return true
end


return {
  name = "prometheus",
  fields = {
    { config = {
        type = "record",
        fields = {
          { per_consumer = { type = "boolean", default = false }, },
        },
        custom_validator = validate_shared_dict,
    }, },
  },
}
