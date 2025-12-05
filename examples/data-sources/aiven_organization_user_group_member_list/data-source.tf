data "aiven_organization_user_group_member_list" "foo" {
  organization_id = var.organization_id
  user_group_id   = var.user_group_id
}
