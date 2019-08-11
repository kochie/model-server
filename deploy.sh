#!	/bin/bash

aws cloudformation deploy \
	--template-file /Users/robekoc/projects/model_server/packaged-template.yaml \
	--stack-name model-server \
	-–capabilities CAPABILITY_IAM
