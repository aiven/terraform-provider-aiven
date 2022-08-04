package clickhouse

// default database used for statements that do not target a particular database
// think CREATE ROLE / GRANT / etc...
const defaultDatabase = "system"

const betaDeprecationMessage = "This Resource is not yet generally available and may be subject to breaking changes " +
	"without warning"
