package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func TestParseEnum_SDKPrefix(t *testing.T) {
	desc := userconfig.EnumPrefix + "`3.9`, `4.2`, `4.3`, and newer. Kafka MirrorMaker major version."
	values, rest := parseEnum(desc)
	assert.Equal(t, []string{"3.9", "4.2", "4.3"}, values)
	assert.Equal(t, "Kafka MirrorMaker major version.", rest)
}

func TestParseEnum_andNewerNotAValue(t *testing.T) {
	desc := userconfig.EnumPrefix + "`8.1`, `9.0`, and newer. Valkey major version."
	values, _ := parseEnum(desc)
	assert.Equal(t, []string{"8.1", "9.0"}, values)
}

func TestParseEnum_possibleValues(t *testing.T) {
	desc := userconfig.Desc("Foo.").PossibleValuesString("bar", "baz", "qux").Build()
	values, rest := parseEnum(desc)
	assert.Equal(t, []string{"bar", "baz", "qux"}, values)
	assert.Equal(t, "Foo.", rest)
}

func TestParseEnum_singleValue(t *testing.T) {
	desc := userconfig.Desc("Intro.").PossibleValuesString("solo").Build()
	values, rest := parseEnum(desc)
	assert.Equal(t, []string{"solo"}, values)
	assert.Equal(t, "Intro.", rest)
}

func TestParseEnum_defaultAfterList(t *testing.T) {
	desc := userconfig.EnumPrefix + "`io_uring`, `sync`, `worker`. EXPERIMENTAL: foo. Default: `worker`."
	values, rest := parseEnum(desc)
	assert.Equal(t, []string{"io_uring", "sync", "worker"}, values)
	assert.Contains(t, rest, "Default: `worker`")
	assert.NotContains(t, rest, userconfig.EnumPrefix)
}

func TestParseEnum_skipsProsePrefix(t *testing.T) {
	desc := "Document The possible values are in the guide. The possible values are `a` and `b`."
	values, rest := parseEnum(desc)
	assert.Equal(t, []string{"a", "b"}, values)
	assert.Equal(t, "Document The possible values are in the guide.", rest)
}

func TestParseEnum_proseEnumWord(t *testing.T) {
	desc := "An enumeration of project states."
	values, rest := parseEnum(desc)
	assert.Nil(t, values)
	assert.Equal(t, desc, rest)
}

func TestParseEnum_none(t *testing.T) {
	desc := "Plain description without values."
	assert.Nil(t, parseEnumValues(desc))
}

func parseEnumValues(desc string) []string {
	v, _ := parseEnum(desc)
	return v
}

func TestParseEnum_roundTrip(t *testing.T) {
	built := userconfig.Desc("An enumeration.").PossibleValuesString("active", "deactivated", "deleted").
		DefaultValue("active").Build()
	values, rest := parseEnum(built)
	assert.Equal(t, []string{"active", "deactivated", "deleted"}, values)
	assert.Equal(t, "An enumeration. The default value is `active`.", rest)
}

func TestAddDetail_fromBuilder(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{
			name: "possible values",
			desc: userconfig.Desc("An enumeration.").PossibleValuesString("a", "b", "c").Build(),
			want: "An enumeration. The possible values are `a`, `b` and `c`.",
		},
		{
			name: "sdk enum sentence",
			desc: fmt.Sprintf("%s%s.", userconfig.EnumPrefix, "`3.9`, `4.2`, `4.3`, and newer") +
				" Kafka MirrorMaker major version.",
			want: "Kafka MirrorMaker major version. The possible values are `3.9`, `4.2` and `4.3`.",
		},
		{
			name: "plain",
			desc: "Plain field description",
			want: "Plain field description.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, addDetail(tt.desc))
		})
	}
}

func TestBrief_eG(t *testing.T) {
	desc := "Use e.g. `foo` for testing. " + strings.Repeat("Padding ", 30)
	got := brief(desc)
	assert.Contains(t, got, "e.g.")
	assert.NotContains(t, got, "Padding")
}

func TestBrief_multiSentence(t *testing.T) {
	desc := "First sentence here. " + strings.Repeat("word ", 80)
	got := brief(desc)
	assert.Equal(t, "First sentence here", got)
}

func TestDescBriefStability(t *testing.T) {
	samples := []string{
		"Enum: `off`, `rdb`. Controls whether Valkey writes RDB dumps to disk. With `rdb`, RDB dumps are written for backups on the backup schedule and, if `frequent_snapshots` is enabled, every 10 minutes so the service can recover recent data after a restart.",
		"Enum: `8.1`, `9.0`, `9.1`, and newer. Valkey major version.",
		"Enum: `bloom`, `json`, `search`.",
		"Enum: `allkeys-lfu`, `allkeys-lru`, `allkeys-random`, `noeviction`, `volatile-lfu`, `volatile-lru`, `volatile-random`, `volatile-ttl`. Valkey maxmemory-policy. Default: `noeviction`.",
		"Enum: `dump`, `replication`. The migration method to be used (currently supported only by Redis, MySQL and PostgreSQL service types).",
	}
	for i, sample := range samples {
		t.Run(fmt.Sprintf("sample%d", i), func(t *testing.T) {
			prose := brief(sample)
			if prose == "" {
				t.Skip("no prose after enum removal")
			}
			rebuilt := userconfig.Desc(prose).Build()
			assert.Equal(t, words(prose), words(strings.TrimSuffix(rebuilt, ".")), "rebuilt=%q", rebuilt)
		})
	}
}

func words(s string) []string {
	s = strings.TrimSuffix(strings.TrimSpace(s), ".")
	return strings.Fields(s)
}
