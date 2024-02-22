package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"

	"github.com/aiven/terraform-provider-aiven/internal/server"
)

//go:generate go test -tags userconfig ./internal/schemautil/userconfig
//go:generate go run ./ucgenerator/... --services cassandra,clickhouse,flink,grafana,influxdb,kafka,kafka_connect,kafka_mirrormaker,m3aggregator,m3db,mysql,opensearch,pg,redis
//go:generate go run ./ucgenerator/... --integrations logs,kafka_mirrormaker,kafka_connect,kafka_logs,metrics,datadog,clickhouse_kafka,clickhouse_postgresql,external_aws_cloudwatch_metrics

// registryPrefix is the registry prefix for the provider.
const registryPrefix = "registry.terraform.io/"

// version is the version of the provider.
var version = "dev"

func main() {
	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")

	flag.Parse()
	ctx := context.Background()
	muxServer, err := server.NewMuxServer(ctx, version)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	if *debugFlag {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	name := registryPrefix + "aiven/aiven"

	//goland:noinspection GoBoolExpressions
	if version == "dev" {
		name = registryPrefix + "aiven-dev/aiven"
	}

	err = tf6server.Serve(
		name,
		func() tfprotov6.ProviderServer {
			return muxServer
		},
		serveOpts...,
	)

	if err != nil {
		log.Fatal(err)
	}
}
