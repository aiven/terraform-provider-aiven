package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
)

type Grantee struct {
	User string
	Role string
}

func (g Grantee) equals(other Grantee) bool {
	return g.User == other.User && g.Role == other.Role
}

type PrivilegeGrant struct {
	Grantee   Grantee
	Database  string
	Table     string
	Column    string
	Privilege string
	WithGrant bool
}

type RoleGrant struct {
	Grantee Grantee
	Role    string
}

func userOrRole(g Grantee) string {
	if g.User != "" {
		return g.User
	}
	return g.Role
}

func CreateRoleGrant(
	ctx context.Context,
	client *aiven.Client,
	projectName string,
	serviceName string,
	grant RoleGrant,
) error {
	query := createRoleGrantStatement(grant)

	log.Println("[DEBUG] Clickhouse: create role grant query: ", query)
	_, err := client.ClickHouseQuery.Query(ctx, projectName, serviceName, defaultDatabase, query)
	return err
}

func RevokeRoleGrant(
	ctx context.Context,
	client *aiven.Client,
	projectName string,
	serviceName string,
	grant RoleGrant,
) error {
	query := revokeRoleGrantStatement(grant)

	log.Println("[DEBUG] privilege revocation query: ", query)
	_, err := client.ClickHouseQuery.Query(ctx, projectName, serviceName, defaultDatabase, query)
	return err
}

func ReadRoleGrants(
	ctx context.Context,
	client *aiven.Client,
	projectName string,
	serviceName string,
	grantee Grantee,
) ([]RoleGrant, error) {
	query := readRoleGrantsStatement()

	log.Println("[DEBUG] Clickhouse: read role grant query: ", query)
	r, err := client.ClickHouseQuery.Query(ctx, projectName, serviceName, defaultDatabase, query)
	if err != nil {
		return nil, err
	}

	roleGrants, err := roleGrantsFromAPIResponse(r)
	if err != nil {
		return nil, err
	}

	res := make([]RoleGrant, 0)
	for _, grant := range roleGrants {
		if !grant.Grantee.equals(grantee) {
			continue
		}
		res = append(res, grant)
	}
	return res, err
}

func CreatePrivilegeGrant(
	ctx context.Context,
	client *aiven.Client,
	projectName string,
	serviceName string,
	grant PrivilegeGrant,
) error {
	query := createPrivilegeGrantStatement(grant)

	log.Println("[DEBUG] Clickhouse: create privilege grant query: ", query)
	_, err := client.ClickHouseQuery.Query(ctx, projectName, serviceName, defaultDatabase, query)
	return err
}

func ReadPrivilegeGrants(
	ctx context.Context,
	client *aiven.Client,
	projectName string,
	serviceName string,
	grantee Grantee,
) ([]PrivilegeGrant, error) {
	query := readPrivilegeGrantsStatement()

	log.Println("[DEBUG] Clickhouse: read privilege grant query: ", query)
	r, err := client.ClickHouseQuery.Query(ctx, projectName, serviceName, defaultDatabase, query)
	if err != nil {
		return nil, err
	}

	privilegeGrants, err := privilegeGrantsFromAPIResponse(r)
	if err != nil {
		return nil, err
	}
	res := make([]PrivilegeGrant, 0)
	for _, grant := range privilegeGrants {
		if !grant.Grantee.equals(grantee) {
			continue
		}
		res = append(res, grant)
	}
	return res, err
}

func RevokePrivilegeGrant(
	ctx context.Context,
	client *aiven.Client,
	projectName string,
	serviceName string,
	grant PrivilegeGrant,
) error {
	query := revokePrivilegeGrantStatement(grant)

	log.Println("[DEBUG] privilege revocation query: ", query)
	_, err := client.ClickHouseQuery.Query(ctx, projectName, serviceName, defaultDatabase, query)
	return err
}

func createRoleGrantStatement(grant RoleGrant) string {
	return fmt.Sprintf("GRANT %s TO %s", escape(grant.Role), escape(userOrRole(grant.Grantee)))
}

func revokeRoleGrantStatement(grant RoleGrant) string {
	return fmt.Sprintf("REVOKE %s FROM %s", escape(grant.Role), escape(userOrRole(grant.Grantee)))
}

func readRoleGrantsStatement() string {
	return "SELECT * FROM system.role_grants"
}

func createPrivilegeGrantStatement(grant PrivilegeGrant) string {
	b := new(strings.Builder)

	b.WriteString("GRANT")
	b.WriteString(" ")
	b.WriteString(grant.Privilege)

	if len(grant.Column) > 0 {
		b.WriteString(fmt.Sprintf("(%s)", escape(grant.Column)))
	}

	// do not escape the asterisk as it is a wildcard
	if grant.Database == "*" {
		b.WriteString(" ON *")
	} else {
		b.WriteString(fmt.Sprintf(" ON %s", escape(grant.Database)))
	}

	if len(grant.Table) > 0 {
		b.WriteString(fmt.Sprintf(".%s", escape(grant.Table)))
	} else {
		b.WriteString(".*")
	}

	b.WriteString(fmt.Sprintf(" TO %s", escape(userOrRole(grant.Grantee))))

	if grant.WithGrant {
		b.WriteString(" WITH GRANT OPTION")
	}

	return b.String()
}

func revokePrivilegeGrantStatement(grant PrivilegeGrant) string {
	b := new(strings.Builder)

	b.WriteString("REVOKE")
	b.WriteString(" ")
	b.WriteString(grant.Privilege)

	if len(grant.Column) > 0 {
		b.WriteString(fmt.Sprintf("(%s)", escape(grant.Column)))
	}

	// do not escape the asterisk as it is a wildcard
	if grant.Database == "*" {
		b.WriteString(" ON *")
	} else {
		b.WriteString(fmt.Sprintf(" ON %s", escape(grant.Database)))
	}

	if len(grant.Table) > 0 {
		b.WriteString(fmt.Sprintf(".%s", escape(grant.Table)))
	} else {
		b.WriteString(".*")
	}

	b.WriteString(fmt.Sprintf(" FROM %s", escape(userOrRole(grant.Grantee))))
	return b.String()
}

func readPrivilegeGrantsStatement() string {
	return "SELECT * FROM system.grants"
}

func roleGrantsFromAPIResponse(r *aiven.ClickhouseQueryResponse) ([]RoleGrant, error) {
	meta := r.Meta
	data := r.Data

	columnNameMap := make(map[string]int)
	for i, md := range meta {
		columnNameMap[md.Name] = i
	}
	for _, columnName := range []string{
		"user_name",
		"role_name",
		"granted_role_name",
	} {
		if _, ok := columnNameMap[columnName]; !ok {
			return nil, fmt.Errorf("'system.role_grants' metadata is missing the '%s' column", columnName)
		}
	}

	var err error
	grants := make([]RoleGrant, 0)
	for i := range data {
		column := data[i].([]interface{})

		getMaybeString := func(columnName string) string {
			f := column[columnNameMap[columnName]]
			if f == nil {
				return ""
			}
			s, ok := f.(string)
			if !ok {
				err = fmt.Errorf("column name '%s' was expected to be a string", columnName)
				return ""
			}
			return s
		}

		grants = append(grants, RoleGrant{
			Grantee: Grantee{
				User: getMaybeString("user_name"),
				Role: getMaybeString("role_name"),
			},
			Role: getMaybeString("granted_role_name"),
		})
	}
	if err != nil {
		return nil, err
	}

	return grants, nil
}

func privilegeGrantsFromAPIResponse(r *aiven.ClickhouseQueryResponse) ([]PrivilegeGrant, error) {
	meta := r.Meta
	data := r.Data
	columnNameMap := make(map[string]int)
	for i, md := range meta {
		columnNameMap[md.Name] = i
	}
	for _, columnName := range []string{
		"is_partial_revoke",
		"user_name",
		"database",
		"table",
		"column",
		"role_name",
		"access_type",
		"grant_option",
	} {
		if _, ok := columnNameMap[columnName]; !ok {
			return nil, fmt.Errorf("'system.grants' metadata is missing the '%s' column", columnName)
		}
	}

	var err error
	grants := make([]PrivilegeGrant, 0)
	for i := range data {
		column := data[i].([]interface{})

		getMaybeString := func(columnName string, defaultValue string) string {
			f := column[columnNameMap[columnName]]
			if f == nil {
				return defaultValue
			}
			s, ok := f.(string)
			if !ok {
				err = fmt.Errorf("column name '%s' was expected to be a string", columnName)
				return ""
			}
			return s
		}

		getBoolean := func(columnName string) bool {
			s, ok := column[columnNameMap[columnName]].(json.Number)
			if !ok {
				err = fmt.Errorf("column name '%s' was expected to be a json number", columnName)
				return false
			}
			return s.String() == "1"
		}

		// skip partial revokes as we dont track them in the schema yet
		if getBoolean("is_partial_revoke") {
			continue
		}

		grants = append(grants, PrivilegeGrant{
			Grantee: Grantee{
				User: getMaybeString("user_name", ""),
				Role: getMaybeString("role_name", ""),
			},
			Database:  getMaybeString("database", "*"),
			Table:     getMaybeString("table", ""),
			Column:    getMaybeString("column", ""),
			Privilege: getMaybeString("access_type", ""),
			WithGrant: getBoolean("grant_option"),
		})
	}
	if err != nil {
		return nil, err
	}

	return grants, nil
}
