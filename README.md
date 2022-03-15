# operator-sample-go

This project contains Kubernetes operator samples that demonstrate best practices how to develop operators with Go and the Operator SDK.

### Setup

The project contains a custom application controller, a database controller simulating an external resource and a sample microservice. The easiest way to get started is to run the application controller only and refer to the database resource definition and the built microservice image on Docker.io.

Follow these [instructions](operator-application/README.md#setup-and-usage) to run the application controller.

See the readme documents for all components:

* [Database operator](operator-application/README.md)
* [Microservice](simple-microservice/README.md)
* [Application operator](operator-application/README.md)

### Current Capabilities

* Kubernetes version and capabilities
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L349)
* Status updates and conditions
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/ca5a50763cf36d4d74786119573a4e4865d4a942/operator-application/controllers/application_controller.go#L564), [snippet](https://github.com/nheidloff/operator-sample-go/blob/ca5a50763cf36d4d74786119573a4e4865d4a942/operator-application/controllers/application_controller.go#L503), [snippet](https://github.com/nheidloff/operator-sample-go/blob/ca5a50763cf36d4d74786119573a4e4865d4a942/operator-application/controllers/application_controller.go#L240)
* Deletions via child first strategy
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L266)
* Deletions via programmatic strategy, for example for external resources
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L101), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L379), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L206)
* Accessing third party custom resources
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L26), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L117), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L270), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/main.go#L31)
* Updates of deployed resources
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/ca5a50763cf36d4d74786119573a4e4865d4a942/operator-application/controllers/application_controller.go#L194), [snippet](https://github.com/nheidloff/operator-sample-go/blob/ca5a50763cf36d4d74786119573a4e4865d4a942/operator-application/controllers/application_controller.go#L591)
* Setup of watchers
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L218)

### Capabilities to be added

* Deployment/bundling
* Setup of RBAC
* Life cycle manager
* Versioning
* Webhooks
* Metrics
* Scope: namespace vs global
* Phase 3 - 5
* Testing
* Leader strategy
* Creations of database schemas
* Templates for customizability
* Stateful resources (via uuid, resourceVersion and in-memory store)

### Go Development Techniques to be documented

* IDE usage and tips
* Debugging
* Where to find documentation
* Imports and packages
* Pointers
* Constants

### Blogs

* [Deleting Resources in Kubernetes Operators](http://heidloff.net/article/deleting-resources-kubernetes-operators/)
* [Accessing third Party Custom Resources in Go Operators](http://heidloff.net/article/accessing-third-party-custom-resources-go-operators/)
* [Finding out the Kubernetes Version in Operators](http://heidloff.net/article/finding-kubernetes-version-capabilities-operators/)
* [Creating Database Schemas in Kubernetes Operators](http://heidloff.net/article/creating-database-schemas-kubernetes-operators/)
* [Resources to build Kubernetes Operators](http://heidloff.net/articles/resources-to-build-kubernetes-operators/)
* [Updating Resources from Kubernetes Operators](http://heidloff.net/article/updating-resources-kubernetes-operators/)
* [Storing State of Kubernetes Resources with Conditions](http://heidloff.net/article/storing-state-status-kubernetes-resources-conditions-operators-go/)