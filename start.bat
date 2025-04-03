@echo off
echo Iniciando o script...

REM Finalizar qualquer instância anterior do ngrok
echo Finalizando instâncias anteriores do ngrok...
taskkill /im ngrok.exe /f > nul 2>&1

REM Iniciar o ngrok em segundo plano
echo Iniciando o ngrok...
start /b ngrok http 8080 > ngrok.log 2>&1
timeout /t 5 > nul

REM Capturar a URL pública do ngrok via API usando PowerShell
echo Capturando a URL pública do ngrok...
set FQDN=
for /f %%A in ('powershell -Command "(Invoke-WebRequest -Uri http://127.0.0.1:4040/api/tunnels).Content | ConvertFrom-Json | Select-Object -ExpandProperty tunnels | Where-Object { $_.proto -eq 'https' } | Select-Object -ExpandProperty public_url"') do set FQDN=%%A

REM Verificar se a URL foi capturada
if "%FQDN%"=="" (
    echo Erro: Não foi possível capturar a URL do ngrok.
    exit /b 1
)

echo URL capturada: %FQDN%

REM Atualizar o arquivo .env com a URL gerada
echo Atualizando o arquivo .env...
(
    echo NGROK_AUTHTOKEN=<id_ngrok>
    echo PORT=8080
    echo CLIENT_ID=<client id do app github>
    echo CLIENT_SECRET=<client secret do app github>
    echo FQDN=%FQDN%
) > .env

REM Rodar o servidor Go
echo Iniciando o servidor Go...
go run .