package io.aiven.test.tf;

import java.util.Collection;
import java.util.Map;
import org.apache.kafka.connect.sink.SinkRecord;
import org.apache.kafka.connect.sink.SinkTask;

public class NoopSinkTask extends SinkTask {
    @Override public String version() { return "1.0.0"; }
    @Override public void start(Map<String, String> props) {}
    @Override public void put(Collection<SinkRecord> records) {}
    @Override public void stop() {}
}
