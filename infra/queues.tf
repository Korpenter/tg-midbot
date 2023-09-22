resource "yandex_message_queue" "midbot-in-ymq" {
  depends_on = [
    yandex_iam_service_account_static_access_key.ymq-static-key,
    yandex_resourcemanager_folder_iam_member.folder-editor,
  ]
  name = "midbot-in-ymq"

  access_key = yandex_iam_service_account_static_access_key.ymq-static-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.ymq-static-key.secret_key
}

resource "yandex_message_queue" "midbot-out-ymq" {
  depends_on = [
    yandex_iam_service_account_static_access_key.ymq-static-key,
    yandex_resourcemanager_folder_iam_member.folder-editor,
  ]
  name = "midbot-out-ymq"

  access_key = yandex_iam_service_account_static_access_key.ymq-static-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.ymq-static-key.secret_key
}

resource "yandex_iam_service_account_static_access_key" "ymq-static-key" {
  service_account_id = yandex_iam_service_account.folder-editor.id
  description        = "static access key for sqs"
}

