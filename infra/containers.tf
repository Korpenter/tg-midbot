resource "yandex_ydb_database_serverless" "midbot-ydb" {
  name      = "midbot-ydb"
  folder_id = var.yandex_folder_id

  deletion_protection = false
}

resource "yandex_serverless_container" "midbot-container" {
  depends_on = [ 
    yandex_iam_service_account.containers-manager
  ]
  name               = var.midbot_container_name
  memory             = 256
  core_fraction      = 5
  execution_timeout  = "20s"
  service_account_id = yandex_iam_service_account.container-runner.id
  image {
    url = "cr.yandex/${var.image_registry_id}/${var.midbot_container_name}@${var.midbot_image_digest}"
    environment = {
      TELEGRAM_APITOKEN     = var.tg_bot_token,
      QUEUE_URL             = yandex_message_queue.midbot-out-ymq.id
      AWS_ACCESS_KEY_ID     = yandex_iam_service_account_static_access_key.ymq-static-key.access_key
      AWS_SECRET_ACCESS_KEY = yandex_iam_service_account_static_access_key.ymq-static-key.secret_key
      AWS_DEFAULT_REGION    = "ru-central1"
			YDB_ENDPOINT          = yandex_ydb_database_serverless.midbot-ydb.document_api_endpoint
    }
  }
}

output "midbot_container_id" {
  value = yandex_serverless_container.midbot-container.id
}