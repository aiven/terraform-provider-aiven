# selproj

A tool to select an empty project's suffix from a list of projects using a prefix.

## Usage

```bash
go run main.go
```

The tool requires the following environment variables to be set:

- `AIVEN_TOKEN` - Aiven authentication token
- `AIVEN_PROJECT_NAME_PREFIX` - A prefix to filter projects by

## Example

```text
$ AIVEN_TOKEN=... AIVEN_PROJECT_NAME_PREFIX=test go run -tags tools .
-project-2
```

## Testing

```bash
go test -v -tags tools ./...
```
