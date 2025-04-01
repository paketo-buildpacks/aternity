# `paketobuildpacks/aternity`
The Paketo Buildpack for Aternity is a Cloud Native Buildpack that contributes the [AppInternals][a] Agent and configures it to connect to the service.

[a]: https://www.riverbed.com/products/steelcentral/steelcentral-appinternals.html

## Behavior
This buildpack will participate if all the following conditions are met

* A binding exists with `type` of `AppInternals`

The buildpack will do the following for Java applications:

* Contributes a Java agent to a layer and configures `$JAVA_TOOL_OPTIONS` to use it
* Transforms the contents of the binding secret to environment variables with the pattern `<KEY>=<VALUE>`

## Notes

While this buildpack is packaged to support ARM64 there are no upstream bindings for Aternity on ARM64. If you try to use this buildpack on ARM64 it will fail attempting to download ARM64 binaries that do not exist. Please contact the vendor if you'd like to see ARM64 support. The buildpack is positioned to support it as soon as it is available upstream.

## Bindings
The buildpack optionally accepts the following bindings:

### Type: `dependency-mapping`
| Key                   | Value   | Description                                                                                       |
| --------------------- | ------- | ------------------------------------------------------------------------------------------------- |
| `<dependency-digest>` | `<uri>` | If needed, the buildpack will fetch the dependency with digest `<dependency-digest>` from `<uri>` |

## License

This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
