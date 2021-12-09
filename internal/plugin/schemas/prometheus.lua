return {
  name = "prometheus",
  fields = {
    { config = {
        type = "record",
        fields = {
          { per_consumer = { type = "boolean", default = false }, },
        },
    }, },
  },
}
