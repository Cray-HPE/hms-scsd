@Library('dst-shared@master') _

dockerBuildPipeline {
        githubPushRepo = "Cray-HPE/hms-scsd"
        repository = "cray"
        imagePrefix = "cray"
        app = "scsd"
        name = "hms-scsd"
        description = "Cray System Config Service"
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "csm"
}
