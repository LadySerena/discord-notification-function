#!/usr/bin/env bash

gcloud functions deploy discord-notifier --entry-point GetBuildMessage --runtime go113 --trigger-http \
--set-env-vars=WEBHOOK_SECRET_NAME=build-notification-url,PROJECT_ID=telvanni-platform \
--service-account=tel-sa-discord-function@telvanni-platform.iam.gserviceaccount.com \
--max-instances=10