# Kube Scraper

A project that lives in *Kubernetes* and scrapes website pages in a very
convenient way.

## Overview

The project is made of three components:

* *Telegram Bot*: You can find the telegram bot
[here](https://github.com/SunSince90/kube-scraper-telegram-bot). This pod
listens for messages sent by user and replies to them according to messages
defined by you. It inserts new user chats on a backend or removes them from
it if the write `/stop`. As of now, only *Firestore* is supported as a backend,
* Backend: You can find the backend pod
[here](https://github.com/SunSince90/kube-scraper-backend). This pod is just
an intermediary between a *scraper* and the actual backend, i.e. *Firestore*.
This is used to prevent having to write backend code on every *scraper* and
uses internal caching to prevent reaching quotas on the backend.
* *Scrapers*: The scraper is defined by this repository. Each scraper is
supposed to scrape one or more pages from the same website or from different
websites as long as the pages have the same html structure. So, in case you
want to scrape a product from different websites, you should deploy different
scrapers.

Feel free to fork the repository and adapt it as you wish. Be aware though that
I am not giving you any warranty about this, although you are welcome to create
issues, discussions and pull requests.

## Run on Kubernetes

This project is intended to work on Kubernetes and I am currently running it
on a *Raspberry Pi 4* running *k3s*.

### ... Or run locally

Nonetheless, you can also run it on your computer as so:

```bash
/scrape /pages/pages.yaml \
--telegram-token <telegram-token> \
--backend-address <address> \
--backend-port 80 \
--admin-chat-id <id> \
--pubsub-topic-name poll-result \
--gcp-service-account /credentials/service-account.json \
--gcp-project-id <project-id> \
--debug
```

## Example use cases

Suppose you want to monitor the price of a product on different websites.

You implement the `scrape` function as explained below differently for each
website page you want to monitor. Then you deploy them on your Kubernetes
cluster.

Whenever the price changes you can load the `ChatID`s from the backend, i.e.
*Firestore* and notify all your users about the price drop on *Telegram*.

## Install

First, learn how the
[Telegram Bot](https://github.com/SunSince90/kube-scraper-telegram-bot) works
and how to install it - also, learning how to create and manage a telegram bot
in general is useful.

Second, learn how the
[backend](https://github.com/SunSince90/kube-scraper-backend) works and how to
install it - also, since only *Firestore* is implemented for now, a good idea
is to learn how it works.

Then, clone the repository:

```bash
git clone https://github.com/SunSince90/kube-scraper.git
cd kube-scraper
```

## Implement

Create a new repository on your account and just copy the contents of
`main.go` and `scrape.go` included on the root folder of this project to the
root folder of your project.

You should only implement the `scrape` function on `scrape.go`, unless you want
to do some more advanced modifications.

The function receives:

* the `HandleOptions` from which you can receive the
*Google pubsub* client, the `ID` of the chat with the admin, the Telegram bot
client, and the backend client.
* The id of the poller that just finished the request (continue reading to know)
what it is.
* The response of the request that just finished.
* The error, if any.

Take a look at `/examples` to learn more.

## Deploy

Please note that the image that is going to be built will run on ARM, as it is
meant to run on a *Raspberry Pi*.
Make sure to edit the `Dockerfile` in case you want to build for another architecture.
Build the container image:

```bash
make docker-build docker-push IMG=<image>
```

### Create the namespace

Skip this if you already have this namespace on your cluster.

```bash
kubectl create namespace kube-scraper
```

### Create the telegram token secret

Skip this step if you already did this for the Telegram Bot.

```bash
kubectl create secret generic telegram-token \
--from-literal=token=<token> \
-n kube-scraper
```

### Create the project ID secret

Skip this step if you already did this for the Telegram Bot or the Backend.

```bash
kubectl create secret generic firebase-project-id \
--from-literal=project-id=<your-project-id> \
-n kube-scraper
```

### Create the service account secret

Skip this step if you already did this for the Telegram Bot or the Backend.

```bash
kubectl create secret generic gcp-service-account \
--from-file=service-account.json=<path-to-your-service-account> \
-n kube-scraper
```

### Create the admin chat id secret

```bash
kubectl create secret generic admin-chat-id \
--from-literal=chat-id=<id> \
-n kube-scraper
```

### Create the pages ConfigMap

Now, create the pages that you want this scraper to scrape. For example,
create the following yaml and call it `pages.yaml`:

```yaml
- id: "phone-12-pro"
  url: https://www.google.com/
  headers:
    "Accept": text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
    "Accept-Language": en-US,it-IT;q=0.8,it;q=0.5,en;q=0.3
    "Cache-Control": no-cache
    "Connection": keep-alive
    "Pragma": no-cache
  userAgentOptions:
    randomUA: true
  pollOptions:
    frequency: 15
- id: "phone-12-min"
  url: https://www.google.com/
  headers:
    "Accept": text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
    "Accept-Language": en-US,it-IT;q=0.8,it;q=0.5,en;q=0.3
    "Cache-Control": no-cache
    "Connection": keep-alive
    "Pragma": no-cache
  userAgentOptions:
    randomUA: true
  pollOptions:
    frequency: 30
```

Remember that each scraper is supposed to scrape just a website, and you should
implement and deploy other scrapers for other websites.

Now deploy this as a config map:

```bash
kubectl create configmap <name-of-the-configmap> \
--from-file=path/to/pages.yaml \
-n kube-scraper
```

### Create the deployment

Take a look at `volumes` in `deploy/deployment.yaml`:

```yaml
      volumes:
      - name: gcp-service-account
        secret:
          secretName: gcp-service-account
      - name: scrape-pages
        configMap:
          name: <name-of-the-configmap>
```

Replace `<name-of-the-configmap>` with the name of the `ConfigMap` you created
in the [ConfigMap](#create-the-pages-configmap).

Now at `env`:

```yaml
        env:
          - name: TELEGRAM_TOKEN
            valueFrom:
              secretKeyRef:
                name: telegram-token
                key: token
          - name: FIREBASE_PROJECT_ID
            valueFrom:
              secretKeyRef:
                name: firebase-project-id
                key: project-id
          - name: ADMIN_CHAT_ID
            valueFrom:
              secretKeyRef:
                name: admin-chat-id
                key: chat-id
```

Remove this if you are not using it. These values are using in `command` as the
already included `deployment.yaml` file. Add or remove values as you see fit.

Replace `<image>` from `deploy/deployment.yaml` with the container image you
published earlier and then:

```bash
kubectl create -f deploy/deployment.yaml
```
