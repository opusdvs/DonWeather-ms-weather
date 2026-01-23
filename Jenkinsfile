pipeline {
    agent any

    environment {
        IMAGE_NAME = 'donweather-ms-weather'
        IMAGE_TAG = "dev"
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Go: Format Check') {
            steps {
                sh 'gofmt -l . | grep -v vendor || true'
            }
        }

        stage('Go: Vet') {
            steps {
                sh 'go vet ./...'
            }
        }

        stage('Go: Test') {
            steps {
                sh 'go test ./...'
            }
        }

        stage('Build Docker Image') {
            steps {
                script {
                    docker.build("${IMAGE_NAME}:${IMAGE_TAG}")
                }
            }
        }
    }
}
