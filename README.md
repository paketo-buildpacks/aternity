# `docker.io/paketobuildpacks/aternity`
The Paketo Buildpack for Aternity is a Cloud Native Buildpack that contributes the [AppInternals][a] Agent and configures it to connect to the service.

[a]: https://www.riverbed.com/products/steelcentral/steelcentral-appinternals.html

## Behavior
This buildpack will participate if all the following conditions are met

* A binding exists with `type` of `AppInternals`

The buildpack will do the following for Java applications:

* Contributes a Java agent to a layer and configures `$JAVA_TOOL_OPTIONS` to use it
* Transforms the contents of the binding secret to environment variables with the pattern `<KEY>=<VALUE>`

## Bindings
The buildpack optionally accepts the following bindings:

### Type: `dependency-mapping`
|Key                   | Value   | Description
|----------------------|---------|------------
|`<dependency-digest>` | `<uri>` | If needed, the buildpack will fetch the dependency with digest `<dependency-digest>` from `<uri>`

## License

This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
