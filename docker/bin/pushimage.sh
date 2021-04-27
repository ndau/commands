#          name: Push ndauimage to ECR
            SHA=$1
            AWS_ACCOUNT="578681496768"
            ECR_REGION="us-east-1"
            
            # upload the image to S3 for public access
            docker tag ndauimage ndauimage:$SHA
            docker save ndauimage:$SHA -o ndauimage-$SHA.docker
            gzip -f ndauimage-$SHA.docker
            aws s3 cp ndauimage-$SHA.docker.gz s3://ndau-images/ndauimage-$SHA.docker.gz
            # update the current-*.txt file for the network we're deploying to
            # but only if we're deploying to master (e.g. don't do this for tagged pushes)
            # retag built image
            docker tag ndauimage $AWS_ACCOUNT.dkr.ecr.$ECR_REGION.amazonaws.com/ndauimage:$SHA
            docker rmi ndauimage:$SHA
            # push the image to ECR
            eval $(aws ecr get-login --no-include-email --region ${ECR_REGION})
            docker push $AWS_ACCOUNT.dkr.ecr.$ECR_REGION.amazonaws.com/ndauimage:$SHA