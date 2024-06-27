Query Rules
-----------

1. All queries interact only with Object models I.E: `Locker.Lock`, `Limiter.Rule`,
   `Limiter.Override`, `Willow.Queue`, `Willow.Item`
2. Queries only interact with the `Object.Spec.DBDefinition.[Unique Name]` field
    1. This field must be a datatypes.KeyValues
    2. APIs for paricular services can enforce specific "Key + Value types" for UIs to display
3. Quries fall into one of 2 operations
    1. Node-Actionable query
    2. Lookup-Pagination query


# Node-Actionable Query

# Lookup-Pagination Query