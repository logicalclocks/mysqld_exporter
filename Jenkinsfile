@Library("jenkins-library@main")

import com.logicalclocks.jenkins.k8s.ImageBuilder

node("local") {

  def goRoot = tool type: 'go', name: 'mysqld_exporter'

  stage('Clone repository') {
    checkout scm
  }

  stage('Build and push image') {
    withEnv(["GOROOT=${goRoot}", "PATH+GO=${goRoot}/bin"]) {
        sh 'make common-build'
    }
    def mysqldExporterVersion = readFile 'VERSION'
    withEnv(["VERSION=${mysqldExporterVersion.trim()}"]) {
        def builder = new ImageBuilder(this)
        m = readFile "build-manifest.json"
        builder.run(m)
    }
  }
}