@Library('dst-shared@release/shasta-1.4') _

dockerBuildPipeline {
        repository = "cray"
        imagePrefix = "cray"
        app = "scsd"
        name = "hms-scsd"
        description = "Cray System Config Service"
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "csm"
}
