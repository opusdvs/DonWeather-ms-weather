pipeline {
    agent {
        kubernetes {
            slaveConnectTimeout 300
            idleMinutes 10
            yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: go
    image: golang:1.24-alpine
    command: ['cat']
    tty: true
    workingDir: /home/jenkins/agent
    volumeMounts:
    - name: workspace-volume
      mountPath: /home/jenkins/agent
    resources:
      requests:
        memory: "256Mi"
        cpu: "500m"
      limits:
        memory: "2Gi"
        cpu: "2000m"
  
  - name: docker
    image: docker:dind
    command: ['dockerd-entrypoint.sh']
    securityContext:
      privileged: true
    volumeMounts:
    - name: docker-sock
      mountPath: /var/run
    - name: docker-storage
      mountPath: /var/lib/docker
    - name: workspace-volume
      mountPath: /home/jenkins/agent
    resources:
      requests:
        memory: "256Mi"
        cpu: "500m"
      limits:
        memory: "2Gi"
        cpu: "2000m"
  
  - name: docker-cli
    image: docker:cli
    command: ['cat']
    tty: true
    workingDir: /home/jenkins/agent
    volumeMounts:
    - name: workspace-volume
      mountPath: /home/jenkins/agent
    - name: docker-sock
      mountPath: /var/run
    resources:
      requests:
        memory: "256Mi"
        cpu: "250m"
      limits:
        memory: "1Gi"
        cpu: "1000m"
  
  volumes:
  - name: workspace-volume
    emptyDir: {}
  - name: docker-sock
    emptyDir: {}
  - name: docker-storage
    emptyDir: {}
"""
        }
    }

    environment {
        IMAGE_NAME = 'donweather-ms-weather'
        IMAGE_TAG = "dev"
        DOCKER_HOST = 'unix:///var/run/docker.sock'
        DOCKER_REGISTRY = 'buildbyte-container-registry.registry.twcstorage.ru'
    }

    stages {
        stage('Checkout') {
            steps {
                container('go') {
                    checkout scm
                }
            }
        }

        stage('Go: Format Check') {
            steps {
                container('go') {
                    sh 'gofmt -l . | grep -v vendor || true'
                }
            }
        }

        stage('Go: Vet') {
            steps {
                container('go') {
                    sh 'go vet ./...'
                }
            }
        }

        stage('Go: Test') {
            steps {
                container('go') {
                    sh 'go test ./...'
                }
            }
        }

        stage('Build Docker Image') {
            steps {
                container('docker-cli') {
                    script {
                        sh '''
                            # Ждем запуска Docker daemon
                            timeout 60 sh -c 'until docker info; do sleep 1; done'
                            docker build -t ${DOCKER_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG} .
                        '''
                    }
                }
            }
        }

        stage('Push Docker Image') {
            steps {
                container('docker-cli') {
                    script {
                        withCredentials([usernamePassword(credentialsId: 'docker-registry', usernameVariable: 'DOCKER_USERNAME', passwordVariable: 'DOCKER_PASSWORD')]) {
                            sh 'docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD} ${DOCKER_REGISTRY}'
                            sh 'docker push ${DOCKER_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}'

                            def digest = sh(
                                script: 'docker inspect --format="{{index .RepoDigests 0}}" ${DOCKER_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG} | cut -d "@" -f 2',
                                returnStdout: true
                            ).trim()
                            env.DOCKER_IMAGE_HASH = digest
                        }
                    }
                }
            }
        }
    }
    post {
        success {
            script {
                currentBuild.description = "Docker Image: ${env.DOCKER_IMAGE_HASH}"
            }
        }
    }
}