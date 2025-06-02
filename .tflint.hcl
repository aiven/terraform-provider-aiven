plugin "terraform" {
  enabled = true
  preset = "recommended"
}

rule "terraform_required_version" {
  enabled = false
}

rule "terraform_typed_variables" {
  enabled = false
}
