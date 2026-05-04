package main

import (
	"context"
	"langchain-go-ollama/internal/documents"
	workflow "langchain-go-ollama/internal/graph"
	"langchain-go-ollama/internal/rag"
	"log"

	_ "langchain-go-ollama/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"github.com/tmc/langchaingo/llms/ollama"
)

// @title Weather Workflow API
// @version 1.0
// @description This API runs LangGraph workflow with Ollama
// @host localhost:8080
// @BasePath /
func main() {
	app := fiber.New()
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html")
	})

	ctx := context.Background()

	connString := "postgres://langchain:langchain@localhost:6024/langchain"
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}

	// verify connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	store, err := rag.NewPGVectorStore(pool)
	if err != nil {
		log.Fatal(err)
	}

	defer store.Close()

	ollamaLLM, err := ollama.New(ollama.WithModel("gemma4:31b-cloud"))
	if err != nil {
		log.Fatal(err)
	}

	workflowService, err := workflow.NewWorkflowService(ollamaLLM)
	if err != nil {
		log.Fatal(err)
	}

	ollamaEmbed, err := rag.NewOllamaEmbedder("qwen3-embedding:0.6b")
	if err != nil {
		log.Fatal(err)
	}

	docRepo, err := documents.NewPostgresDocumentRepository(pool)
	if err != nil {
		log.Fatal(err)
	}

	ingestion, err := rag.NewIngestionService(store, ollamaEmbed, docRepo)
	if err != nil {
		log.Fatal(err)
	}

	retriever, err := rag.NewRetriever(store, ollamaEmbed)
	if err != nil {
		log.Fatal(err)
	}

	generator, err := rag.NewGenerator(ollamaLLM)
	if err != nil {
		log.Fatal(err)
	}

	ragService, err := rag.NewService(retriever, generator)
	if err != nil {
		log.Fatal(err)
	}

	app.Post("/weather", getWeatherHandler(workflowService))
	app.Post("/embed", getTextEmbed(ollamaEmbed))
	app.Post("/documents/raw", insertDocuments(ingestion))
	app.Post("/documents/search", similaritySearch(retriever))
	app.Post("/documents/ask", askHandler(ragService))

	log.Fatal(app.Listen(":8080"))
}

// getWeatherHandler godoc
// @Summary Get weather summary
// @Description Run LangGraph workflow for weather query
// @Tags weather
// @Accept json
// @Produce json
// @Param request body workflow.UserInput true "Weather Query"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /weather [post]
func getWeatherHandler(service *workflow.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req workflow.UserInput
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		result, err := service.Run(c.Context(), req.Query)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(Response{Data: result})
	}
}

type Response struct {
	Data  any `json:"data"`
	Count int `json:"count"`
}

// getTextEmbed godoc
// @Summary Generate text embedding
// @Description Convert input text into embedding vector using Ollama
// @Tags embedding
// @Accept json
// @Produce json
// @Param request body rag.UserInput true "Text input"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /embed [post]
func getTextEmbed(service *rag.OllamaEmbedder) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req rag.UserInput
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		result, err := service.EmbedBatch(c.Context(), req.Query)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(Response{Data: result})
	}
}

// InsertDocuments godoc
// @Summary Insert documents into vector store
// @Description Insert documents with embeddings into pgvector (LangChain schema)
// @Tags vector
// @Accept json
// @Produce json
// @Param request body rag.IngestionRequest true "Documents to insert"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /documents/raw [post]
func insertDocuments(ingestion *rag.IngestionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req rag.IngestionRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		// insert into store
		docs, err := ingestion.Ingest(c.Context(), req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(Response{
			Data:  docs,
			Count: len(docs),
		})
	}
}

// similaritySearch godoc
// @Summary Semantic similarity search
// @Description Perform vector similarity search using an embedded query and return top-k matching chunks
// @Tags vector
// @Accept json
// @Produce json
// @Param request body rag.SearchRequest true "Search query request"
// @Success 200 {object} Response{data=[]rag.VectorDocument}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /documents/search [post]
func similaritySearch(retriever *rag.Retriever) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req rag.SearchRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		// search the query
		docs, err := retriever.Retrieve(c.Context(), req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(Response{
			Data:  docs,
			Count: len(docs),
		})
	}
}

// askHandler godoc
// @Summary Ask question using RAG
// @Description Retrieve relevant documents using vector search and generate an answer using LLM
// @Tags rag
// @Accept json
// @Produce json
// @Param request body rag.SearchRequest true "RAG query request"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /documents/ask [post]
func askHandler(ragService *rag.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req rag.SearchRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid request",
			})
		}

		answer, err := ragService.Ask(c.Context(), req.Query, req.K)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(Response{
			Data:  answer,
			Count: len(answer),
		})
	}
}
