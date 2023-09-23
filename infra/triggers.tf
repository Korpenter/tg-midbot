resource "yandex_function_trigger" "midbot-trigger" {
  folder_id = var.yandex_folder_id
  name      = "midbot-trigger"
  message_queue {
    queue_id           = yandex_message_queue.midbot-in-ymq.arn
    batch_cutoff       = 1
    batch_size         = 1
    service_account_id = yandex_iam_service_account.folder-editor.id
  }
  container {
    id                 = yandex_serverless_container.midbot-container.id
    service_account_id = yandex_iam_service_account.containers-manager.id
  }
}

resource "yandex_function_trigger" "midbot-out-container-trigger" {
  folder_id = var.yandex_folder_id
  name      = "midbot-out-container-trigger"
  message_queue {
    queue_id           = yandex_message_queue.midbot-out-ymq.arn
    batch_cutoff       = 1
    batch_size         = 1
    service_account_id = yandex_iam_service_account.folder-editor.id
  }
  function {
    id                 = yandex_function.midbot-function.id
    tag                = "midbot"
    service_account_id = yandex_iam_service_account.funstions-admin.id
  }
}


resource "yandex_function_trigger" "midbot-notify-function-trigger" {
  folder_id = var.yandex_folder_id
  name      = "midbot-notify-function-trigger"
 timer {
   cron_expression = "0 * ? * * *"
 }
  function {
    id                 = yandex_function.midbot-notify-function.id
    tag                = "midbot-notify"
    service_account_id = yandex_iam_service_account.funstions-admin.id
  }
}