@@Meta.Output="/Jenkinsfile"

pipeline {
    agent {
        node {
            label 'maven'
        }
    }

    parameters {
        string(name: 'TAG_NAME', defaultValue: '', description: '')
    }

    environment {
        DOCKER_CREDENTIAL_ID = 'dockerhub-id'
        GITHUB_CREDENTIAL_ID = 'chsendev-github'
        KUBECONFIG_CREDENTIAL_ID = 'demo-kubeconfig'
        REGISTRY = '9.134.247.85:7070'
        DOCKERHUB_NAMESPACE = 'library' // change me
        GITHUB_ACCOUNT = 'chsendev' // change me
    }

    stages {
        stage ('checkout scm') {
            steps {
                checkout(scm)
            }
        }

        stage('build & push') {
            steps {
                container('maven') {
                    sh 'mvn clean package -DskipTests'
                    sh 'podman build -f Dockerfile-online -t $REGISTRY/$DOCKERHUB_NAMESPACE/{{.Config.ProjectName}}:SNAPSHOT-$BRANCH_NAME-$BUILD_NUMBER .'
                    withCredentials([usernamePassword(passwordVariable: 'DOCKER_PASSWORD', usernameVariable: 'DOCKER_USERNAME', credentialsId: "$DOCKER_CREDENTIAL_ID",)]) {
                        sh 'echo "$DOCKER_PASSWORD" | podman login --tls-verify=false $REGISTRY -u "$DOCKER_USERNAME" --password-stdin'
                        sh 'podman push --tls-verify=false $REGISTRY/$DOCKERHUB_NAMESPACE/{{.Config.ProjectName}}:SNAPSHOT-$BRANCH_NAME-$BUILD_NUMBER'
                    }
                }
            }
        }

        stage('push latest') {
            when {
                branch 'master'
            }
            steps {
                container('maven') {
                    sh 'podman tag $REGISTRY/$DOCKERHUB_NAMESPACE/{{.Config.ProjectName}}:SNAPSHOT-$BRANCH_NAME-$BUILD_NUMBER $REGISTRY/$DOCKERHUB_NAMESPACE/{{.Config.ProjectName}}:latest '
                    sh 'podman push --tls-verify=false $REGISTRY/$DOCKERHUB_NAMESPACE/{{.Config.ProjectName}}:latest '
                }
            }
        }

        stage('push with tag') {
            when {
                expression {
                    return params.TAG_NAME =~ /v.*/
                }
            }
            steps {
                container('maven') {
                    input(id: 'release-image-with-tag', message: 'release image with tag?')
                    withCredentials([usernamePassword(credentialsId: "$GITHUB_CREDENTIAL_ID", passwordVariable: 'GIT_PASSWORD', usernameVariable: 'GIT_USERNAME')]) {
                        sh 'git config --global user.email "kubesphere@yunify.com" '
                        sh 'git config --global user.name "kubesphere" '
                        sh 'git tag -a $TAG_NAME -m "$TAG_NAME" '
                        sh 'git push http://$GIT_USERNAME:$GIT_PASSWORD@github.com/$GITHUB_ACCOUNT/devops-maven-sample.git --tags --ipv4'
                    }
                    sh 'podman tag $REGISTRY/$DOCKERHUB_NAMESPACE/{{.Config.ProjectName}}:SNAPSHOT-$BRANCH_NAME-$BUILD_NUMBER $REGISTRY/$DOCKERHUB_NAMESPACE/{{.Config.ProjectName}}:$TAG_NAME '
                    sh 'podman push --tls-verify=false $REGISTRY/$DOCKERHUB_NAMESPACE/{{.Config.ProjectName}}:$TAG_NAME '
                }
            }
        }
    }
}
