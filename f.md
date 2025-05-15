  filter/drop_spanner_spans:
    error_mode: ignore
    traces:
      span:
        - |-
          resource.attributes["library.name"] == "cloud.google.com/go" and (IsMatch(name, "cloud.google.com/go/spanner*") or IsMatch(name, "Acquir*") or name == "Starting transaction attempt" or name == "exception")
        - |-
          resource.attributes["library.name"] == "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" and IsMatch(name, "google.spanner.v1.Spanner*")

