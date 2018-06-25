package view.run_input_required

node {

 checkout scm
    imageName = "${env.BRANCH_NAME}-${env.BUILD_ID}"
    stage('Build Docker') {
        docker.withRegistry('http://h1.ywopt.com', '1efee65b-bc30-4bfa-9c40-222dd648e4a5') {
            def customImage = docker.build("h1.ywopt.com/ywopt.com/qpush:${imageName}", "-f ./deploy/dockerfile/Dockerfile ./")
            customImage.push()
        }

    }

     timeout(1){

         stage ('Promotion'){

             input '发布？'
         }


         stage ('Deploy'){

             if (env.BRANCH_NAME == 'master') {

                 sh "sed 's/qpush:latest/qpush:${imageName}/' ${env.WORKSPACE}/deploy/kubectl-yaml/server-deployment.yaml | /usr/local/bin/kubectl apply -f - "

             } else if(env.BRANCH_NAME == 'deployment1') {


                 sh "sed 's/carrier:latest/qpush:${imageName}/' ${env.WORKSPACE}/deployments/carrier.yaml | /usr/local/bin/kubectl apply -f - "


             }
         }


     }

}