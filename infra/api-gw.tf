resource "yandex_api_gateway" "midbot-api-gateway" {
  name        = "midbot-api-gw"
  description = "midbot API Gateway"
  spec        = <<-EOT
openapi: 3.0.0
info:
  title: midbot API
  version: 1.0.0
paths:
  /:
    post:
      x-yc-apigateway-integration:
        type: cloud_ymq
        action: SendMessage
        queue_url: ${yandex_message_queue.midbot-in-ymq.id}
        folder_id: ${var.yandex_folder_id}
        service_account_id: ${yandex_iam_service_account.folder-editor.id}
EOT
}

output "apigw-url" {
  value = "https://${yandex_api_gateway.midbot-api-gateway.domain}/"
}