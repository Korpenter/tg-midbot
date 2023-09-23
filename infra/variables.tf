variable "yandex_token" {
  type        = string
  description = "IAM token to for Yandex Cloud. See https://cloud.yandex.com/en/docs/iam/operations/iam-token/create"
}

variable "yandex_cloud_id" {
  type = string
}

variable "yandex_folder_id" {
  type = string
}

variable "midbot_container_name" {
  type        = string
}

variable "tg_bot_token" {
  type        = string
  description = "Telegram bot token. See https://core.telegram.org/bots/features#creating-a-new-bot"
}

variable "image_registry_id" {
  type        = string
}

variable "midbot_image_digest" {
  type        = string
}

variable "midbot_function_hash" {
  type        = string
}

variable "midbot_notify_function_hash" {
  type        = string
}
