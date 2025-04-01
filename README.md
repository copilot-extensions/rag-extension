# Copilot RAG

## Descrição
Este projeto é uma aplicação Go que usa a geração aumentada por recuperação (RAG) em uma extensão baseada em agentes para o GitHub Copilot.

## Pré-requisitos

- Go 1.16 ou superior
- `ngrok`: [Instalar ngrok](https://ngrok.com/download)
- Configure as seguintes variáveis de ambiente (exemplo abaixo):

```bash
export PORT=8080
export CLIENT_ID=Iv1.0ae52273ad3193eb # ID da aplicação
export CLIENT_SECRET="seu_client_secret" # Gere um novo client secret para sua aplicação
export FQDN=https://6de513480979.ngrok.app # Use o ngrok para expor uma URL
```

## Instalação

1. Clone o repositório:

```bash
git clone git@github.com:copilot-extensions/rag-extension.git
cd rag-extension
```

2. Instale as dependências:

```bash
go mod tidy
```

## Uso

1. Inicie o `ngrok` com a porta configurada:

```bash
ngrok http http://localhost:8080
```

2. Configure as variáveis de ambiente (use a URL gerada pelo `ngrok` para o `FQDN`).
3. Execute a aplicação:

```bash
go run .
```

### Opcionalmente

- execute o arquivo start.bat
- copie a rota do ngrok para seguir os passos abaixo.

## Configurando o Agente no Chat

1. No **Copilot** na aba de configurações da sua aplicação (`https://github.com/settings/apps/<nome_da_aplicacao>/agent`):
   - Configure a URL com o endpoint `/agent` (ex.: `https://<sua-url-ngrok>.ngrok-free.app/agent`).
   - Configure a URL de Pré-Autorização com o endpoint `/auth/authorization` (ex.: `https://<sua-url-ngrok>.ngrok-free.app/auth/authorization`).

2. Na aba **Geral** das configurações da sua aplicação (`https://github.com/settings/apps/<nome_da_aplicacao>`):
   - Configure a URL de Callback com o endpoint `/auth/callback` (ex.: `https://<sua-url-ngrok>.ngrok-free.app/auth/callback`).
   - Configure a URL da Homepage com o endpoint base do `ngrok` (ex.: `https://<sua-url-ngrok>.ngrok-free.app`).

3. Certifique-se de que as permissões estão habilitadas em **Permissions & events**:
   - **Account Permissions** > **Copilot Chat** > **Access: Read Only**.

4. Instale sua aplicação em (`https://github.com/apps/<nome_da_aplicacao>`).
5. Agora, no GitHub Copilot (`https://github.com/copilot`), você pode mencionar seu agente usando o nome da sua aplicação.

### Diagrama Arquitetura

1. O **usuário** faz uma requisição para o servidor Go.
2. O servidor Go:
    - Processa arquivos e gera embeddings.
    - Lida com consultas ao agente.
    - Gerencia o fluxo de autenticação OAuth.
3. O **ngrok** expõe o servidor local para acesso externo.
4. O **Github** é usado para autenticação OAuth.

### Diagrama (Visual)

```
[Usuário] --> [Servidor Go]
    |               |
    |               +--> [Processamento de Arquivos]
    |               +--> [Geração de Embeddings]
    |               +--> [Consulta ao Agente]
    |
[ngrok] <-->  [Servidor Go]
[GitHub] <-->  [Servidor Go] (Autenticação OAuth)
```



## O Que Ele Pode Fazer

Teste o agente com os seguintes comandos:

| Descrição | Comando |
| --- | --- |
| Perguntar ao agente como configurar uma extensão do Copilot | `@agent Como configuro uma extensão do Copilot?` |
| Perguntar ao agente como é o formato de resposta de uma extensão do Copilot | `@agent Qual é o formato de resposta de uma extensão do Copilot?` |

## Documentação de Extensões do Copilot

- [Usando Extensões do Copilot](https://docs.github.com/en/copilot/using-github-copilot/using-extensions-to-integrate-external-tools-with-copilot-chat)
- [Sobre a construção de Extensões do Copilot](https://docs.github.com/en/copilot/building-copilot-extensions/about-building-copilot-extensions)
- [Processo de configuração](https://docs.github.com/en/copilot/building-copilot-extensions/setting-up-copilot-extensions)
- [Comunicando-se com a plataforma Copilot](https://docs.github.com/en/copilot/building-copilot-extensions/building-a-copilot-agent-for-your-copilot-extension/configuring-your-copilot-agent-to-communicate-with-the-copilot-platform)
- [Comunicando-se com o GitHub](https://docs.github.com/en/copilot/building-copilot-extensions/building-a-copilot-agent-for-your-copilot-extension/configuring-your-copilot-agent-to-communicate-with-github)
