resource "yandex_iam_service_account" "container-runner" {
  name = "container-runner"
}

resource "yandex_resourcemanager_folder_iam_member" "container_puller" {
  folder_id = var.yandex_folder_id
  member    = "serviceAccount:${yandex_iam_service_account.container-runner.id}"
  role      = "container-registry.images.puller"
}

resource "yandex_iam_service_account" "containers-manager" {
  name = "containers-manager"
}

resource "yandex_resourcemanager_folder_iam_member" "container-ydb-admin" {
  folder_id = var.yandex_folder_id
  member    = "serviceAccount:${yandex_iam_service_account.containers-manager.id}"
  role      = "ydb.admin"
}

resource "yandex_resourcemanager_folder_iam_member" "containers-invoker" {
  folder_id = var.yandex_folder_id
  member    = "serviceAccount:${yandex_iam_service_account.containers-manager.id}"
  role      = "serverless.containers.invoker"
}

output "container_sa_id" {
  value = yandex_iam_service_account.containers-manager.id
}

#resource "yandex_resourcemanager_folder_iam_member" "containers-editor" {
 # folder_id = var.yandex_folder_id
 # member    = "serviceAccount:${yandex_iam_service_account.containers-manager.id}"
#  role      = "serverless.containers.editor"
#}

resource "yandex_iam_service_account" "funstions-admin" {
  name = "funstions-admin"
}

resource "yandex_resourcemanager_folder_iam_member" "funstions-admin" {
  folder_id = var.yandex_folder_id
  member    = "serviceAccount:${yandex_iam_service_account.funstions-admin.id}"
  role      = "serverless.functions.admin"
}


resource "yandex_iam_service_account" "folder-editor" {
  name = "folder-editor"
}

resource "yandex_resourcemanager_folder_iam_member" "folder-editor" {
  folder_id = var.yandex_folder_id
  member    = "serviceAccount:${yandex_iam_service_account.folder-editor.id}"
  role      = "editor"
}
