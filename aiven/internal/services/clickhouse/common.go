// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package clickhouse

// default database used for statements that do not target a particular database
// think CREATE ROLE / GRANT / etc...
const defaultDatabase = "system"
