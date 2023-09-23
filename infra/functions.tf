resource "yandex_function_scaling_policy" "midbot-function-scaling-policy" {
  function_id = yandex_function.midbot-function.id
  policy {
    tag = "midbot"
    zone_instances_limit = 10
  }
}
resource "yandex_function_scaling_policy" "midbot-notify-function-scaling-policy" {
  function_id = yandex_function.midbot-notify-function.id
  policy {
    tag = "midbot-notify"
    zone_instances_limit = 1
  }
}

resource "yandex_function" "midbot-function" {
  name               = "midbot-function"
  description        = "midbot function to send messages to users"
  runtime            = "golang119"
  entrypoint         = "main.Handler"
  memory             = "128"
  execution_timeout  = "5"
  tags               = ["midbot", "latest"]
  user_hash          = var.midbot_function_hash
  environment = {
    TELEGRAM_APITOKEN     = var.tg_bot_token,
  }
  content {
    zip_filename = "../midbot-out/dist/dist.zip"
  }
}

resource "yandex_function" "midbot-notify-function" {
  name               = "midbot-notify-function"
  description        = "midbot function to check application statuses and send notifications"
  runtime            = "golang119"
  entrypoint         = "main.Handler"
  memory             = "256"
  execution_timeout  = "50"
  tags               = ["midbot-notify"]
  user_hash          = var.midbot_notify_function_hash
  environment = {
    QUEUE_URL             = yandex_message_queue.midbot-out-ymq.id
    AWS_ACCESS_KEY_ID     = yandex_iam_service_account_static_access_key.ymq-static-key.access_key
    AWS_SECRET_ACCESS_KEY = yandex_iam_service_account_static_access_key.ymq-static-key.secret_key
    AWS_DEFAULT_REGION    = "ru-central1"
		YDB_ENDPOINT          = yandex_ydb_database_serverless.midbot-ydb.document_api_endpoint
  }
  content {
    zip_filename = "../midbot-notify/dist/dist.zip"
  }
}

