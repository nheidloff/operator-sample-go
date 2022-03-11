# operator-sample-go

This project contains Kubernetes operator samples that demonstrate best practices how to develop operators with Go and the Operator SDK.

Work in progress ...

See the readmes for draft setup instructions:

* [Application operator](operator-application/ReadMe.md)
* [Database operator](operator-application/ReadMe.md)
* [Microservice](simple-microservice/ReadMe.md)

Current Capabilities:

* Kubernetes version and capabilities
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L349)
* Status updates and conditions
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L390)
* Deletions via child first strategy
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L266)
* Deletions via programmatic strategy, for example for external resources
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L101), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L379), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L206)
* Accessing third party custom resources
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L26), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L117), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L270), [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/main.go#L31)
* Updates of deployed resources
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L170)
* Setup of watchers
    * Code: [snippet](https://github.com/nheidloff/operator-sample-go/blob/aa9fd15605a54f712e1233423236bd152940f238/operator-application/controllers/application_controller.go#L218)

To be added:

* Versioning
* Change checks via hash
* Creations of database schemas
* Setup of RBAC
* Metrics
* Scope: namespace vs global
* Life cycle manager
* Phase 3 - 5
* Webhooks

Go Development Techniques:

* IDE usage and tips
* Debugging
* Where to find documentation
* Imports and packages
* Pointers
* Constants