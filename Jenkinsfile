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

def bitbucketNotify(status, branch_name, git_commit) {
    withCredentials([usernamePassword(credentialsId: 'thiagofreitas', usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
        sh "curl --location --request POST 'https://api.bitbucket.org/2.0/repositories/sistema_vert/gocusign/commit/"+git_commit+"/statuses/build'" \
            + " --user $USERNAME:$PASSWORD " \
            + " --header 'Content-Type: application/json' " \
            + " --data '{" \
            + "    \"state\": \""+status+"\"," \
            + "    \"key\": \""+git_commit+"\"," \
            + "    \"name\": \"Jenkins: "+branch_name+"\"," \
            + "    \"url\": \"https://ci.vert-capital.com/blue/organizations/jenkins/gocusign/activity\"" \
            + "}'"
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
                script {
                    bitbucketNotify('INPROGRESS', env.BRANCH_NAME, env.GIT_COMMIT)
                }
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

        stage('run python lint') {
            steps {
                script{
                    try{
                        sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml exec -T app isort . --check-only'
                        sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml exec -T app flake8 .'
                    }catch(e){
                        sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                        throw e
                    }
                }

            }
        }

        stage('run tests') {
            steps {
                script{
                    try{
                        sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml exec -T app pytest --cov --cov-report xml:coverage.xml'
                    }catch(e){
                        sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                        throw e
                    }
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
                        dockerImageName = "gocusign-stg"
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

        stage('build Container Register Homolog') {
            when {
                expression {
                    return env.GIT_BRANCH == 'homolog'
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

        stage('Deploy to Develop Environment') {
            when {
                expression {
                    return env.GIT_BRANCH == 'develop'
                }
            }

            steps {
                script {
                    withAWS(credentials:'aws_kube_homol', region: 'us-east-1') {
                        withKubeConfig([credentialsId: 'k8s_config_vert_homol']) {
                            sh "cat k8s/staging/deployment.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
                            sh "cat k8s/staging/deployment-worker.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
                            sh "cat k8s/staging/deployment-kafka.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
                        }
                    }
                }
            }
        }

        stage('Deploy to Homolog Environment') {
            when {
                expression {
                    return env.GIT_BRANCH == 'homolog'
                }
            }

            steps {
                script {
                    withAWS(credentials:'aws_kube_homol', region: 'us-east-1') {
                        withKubeConfig([credentialsId: 'k8s_config_vert_homol']) {
                            sh "cat k8s/homologation/deployment.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
                            sh "cat k8s/homologation/deployment-worker.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
                            sh "cat k8s/homologation/deployment-kafka.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
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
                            sh "cat k8s/production/deployment-worker.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
                            sh "cat k8s/production/deployment-kafka.yaml | sed 's/{{BUILD_NUMBER}}/$BUILD_NUMBER/g' | kubectl apply -f -"
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
                bitbucketNotify('SUCCESSFUL', env.BRANCH_NAME, env.GIT_COMMIT)
            }
        }

        failure {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                bitbucketNotify('FAILED', env.BRANCH_NAME, env.GIT_COMMIT)
            }
        }

        aborted {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                bitbucketNotify('FAILED', env.BRANCH_NAME, env.GIT_COMMIT)
            }
        }

        unsuccessful {
            echo "Notify bitbucket failure"
            script {
                sh 'docker-compose -f docker-compose.yml -f docker-compose.tests.yml down'
                bitbucketNotify('FAILED', env.BRANCH_NAME, env.GIT_COMMIT)
            }
        }

    }
}