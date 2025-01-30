package template

// externalTemplates holds external templates for resources that are not part of the provider,
// but can be re-used across multiple tests
var externalTemplates = map[string]string{
	// AWS Templates
	"aws_vpc": `resource "aws_vpc" "{{ .resource_name }}" {
        cidr_block           = "{{ .cidr_block }}"
        enable_dns_hostnames = true
        enable_dns_support   = true
        tags = {
            Name = "{{ .vpc_name }}"
        }
    }`,
	"aws_route_table": `resource "aws_route_table" "{{ .resource_name }}" {
        vpc_id = {{ .vpc_id }}
        tags = {
            Name = "{{ .route_table_name }}"
        }
    }`,
	"aws_route": `resource "aws_route" "{{ .resource_name }}" {
        route_table_id            = {{ .route_table_id }}
        destination_cidr_block    = {{ .destination_cidr }}
        vpc_peering_connection_id = {{ .peering_connection_id }}
    }`,
	"aws_vpc_peering_accepter": `resource "aws_vpc_peering_connection_accepter" "{{ .resource_name }}" {
        vpc_peering_connection_id = {{ .peering_connection_id }}
        auto_accept              = true
        tags = {
            Name = "{{ .peering_name }}"
        }
    }`,

	// GCP Templates
	"gcp_provider": `provider "google" {
		project = "{{ .gcp_project }}"
		region  = "{{ .gcp_region }}"
	}`,

	"google_compute_network": `resource "google_compute_network" "{{ .resource_name }}" {
        name                    = "{{ .network_name }}"
        auto_create_subnetworks = {{ .auto_create_subnetworks }}
    }`,
}
