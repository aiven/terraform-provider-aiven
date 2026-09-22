package io.aiven.test.tf;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import org.apache.kafka.common.config.ConfigDef;
import org.apache.kafka.connect.connector.Task;
import org.apache.kafka.connect.sink.SinkConnector;

public class NoopSinkConnector extends SinkConnector {
    private Map<String, String> props;

    @Override public String version() { return "1.0.0"; }
    @Override public void start(Map<String, String> props) { this.props = props; }
    @Override public Class<? extends Task> taskClass() { return NoopSinkTask.class; }

    @Override
    public List<Map<String, String>> taskConfigs(int maxTasks) {
        List<Map<String, String>> configs = new ArrayList<>();
        for (int i = 0; i < maxTasks; i++) {
            configs.add(props);
        }
        return configs;
    }

    @Override public void stop() {}
    @Override public ConfigDef config() { return new ConfigDef(); }
}
