package agent

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"github.com/copilot-extensions/rag-extension/copilot"
)

type Service struct {
	pubKey *ecdsa.PublicKey
}

func NewService(pubKey *ecdsa.PublicKey) *Service {
	return &Service{pubKey: pubKey}
}

func (s *Service) ChatCompletion(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("Github-Public-Key-Signature")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(fmt.Errorf("falha ao ler o corpo da requisição: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	isValid, err := validPayload(body, sig, s.pubKey)
	if err != nil {
		fmt.Printf("falha ao validar a assinatura do payload: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isValid {
		http.Error(w, "assinatura do payload inválida", http.StatusUnauthorized)
		return
	}

	apiToken := r.Header.Get("X-GitHub-Token")
	integrationID := r.Header.Get("Copilot-Integration-Id")

	var req *copilot.ChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Printf("falha ao desserializar a requisição: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.generateCompletion(r.Context(), integrationID, apiToken, req, w); err != nil {
		fmt.Printf("falha ao executar o agente: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Service) generateCompletion(ctx context.Context, integrationID, apiToken string, req *copilot.ChatRequest, w io.Writer) error {
	var messages []copilot.ChatMessage

	for i := len(req.Messages) - 1; i >= 0; i-- {
		msg := req.Messages[i]
		if msg.Role != "user" || msg.Content == "" {
			continue
		}

		userEmbedding, err := createEmbedding(ctx, integrationID, apiToken, msg.Content)
		if err != nil {
			return fmt.Errorf("erro ao criar embedding para a mensagem do usuário: %w", err)
		}

		relevantChunks, err := s.retrieveRelevantChunks(userEmbedding, 3)
		if err != nil {
			return fmt.Errorf("erro ao recuperar chunks relevantes: %w", err)
		}

		context := "Use o seguinte contexto para responder à mensagem do usuário:\n\n"
		for _, chunk := range relevantChunks {
			text, err := os.ReadFile(strings.ReplaceAll(chunk, ".embedding", ".txt"))
			if err == nil {
				context += string(text) + "\n"
			}
		}

		messages = append(messages, copilot.ChatMessage{
			Role:    "system",
			Content: context,
		})
		break
	}

	messages = append(messages, req.Messages...)

	chatReq := &copilot.ChatCompletionsRequest{
		Model:    copilot.ModelGPT35,
		Messages: messages,
		Stream:   true,
	}

	stream, err := copilot.ChatCompletions(ctx, "copilot-chat", apiToken, chatReq)
	if err != nil {
		return fmt.Errorf("falha ao obter o stream de completions: %w", err)
	}
	defer stream.Close()

	reader := bufio.NewScanner(stream)
	for reader.Scan() {
		buf := reader.Bytes()
		_, err := w.Write(buf)
		if err != nil {
			return fmt.Errorf("falha ao escrever no stream: %w", err)
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return fmt.Errorf("falha ao escrever delimitador no stream: %w", err)
		}
	}

	if err := reader.Err(); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("falha ao ler do stream: %w", err)
	}

	return nil
}

func createEmbedding(ctx context.Context, integrationID, apiToken, content string) ([]float32, error) {
	embeddingResp, err := copilot.Embeddings(ctx, integrationID, apiToken, &copilot.EmbeddingsRequest{
		Model: copilot.ModelEmbeddings,
		Input: []string{content},
	})
	if err != nil {
		return nil, err
	}
	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("nenhum embedding retornado")
	}
	return embeddingResp.Data[0].Embedding, nil
}

func (s *Service) retrieveRelevantChunks(userEmbedding []float32, k int) ([]string, error) {
	embeddingDir := "embeddings/embeddings"
	files, err := filepath.Glob(filepath.Join(embeddingDir, "*.embedding"))
	if err != nil {
		return nil, fmt.Errorf("erro ao listar arquivos de embeddings: %w", err)
	}

	var results []struct {
		Chunk      string
		Similarity float32
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var embedding []float32
		err = json.Unmarshal(content, &embedding)
		if err != nil {
			continue
		}

		similarity := cosineSimilarityVectors(userEmbedding, embedding)
		results = append(results, struct {
			Chunk      string
			Similarity float32
		}{file, similarity})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	var topChunks []string
	for i := 0; i < k && i < len(results); i++ {
		topChunks = append(topChunks, results[i].Chunk)
	}

	return topChunks, nil
}

func cosineSimilarityVectors(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, magA, magB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dotProduct / (float32(math.Sqrt(float64(magA))) * float32(math.Sqrt(float64(magB))))
}

type asn1Signature struct {
	R *big.Int
	S *big.Int
}

func validPayload(data []byte, sig string, publicKey *ecdsa.PublicKey) (bool, error) {
	asnSig, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return false, err
	}
	parsedSig := asn1Signature{}
	rest, err := asn1.Unmarshal(asnSig, &parsedSig)
	if err != nil || len(rest) != 0 {
		return false, err
	}
	digest := sha256.Sum256(data)
	return ecdsa.Verify(publicKey, digest[:], parsedSig.R, parsedSig.S), nil
}
