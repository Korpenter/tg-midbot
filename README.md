# MID Bot
![untitled](https://github.com/Korpenter/tg-midbot/assets/141184937/d128b43b-32ad-4df1-b9ee-bf0d327bc999)

A bot for the Russian consulate website.

**Version 0.2**  
Supported Features:
- Check application status
- Track application with notification on status change

## Prerequisites

- [YC CLI](https://cloud.yandex.com/en-ru/docs/cli/operations/install-cli)
- Terraform
- Golang
- Docker

## Quick Start

1. **Configure Terraform**  
   Follow the [Terraform Quickstart](https://cloud.yandex.com/en/docs/tutorials/infrastructure-management/terraform-quickstart) guide.

2. **Configure Docker**  
   Follow the [Docker Quickstart](https://cloud.yandex.com/en/docs/container-registry/quickstart/) guide.

3. **Set Up Container Registry**  
   Create a container registry and add its details to `.env`.

4. **Add Telegram Bot API Token**  
   Obtain a Telegram bot API token (see [Creating a New Bot](https://core.telegram.org/bots/features#creating-a-new-bot)) and add it to `.env`.

5. **Deploy the bot**  
   ```shell
   make all
   ```
7. **Destroy Teardown**
   ```shell
   make teardown
   ```
