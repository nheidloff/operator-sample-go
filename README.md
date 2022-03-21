# operator-sample-go

This project contains Kubernetes operator samples that demonstrate best practices how to develop operators with Go and the Operator SDK.

### Setup

The project contains a custom application controller, a database controller simulating an external resource and a sample microservice. The easiest way to get started is to run the application controller only and refer to the database resource definition and the built microservice image on Docker.io.

Follow these [instructions](operator-application/README.md#setup-and-local-usage) to run the application controller.

See the readme documents for all components:

* [Database operator](operator-application/README.md)
* [Microservice](simple-microservice/README.md)
* [Application operator](operator-application/README.md)

### Documentation

* [Creating and updating Resources from Operators](http://heidloff.net/article/updating-resources-kubernetes-operators/)
* [Deleting Resources in Operators](http://heidloff.net/article/deleting-resources-kubernetes-operators/)
* [Storing State of Resources with Conditions](http://heidloff.net/article/storing-state-status-kubernetes-resources-conditions-operators-go/)
* [Accessing third Party Custom Resources in Go Operators](http://heidloff.net/article/accessing-third-party-custom-resources-go-operators/)
* [Finding out the Kubernetes Versions and Capabilities in Operators](http://heidloff.net/article/finding-kubernetes-version-capabilities-operators/)
* [Importing Go Modules in Operators](http://heidloff.net/article/importing-go-modules-kubernetes-operators/)
* [Using object-oriented Concepts in Golang based Operators](http://heidloff.net/article/object-oriented-concepts-golang/)
* [Manually deploying Operators to Kubernetes](http://heidloff.net/article/manually-deploying-operators-to-kubernetes/)
* [Deploying Operators with the Operator Lifecycle Manager](http://heidloff.net/article/deploying-operators-operator-lifecycle-manager-olm/)

### Undocumented Capabilities

* Setup of watchers
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L218)

### Capabilities To Be Added

* Deployment/bundling
* Setup of RBAC
* Life cycle manager
* Versioning
* Webhooks
* Metrics
* Events
* Operator cleanup
* Scope: namespace vs global
* Phase 3 - 5
* Testing
* Leader strategy
* Creations of database schemas
* Templates for customizability
* Stateful resources (via uid, resourceVersion and in-memory store)

### Resources

* [Operator SDK](https://sdk.operatorframework.io/docs/overview/)
* [Operator SDK Best Practices](https://sdk.operatorframework.io/docs/best-practices/best-practices/)
* [Intro to the Operator Lifecycle Manager](https://www.youtube.com/watch?v=5PorcMTYZTo)
* [Go Modules](https://www.youtube.com/watch?v=Z1VhG7cf83M)
* [Resources to build Kubernetes Operators](http://heidloff.net/articles/resources-to-build-kubernetes-operators/)
* [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
* [Quarkus sample](https://github.com/nheidloff/quarkus-operator-microservice-database)