@NonCPS
def cancelPreviousBuilds() {
  String jobname = env.JOB_NAME
  int currentBuildNum = env.BUILD_NUMBER.toInteger()

  def job = Jenkins.instance.getItemByFullName(jobname)
//   print('jobname: ' + jobname)
  for (build in job.builds) {
    //   print('build.getNumber(): ' + build.getNumber())

    if (build.isBuilding() && currentBuildNum > build.getNumber().toInteger()) {
      build.doStop();
      echo "Build ${build.toString()} cancelled"
    }
  }
}

pipeline {

    environment {
        registry = "197272534240.dkr.ecr.us-east-1.amazonaws.com"
        registryCredential = "ecr:us-east-1:aws_vert"
        dockerImageName = ""
    }

    agent {
        docker {
            image "akaytatsu/cibuilder:1.2.17"
        }
    }

    stages {

        stage('Init') {
            steps {
                cancelPreviousBuilds()
            }
        }

        stage('Code Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Build Docker Images') {
            steps {
                script {
                    sh 'cp -f src/.env.sample src/.env'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml build'
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml up -d --no-build'
                }
            }
        }

        stage('stop containers') {
            steps {
                script {
                    sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                }
            }
        }

        stage('build Container Register Develop') {
            when {
                expression {
                    return env.GIT_BRANCH == 'develop'
                }
            }

            steps {
                script {
                    docker.withRegistry("https://$registry", registryCredential) {
                        dockerImageName = "gocusign-hml"
                        dockerImage = docker.build(dockerImageName, "./src")
                        dockerImage.push("$BUILD_NUMBER")
                        dockerImage.push("latest")
                    }
                }

                script{
                    sh "docker rmi $registry/$dockerImageName:$BUILD_NUMBER"
                    sh "docker rmi $registry/$dockerImageName:latest"
                }
            }
        }

        stage('build Container Register Production') {
            when {
                expression {
                    return env.GIT_BRANCH == 'master'
                }
            }

            steps {
                script {
                    docker.withRegistry("https://$registry", registryCredential) {
                        dockerImageName = "gocusign-prd"
                        dockerImage = docker.build(dockerImageName, "./src")
                        dockerImage.push("$BUILD_NUMBER")
                        dockerImage.push("latest")
                    }
                }

                script{
                    sh "docker rmi $registry/$dockerImageName:$BUILD_NUMBER"
                    sh "docker rmi $registry/$dockerImageName:latest"
                }
            }
        }

        stage('Deploy to Homolog Environment') {
            when {
                expression {
                    return env.GIT_BRANCH == 'develop'
                }
            }

            steps {
                script {
                    withAWS(credentials:'aws_kube_homol', region: 'us-east-1') {
                        withKubeConfig([credentialsId: 'k8s_config_vert_homol']) {
                            sh "cat k8s/homologation/deployment.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
                        }
                    }
                }
            }
        }

        stage('Deploy to Production Environment') {
            when {
                expression {
                    return env.GIT_BRANCH == 'master'
                }
            }

            steps {
                script {
                    withAWS(credentials:'aws_kube_homol', region: 'us-east-1') {
                        withKubeConfig([credentialsId: 'k8s_config_vert_prod']) {
                            sh "cat k8s/production/deployment.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
                        }
                    }
                }
            }
        }

    }

    post {
        always {
            echo "Stop Docker image"
            script{
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
            }
        }

        success {
            echo "Notify bitbucket success"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
            }
        }

        failure {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
            }
        }

        aborted {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
            }
        }

        unsuccessful {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
            }
        }

    }
}