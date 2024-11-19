# Retrevial Augmented Generation Extensions Sample

> [!NOTE]
> Copilot Extensions are in public beta and may be subject to change.

## Description
This project is a Go application that demonstrates how to use retrevial augmented generation in an agent-based GitHub Copilot Extension.

## Prerequisites

- Go 1.16 or higher
- Set the following environment variables (example below):

```
export PORT=8080
export CLIENT_ID=Iv1.0ae52273ad3193eb // the application id
export CLIENT_SECRET="your_client_secret" // generate a new client secret for your application
export FQDN=https://6de513480979.ngrok.app // use ngrok to expose a url
```

## Installation:
1. Clone the repository: 

```
git clone git@github.com:copilot-extensions/rag-extension.git
cd rag-extension
```

2. Install dependencies:

```
go mod tidy
```

## Usage

1. Start up ngrok with the port provided:

```
ngrok http http://localhost:8080
```

2. Set the environment variables (use the ngrok generated url for the `FDQN`)
3. Run the application:

```
go run .
```

## Accessing the Agent in Chat:

1. In the `Copilot` tab of your Application settings (`https://github.com/settings/apps/<app_name>/agent`)
- Set the URL that was set for your FQDN above with the endpoint `/agent` (e.g. `https://6de513480979.ngrok.app/agent`)
- Set the Pre-Authorization URL with the endpoint `/auth/authorization` (e.g. `https://6de513480979.ngrok.app/auth/authorization`)
2. In the `General` tab of your application settings (`https://github.com/settings/apps/<app_name>`)
- Set the `Callback URL` with the `/auth/callback` endpoint (e.g. `https://6de513480979.ngrok.app/auth/callback`)
- Set the `Homepage URL` with the base ngrok endpoint (e.g. `https://6de513480979.ngrok.app/auth/callback`)
3. Ensure your permissions are enabled in `Permissions & events` > 
- `Account Permissions` > `Copilot Chat` > `Access: Read Only`
4. Ensure you install your application at (`https://github.com/apps/<app_name>`)
5. Now if you go to `https://github.com/copilot` you can `@` your agent using the name of your application.

## What Can It Do

Test out the agent with the following commands!

| Description | Prompt |
| --- |--- |
| User asking `@agent` how to configure a Copilot extension | `@agent How do I configure a copilot extension?` |
| User asking `@agent` what a Copilot extension looks like | `@agent What is the response format for a copilot extension?` |

## Copilot Extensions Documentation
- [Using Copilot Extensions](https://docs.github.com/en/copilot/using-github-copilot/using-extensions-to-integrate-external-tools-with-copilot-chat)
- [About building Copilot Extensions](https://docs.github.com/en/copilot/building-copilot-extensions/about-building-copilot-extensions)
- [Set up process](https://docs.github.com/en/copilot/building-copilot-extensions/setting-up-copilot-extensions)
- [Communicating with the Copilot platform](https://docs.github.com/en/copilot/building-copilot-extensions/building-a-copilot-agent-for-your-copilot-extension/configuring-your-copilot-agent-to-communicate-with-the-copilot-platform)
- [Communicating with GitHub](https://docs.github.com/en/copilot/building-copilot-extensions/building-a-copilot-agent-for-your-copilot-extension/configuring-your-copilot-agent-to-communicate-with-github)
