package userconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeSpace(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"  First line\n  wrapped.\n \nSecond   paragraph.\n": "First line wrapped.\n\nSecond paragraph.",
		"~> **Warning**\nMind the gap.":                      "~> **Warning**\nMind the gap.",
		"Options:\n- one\n- two":                             "Options:\n- one\n- two",
		"Example:\n```\na  b\n```":                           "Example:\n```\na  b\n```",
	}
	for in, want := range cases {
		assert.Equal(t, want, normalizeSpace(in), "input: %q", in)
	}
}

func TestFormatMarkdown(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		// double-quoted symbols, single or a run, literal or &quot; entity
		`use "/" as a separator`:       "use `/` as a separator",
		`defaults to [".*"]`:           "defaults to [`.*`]",
		`defaults to [&quot;.*&quot;]`: "defaults to [`.*`]",
		// a lazy run stops at the closing quote instead of spanning two of them
		`either ".*" or "-" works`: "either `.*` or `-` works",
		// the gap between adjacent quoted values is not a symbol
		`["a","b"]`:                     `["a","b"]`,
		`[&quot;a&quot;,&quot;b&quot;]`: `[&quot;a&quot;,&quot;b&quot;]`,
		`["/","."]`:                     "[`/`,`.`]",
		// quoted words stay
		`matched by "topics"`:    `matched by "topics"`,
		`a digit like "5" stays`: `a digit like "5" stays`,

		// identifiers with "_"
		"One of key or sas_token should be specified":                     "One of key or `sas_token` should be specified",
		"user_operations records calls":                                   "`user_operations` records calls",
		"set preferred_zones.":                                            "set `preferred_zones`.",
		"a table's pg_class.relfrozenxid field":                           "a table's `pg_class.relfrozenxid` field",
		"pg_partman's role":                                               "`pg_partman`'s role",
		`options are 'mysql_native_password' and "caching_sha2_password"`: "options are `mysql_native_password` and `caching_sha2_password`",
		`&quot;read_committed&quot; level`:                                "`read_committed` level",
		// lowercase dotted paths, but not abbreviations
		"set min.insync.replicas to 2, e.g. i.e. a.b": "set `min.insync.replicas` to 2, e.g. i.e. a.b",
		`matched by "topics.exclude"`:                 "matched by `topics.exclude`",
		// a setting is wrapped with its value
		"for io_method=worker.":                        "for `io_method=worker`.",
		"the GROUP_CONCAT() function, max_len (bytes)": "the `GROUP_CONCAT()` function, `max_len` (bytes)",
		"adds a diskless_az= hint":                     "adds a `diskless_az=` hint",
		"set 'ssl_mode=require' first":                 "set `ssl_mode=require` first",
		"use log_level=debug_all here":                 "use `log_level=debug_all` here",
		"use log_level=DEBUG_ALL here":                 "use `log_level=DEBUG_ALL` here",
		// colon after identifier (not a JSON key)
		"provider: aws, auth_method: credentials": "provider: aws, `auth_method`: credentials",
		// uppercase words with "_", but not acronyms
		"set AIVEN_TOKEN first":                           "set `AIVEN_TOKEN` first",
		"from PERFORMANCE_SCHEMA.EVENT_WAITS.":            "from `PERFORMANCE_SCHEMA.EVENT_WAITS`.",
		"responds with 'UNSUPPORTED_VERSION' error":       "responds with `UNSUPPORTED_VERSION` error",
		"Run server_reset_query (DISCARD ALL) in the VPC": "Run `server_reset_query` (DISCARD ALL) in the VPC",
		// links stay, link text doesn't
		"see [ACLs](https://registry.terraform.io/docs/resources/kafka_acl).": "see [ACLs](https://registry.terraform.io/docs/resources/kafka_acl).",
		"https://docs.datadoghq.com/getting_started/tagging":                  "https://docs.datadoghq.com/getting_started/tagging",
		"[aiven_kafka_acl](../kafka_acl.md)":                                  "[`aiven_kafka_acl`](../kafka_acl.md)",
		// not identifiers
		"Max_size stays":                                        "Max_size stays",
		"(Long.MAX_VALUE) and SO_RCVBUF-like":                   "(Long.MAX_VALUE) and SO_RCVBUF-like",
		"FOO_bar and aFOO_BAR":                                  "FOO_bar and aFOO_BAR",
		"a-b_c, path/to_file, my_user@host, $foo_bar, _italic_": "a-b_c, path/to_file, my_user@host, $foo_bar, _italic_",
		"already `sas_token` wrapped":                           "already `sas_token` wrapped",

		// JSON keys and string values stay quoted, JSON objects stay as is
		`{"key_name": "value_with_underscore"}`:             `{"key_name": "value_with_underscore"}`,
		`{&quot;error_code&quot;: &quot;not_powered&quot;}`: `{&quot;error_code&quot;: &quot;not_powered&quot;}`,
		`Set {"key_name":"value_name"} for key_name.`:       "Set {\"key_name\":\"value_name\"} for `key_name`.",
		`{"a_b":1,"sep":"."}`:                               `{"a_b":1,"sep":"."}`,
		`{"outer": {"key_name": "value_name"}}`:             `{"outer": {"key_name": "value_name"}}`,

		// placeholders, without the brackets, quotes and punctuation around them
		`referenced with ${<provider_name>:<key_name>} and "."`: "referenced with `${<provider_name>:<key_name>}` and `.`",
		"set ${foo_bar} here":                  "set `${foo_bar}` here",
		"a group (<group ID>) or (see <role>)": "a group (`<group ID>`) or (see `<role>`)",
		`set "<key>" or &quot;<key>&quot;`:     "set `<key>` or `<key>`",
		"must be <a>!":                         "must be `<a>`!",

		// lone * in a sentence, but not a list bullet
		"Options:\n* one * two":      "Options:\n* one `*` two",
		"Options:\n  * one\n\t* two": "Options:\n  * one\n\t* two",
		"`a` * `b`":                  "`a` `*` `b`",

		// dotted numbers
		"Some examples are: 0.8.2, 0.9.0.0, 0.10.0.":                          "Some examples are: `0.8.2`, `0.9.0.0`, `0.10.0`.",
		"Default is 0.5 (or 10.0).":                                           "Default is `0.5` (or `10.0`).",
		"Requires Kafka 4.0: on [1.2] (or 3.3); 9.0":                          "Requires Kafka `4.0`: on [`1.2`] (or `3.3`); `9.0`",
		"Valkey 9.1+, 1MB + 0.075% of RAM, 1.0-2.0":                           "Valkey 9.1+, 1MB + 0.075% of RAM, 1.0-2.0",
		"at 10.0.0.1:9092 or 1.2.3x":                                          "at 10.0.0.1:9092 or 1.2.3x",
		"If set to -1, no limit (or -0.5). Min -1.":                           "If set to `-1`, no limit (or `-0.5`). Min `-1`.",
		"UTF-8, 2024-01-01, 1-2, x-1, -5%, - 1":                               "UTF-8, 2024-01-01, 1-2, x-1, -5%, - 1",
		`a block like 10.0.0.0/24 or "1.5"`:                                   "a block like `10.0.0.0/24` or `1.5`",
		"v1.2, pkg-1.2.3, 2.7.14.dev0 and a/1.2":                              "v1.2, pkg-1.2.3, 2.7.14.dev0 and a/1.2",
		"see https://example.com/v/1.2.3 or [1.12](https://example.com/1.12)": "see https://example.com/v/1.2.3 or [`1.12`](https://example.com/1.12)",
		"for io_method=1.5 here":                                              "for `io_method=1.5` here",
	}
	for in, want := range cases {
		assert.Equal(t, want, formatMarkdown(in), "input: %q", in)
	}
}
