---
layout: default
title: Integration Endpoints User Config
parent: Resources Schema
nav_order: 4
---

This is the reference documentation for User Config objects within the Aiven API that relates to integration endpoints for integrations (such as External Elasticsearch)

## Properties

| Property                      | Type                                   | Required | Description |
|-------------------------------|----------------------------------------|----------|-------------|
| `datadog`                     | [object](#datadog)                     | **Yes**  |             |
| `external_elasticsearch_logs` | [object](#external_elasticsearch_logs) | **Yes**  |             |
| `jolokia`                     | [object](#jolokia)                     | **Yes**  |             |
| `prometheus`                  | [object](#prometheus)                  | **Yes**  |             |
| `rsyslog`                     | [object](#rsyslog)                     | **Yes**  |             |
| `signalfx`                    | [object](#signalfx)                    | **Yes**  |             |

## datadog

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property                 | Type                              | Required | Description |
|--------------------------|-----------------------------------|----------|-------------|
| `datadog_api_key`        | [object](#datadog_api_key)        | **Yes**  |             |
| `disable_consumer_stats` | [object](#disable_consumer_stats) | **Yes**  |             |
| `max_partition_contexts` | [object](#max_partition_contexts) | **Yes**  |             |
| `site`                   | [object](#site)                   | **Yes**  |             |

#### datadog_api_key

##### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `minLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

#### disable_consumer_stats

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### max_partition_contexts

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### site

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `enum`    | [array](#enum) | **Yes**  |             |
| `example` | string         | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | string         | **Yes**  |             |

## external_elasticsearch_logs

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property         | Type                      | Required | Description |
|------------------|---------------------------|----------|-------------|
| `ca`             | [object](#ca)             | **Yes**  |             |
| `index_days_max` | [object](#index_days_max) | **Yes**  |             |
| `index_prefix`   | [object](#index_prefix)   | **Yes**  |             |
| `timeout`        | [object](#timeout)        | **Yes**  |             |
| `url`            | [object](#url)            | **Yes**  |             |

#### ca

##### Properties

| Property    | Type           | Required | Description |
|-------------|----------------|----------|-------------|
| `example`   | string         | **Yes**  |             |
| `maxLength` | integer        | **Yes**  |             |
| `title`     | string         | **Yes**  |             |
| `type`      | [array](#type) | **Yes**  |             |

#### index_days_max

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | integer | **Yes**  |             |
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### index_prefix

##### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `default`    | string  | **Yes**  |             |
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `minLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

#### timeout

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | integer | **Yes**  |             |
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### url

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `minLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

## jolokia

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property              | Type                           | Required | Description |
|-----------------------|--------------------------------|----------|-------------|
| `basic_auth_password` | [object](#basic_auth_password) | **Yes**  |             |
| `basic_auth_username` | [object](#basic_auth_username) | **Yes**  |             |

#### basic_auth_password

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `minLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### basic_auth_username

##### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `minLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

## prometheus

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property              | Type                           | Required | Description |
|-----------------------|--------------------------------|----------|-------------|
| `basic_auth_password` | [object](#basic_auth_password) | **Yes**  |             |
| `basic_auth_username` | [object](#basic_auth_username) | **Yes**  |             |

#### basic_auth_password

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `minLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### basic_auth_username

##### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `minLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

## rsyslog

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property  | Type               | Required | Description |
|-----------|--------------------|----------|-------------|
| `ca`      | [object](#ca)      | **Yes**  |             |
| `cert`    | [object](#cert)    | **Yes**  |             |
| `format`  | [object](#format)  | **Yes**  |             |
| `key`     | [object](#key)     | **Yes**  |             |
| `logline` | [object](#logline) | **Yes**  |             |
| `port`    | [object](#port)    | **Yes**  |             |
| `sd`      | [object](#sd)      | **Yes**  |             |
| `server`  | [object](#server)  | **Yes**  |             |
| `tls`     | [object](#tls)     | **Yes**  |             |

#### ca

##### Properties

| Property    | Type           | Required | Description |
|-------------|----------------|----------|-------------|
| `example`   | string         | **Yes**  |             |
| `maxLength` | integer        | **Yes**  |             |
| `title`     | string         | **Yes**  |             |
| `type`      | [array](#type) | **Yes**  |             |

#### cert

##### Properties

| Property    | Type           | Required | Description |
|-------------|----------------|----------|-------------|
| `example`   | string         | **Yes**  |             |
| `maxLength` | integer        | **Yes**  |             |
| `title`     | string         | **Yes**  |             |
| `type`      | [array](#type) | **Yes**  |             |

#### format

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `default` | string         | **Yes**  |             |
| `enum`    | [array](#enum) | **Yes**  |             |
| `example` | string         | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | string         | **Yes**  |             |

#### key

##### Properties

| Property    | Type           | Required | Description |
|-------------|----------------|----------|-------------|
| `example`   | string         | **Yes**  |             |
| `maxLength` | integer        | **Yes**  |             |
| `title`     | string         | **Yes**  |             |
| `type`      | [array](#type) | **Yes**  |             |

#### logline

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `minLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### port

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | integer | **Yes**  |             |
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### sd

##### Properties

| Property    | Type           | Required | Description |
|-------------|----------------|----------|-------------|
| `example`   | string         | **Yes**  |             |
| `maxLength` | integer        | **Yes**  |             |
| `title`     | string         | **Yes**  |             |
| `type`      | [array](#type) | **Yes**  |             |

#### server

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `minLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### tls

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

## signalfx

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property           | Type                        | Required | Description |
|--------------------|-----------------------------|----------|-------------|
| `enabled_metrics`  | [object](#enabled_metrics)  | **Yes**  |             |
| `signalfx_api_key` | [object](#signalfx_api_key) | **Yes**  |             |
| `signalfx_realm`   | [object](#signalfx_realm)   | **Yes**  |             |

#### enabled_metrics

##### Properties

| Property   | Type              | Required | Description |
|------------|-------------------|----------|-------------|
| `default`  | [array](#default) | **Yes**  |             |
| `example`  | [array](#example) | **Yes**  |             |
| `items`    | [object](#items)  | **Yes**  |             |
| `maxItems` | integer           | **Yes**  |             |
| `title`    | string            | **Yes**  |             |
| `type`     | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### signalfx_api_key

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `minLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### signalfx_realm

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `default`   | string  | **Yes**  |             |
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `minLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |


