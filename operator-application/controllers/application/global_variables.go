package applicationcontroller

var secretName string
var deploymentName string
var serviceName string
var containerName string

const image = "docker.io/nheidloff/simple-microservice:latest"
const port int32 = 8081
const nodePort int32 = 30548
const labelKey = "app"
const labelValue = "myapplication"
const greetingMessage = "World"
const secretGreetingMessageLabel = "GREETING_MESSAGE"

// Note: For simplication purposes database properties are hardcoded
const databaseUser string = "name"
const databasePassword string = "password"
const databaseUrl string = "url"
const databaseCertificate string = "certificate"
