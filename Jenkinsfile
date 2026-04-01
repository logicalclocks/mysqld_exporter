@Library("jenkins-library@main")

import com.logicalclocks.jenkins.k8s.ImageBuilder

node("local") {

  stage('Clone repository') {
    checkout scm
  }

  stage('Build and push image') {
    def mysqldExporterVersion = readFile 'VERSION'
    withEnv(["VERSION=${mysqldExporterVersion.trim()}"]) {
        def builder = new ImageBuilder(this)
        m = readFile "build-manifest.json"
        builder.run(m)
    }
  }
}