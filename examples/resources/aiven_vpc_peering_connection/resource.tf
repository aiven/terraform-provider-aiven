resource "aiven_vpc_peering_connection" "mypeeringconnection" {
    vpc_id = aiven_project_vpc.myvpc.id
    peer_cloud_account = "<PEER_ACCOUNT_ID>"
    peer_vpc = "<PEER_VPC_ID/NAME>"
    peer_region = "<PEER_REGION>"

    timeouts {
        create = "10m"
    }
}

