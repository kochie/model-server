.PHONY: deps clean build

AWS_REGION="ap-southeast-1"

deps:
	go get -u ./...

clean: 
	rm -rf ./model/get_model/get_model
	rm -rf ./model/remove_model/remove_model
	rm -rf ./model/set_model/set_model
	
build:
	GOOS=linux GOARCH=amd64 go build -o model/get_model/get_model ./model/get_model
	GOOS=linux GOARCH=amd64 go build -o model/remove_model/remove_model ./model/remove_model
	GOOS=linux GOARCH=amd64 go build -o model/set_model/set_model ./model/set_model

package:
	sam package \
		--template-file ./template.yaml \
    	--s3-bucket sam-templates-robekoc \
    	--output-template-file packaged.yaml

deploy:
	aws cloudformation deploy \
		--template-file ./packaged.yaml \
		--stack-name model-server \
		--capabilities CAPABILITY_IAM

publish:
	sam publish \
	    --template packaged.yaml \
	    --region ${AWS_REGION}