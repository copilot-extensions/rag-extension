package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/copilot-extensions/rag-extension/agent"
	"github.com/copilot-extensions/rag-extension/config"
	"github.com/copilot-extensions/rag-extension/embeddings"
	"github.com/copilot-extensions/rag-extension/oauth"
	"github.com/joho/godotenv"
	"github.com/ledongthuc/pdf"
)

func init() {
	ngrokToken := os.Getenv("NGROK_AUTHTOKEN")
	if ngrokToken != "" {
		cmd := exec.Command("ngrok", "config", "add-authtoken", ngrokToken)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Erro ao configurar o token do ngrok: %v", err)
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}

	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	pubKey, err := fetchPublicKey()
	if err != nil {
		return fmt.Errorf("failed to fetch public key: %w", err)
	}

	config, err := config.New()
	if err != nil {
		return fmt.Errorf("error fetching config: %w", err)
	}

	me, err := url.Parse(config.FQDN)
	if err != nil {
		return fmt.Errorf("unable to parse HOST environment variable: %w", err)
	}
	me.Path = "auth/callback"

	oauthService := oauth.NewService(config.ClientID, config.ClientSecret, me.String())

	// ✅ Servir a interface HTML
	http.Handle("/", http.FileServer(http.Dir("public")))

	http.HandleFunc("/auth/authorization", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Requisição recebida: %s %s", r.Method, r.URL.Path)

		state := generateRandomState()

		http.SetCookie(w, &http.Cookie{
			Name:     "state",
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Secure:   strings.HasPrefix(config.FQDN, "https"),
		})

		authURL := fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s",
			url.QueryEscape(config.ClientID),
			url.QueryEscape(me.String()),
			url.QueryEscape(state),
		)

		http.Redirect(w, r, authURL, http.StatusFound)
	})

	http.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Requisição recebida: %s %s", r.Method, r.URL.Path)
		stateCookie, err := r.Cookie("state")
		if err != nil {
			http.Error(w, "state cookie not found", http.StatusBadRequest)
			return
		}

		state := r.URL.Query().Get("state")
		if state != stateCookie.Value {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		oauthService.PostAuth(w, r)
	})

	http.HandleFunc("/process-files", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		token := r.Header.Get("X-GitHub-Token")
		integrationID := r.Header.Get("Copilot-Integration-Id")

		if token == "" {
			cookie, err := r.Cookie("github_token")
			if err == nil {
				token = cookie.Value
				integrationID = "from-cookie"
				fmt.Println("🔄 Token recuperado via cookie.")
			}
		}

		if token == "" || integrationID == "" {
			http.Error(w, "Headers Copilot-Integration-Id e X-GitHub-Token são obrigatórios", http.StatusBadRequest)
			return
		}

		chunkSize := 20000
		chunkOverlap := 100
		embeddingService := embeddings.NewService(chunkSize, chunkOverlap, integrationID, token)

		rawDir := "data/raw"
		processedDir := "data/processed"
		embeddingDir := "embeddings/embeddings"

		err := processFiles(rawDir, processedDir, embeddingDir, embeddingService)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao processar arquivos: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Arquivos processados com sucesso!"}`))
	})

	agentService := agent.NewService(pubKey)
	http.HandleFunc("/agent", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Requisição recebida: %s %s", r.Method, r.URL.Path)
		agentService.ChatCompletion(w, r)
	})

	log.Printf("Servidor rodando na porta %s...", config.Port)
	return http.ListenAndServe(":"+config.Port, nil)
}

func processFiles(rawDir, processedDir, embeddingDir string, embeddingService *embeddings.Service) error {
	files, err := filepath.Glob(filepath.Join(rawDir, "*"))
	if err != nil {
		return fmt.Errorf("erro ao listar arquivos na pasta %s: %w", rawDir, err)
	}

	for _, filePath := range files {
		fileName := filepath.Base(filePath)
		contentStr, err := readFileContent(filePath)
		if err != nil {
			log.Printf("Erro ao ler o conteúdo do arquivo %s: %v", fileName, err)
			continue
		}

		err = embeddingService.ProcessAndStore(fileName, contentStr, embeddingDir)
		if err != nil {
			log.Printf("Erro ao processar embeddings para o arquivo %s: %v", fileName, err)
			continue
		}

		destPath := filepath.Join(processedDir, fileName)
		err = os.Rename(filePath, destPath)
		if err != nil {
			log.Printf("Erro ao mover o arquivo %s para %s: %v", fileName, processedDir, err)
			continue
		}

		log.Printf("Arquivo %s processado e movido para %s", fileName, processedDir)
	}

	return nil
}

func readFileContent(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".txt", ".md", ".html":
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("erro ao ler arquivo %s: %w", path, err)
		}
		return string(content), nil
	case ".pdf":
		f, r, err := pdf.Open(path)
		if err != nil {
			return "", fmt.Errorf("erro ao abrir PDF %s: %w", path, err)
		}
		defer f.Close()

		var builder strings.Builder
		numPages := r.NumPage()
		for i := 1; i <= numPages; i++ {
			p := r.Page(i)
			if p.V.IsNull() {
				continue
			}
			content := p.Content()
			for _, text := range content.Text {
				builder.WriteString(text.S)
				builder.WriteString(" ")
			}
		}

		return builder.String(), nil
	default:
		return "", fmt.Errorf("formato de arquivo não suportado: %s", ext)
	}
}

func fetchPublicKey() (*ecdsa.PublicKey, error) {
	resp, err := http.Get("https://api.github.com/meta/public_keys/copilot_api")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch public key: %s", resp.Status)
	}

	var respBody struct {
		PublicKeys []struct {
			Key       string `json:"key"`
			IsCurrent bool   `json:"is_current"`
		} `json:"public_keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	var rawKey string
	for _, pk := range respBody.PublicKeys {
		if pk.IsCurrent {
			rawKey = pk.Key
			break
		}
	}
	if rawKey == "" {
		return nil, fmt.Errorf("could not find current public key")
	}

	pubPemStr := strings.ReplaceAll(rawKey, "\\n", "\n")
	block, _ := pem.Decode([]byte(pubPemStr))
	if block == nil {
		return nil, fmt.Errorf("error parsing PEM block with GitHub public key")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ecdsaKey, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("GitHub key is not ECDSA")
	}

	return ecdsaKey, nil
}

func generateRandomState() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
