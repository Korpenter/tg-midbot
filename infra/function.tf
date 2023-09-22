resource "yandex_function_scaling_policy" "midbot-function-scaling-policy" {
  function_id = yandex_function.midbot-function.id
  policy {
    tag = "midbot"
    zone_instances_limit = 10
  }
}

resource "yandex_function" "midbot-function" {
  name               = "midbot-function"
  description        = "midbot function to send messages to users"
  runtime            = "golang119"
  entrypoint         = "main.Handler"
  memory             = "128"
  execution_timeout  = "5"
  tags               = ["midbot"]
  user_hash          = var.midbot_function_hash
  environment = {
    TELEGRAM_APITOKEN     = var.tg_bot_token,
  }
  content {
    zip_filename = "../midbot-out/dist/dist.zip"
  }
}

