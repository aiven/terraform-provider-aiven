#!/usr/bin/env bash
# Rebuilds testdata/noop-sink-connector.jar, the minimal Kafka Connect plugin the acceptance
# tests upload. Needs a JDK and the connect-api and kafka-clients jars of any Kafka release:
#   KAFKA_LIBS=/path/to/kafka/libs ./build.sh
set -euo pipefail

cd "$(dirname "$0")"
: "${KAFKA_LIBS:?set KAFKA_LIBS to a directory with connect-api-*.jar and kafka-clients-*.jar}"

classes=$(mktemp -d)
trap 'rm -rf "$classes"' EXIT

# Joined with tr rather than paste -sd:, which needs an explicit "-" on macOS.
classpath=$(ls "$KAFKA_LIBS"/connect-api-*.jar "$KAFKA_LIBS"/kafka-clients-*.jar | tr '\n' ':')

javac --release 17 -cp "$classpath" -d "$classes" src/io/aiven/test/tf/*.java

# Without this descriptor the plugin isn't discovered as a sink connector.
mkdir -p "$classes/META-INF/services"
echo "io.aiven.test.tf.NoopSinkConnector" >"$classes/META-INF/services/org.apache.kafka.connect.sink.SinkConnector"

rm -f noop-sink-connector.jar
(cd "$classes" && jar cf "$OLDPWD/noop-sink-connector.jar" .)
