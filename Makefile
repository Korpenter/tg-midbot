include .env

build_midbot:
	cd midbot-in && docker build -t cr.yandex/$(YC_IMAGE_REGISTRY_ID)/$(MIDBOT_CONTAINER_NAME) .

push_midbot: build_midbot
	cd midbot-in && docker push cr.yandex/$(YC_IMAGE_REGISTRY_ID)/$(MIDBOT_CONTAINER_NAME) | \
	grep "digest: sha256" | awk -F'digest: ' '{print $$2}' | cut -d' ' -f1 > ../.midbot_image_digest

zip_midbot_out:
	mkdir -p midbot-out/dist
	cd midbot-out && zip -r dist/dist.zip * -x "dist/*"
	sha256sum midbot-out/dist/dist.zip | cut -d' ' -f1 > .midbot_function_hash

webhook_info:
	curl --request POST --url "https://api.telegram.org/bot$(TELEGRAM_APITOKEN)/getWebhookInfo"

webhook_delete:
	curl --request POST --url "https://api.telegram.org/bot$(TELEGRAM_APITOKEN)/deleteWebhook"

webhook_create: webhook_delete
	curl --request POST --url "https://api.telegram.org/bot$(TELEGRAM_APITOKEN)/setWebhook" --header 'content-type: application/json' --data "{\"url\": \"$(shell cat infra/.gateway_url) \"}"

deploy_infra:
	$(shell sed 's/=.*/=/' .env > .env.example)
	cd infra && terraform init && \
	terraform apply -auto-approve \
	-var="yandex_token=$(YC_TOKEN)" \
	-var="yandex_cloud_id=$(YC_CLOUD_ID)" \
	-var="yandex_folder_id=$(YC_FOLDER_ID)" \
	-var="midbot_container_name=$(MIDBOT_CONTAINER_NAME)" \
	-var="image_registry_id=$(YC_IMAGE_REGISTRY_ID)" \
	-var="tg_bot_token=$(TELEGRAM_APITOKEN)" \
	-var="midbot_image_digest=$(shell cat .midbot_image_digest)" \
	-var="midbot_function_hash=$(shell cat .midbot_function_hash)"
	cd infra && terraform output -raw apigw-url > .gateway_url
	$(MAKE) webhook_create
	$(MAKE) create_timer_trigger

destroy_infra:
	cd infra && terraform init && \
	terraform destroy -auto-approve \
	-var="yandex_token=$(YC_TOKEN)" \
	-var="yandex_cloud_id=$(YC_CLOUD_ID)" \
	-var="yandex_folder_id=$(YC_FOLDER_ID)" \
	-var="midbot_container_name=$(MIDBOT_CONTAINER_NAME)" \
	-var="image_registry_id=$(YC_IMAGE_REGISTRY_ID)" \
	-var="tg_bot_token=$(TELEGRAM_APITOKEN)" \
	-var="midbot_image_digest=$(shell cat .midbot_image_digest)" \
	-var="midbot_function_hash=$(shell cat .midbot_function_hash)"

delete_timer_trigger:
	-yc serverless trigger delete midbot-container-timer

create_timer_trigger: delete_timer_trigger
	yc serverless trigger create timer \
	  --name "midbot-container-timer" \
	  --cron-expression '* * * * ? *' \
	  --payload "/notify" \
	  --invoke-container-id $(shell cd infra && terraform output --raw midbot_container_id) \
	  --invoke-container-service-account-id $(shell cd infra && terraform output  --raw container_sa_id) \
	  --retry-attempts 1 \
	  --retry-interval 10s

all: push_midbot zip_midbot_out deploy_infra

teardown: destroy_infra webhook_delete delete_timer_trigger