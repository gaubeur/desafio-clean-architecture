package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/vektah/gqlparser/v2/ast"

	// Importe o pacote necessário para a conversão
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gaubeur/desafio-posgraduacao-golang-fullcycle/desafio-clean-architecture/graph"
	"github.com/gaubeur/desafio-posgraduacao-golang-fullcycle/desafio-clean-architecture/internal/database"
	pb "github.com/gaubeur/desafio-posgraduacao-golang-fullcycle/desafio-clean-architecture/proto/pb"
)

const defaultPortGraphql = "8082"

type Order struct {
	ID            int64     `json:"id"`
	Customer_name string    `json:"customer_name"`
	Product_name  string    `json:"product_name"`
	Quantity      int64     `json:"quantity"`
	CreatedAt     time.Time `json:"createdAt"`
}

// OrderServiceServer é a implementação do nosso servidor gRPC.
// Ela precisa incorporar o 'unimplemented' para compatibilidade.
type OrderServiceServer struct {
	pb.UnimplementedOrderServiceServer
	db *sql.DB
}

// ListOrders é o método do serviço gRPC que implementa a lógica
// de listar os pedidos do banco de dados.
func (s *OrderServiceServer) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	log.Println("Requisição gRPC recebida para ListOrders")
	rows, err := s.db.QueryContext(ctx, "SELECT id, customer_name, product_name, quantity, createdAt FROM orders ORDER BY id DESC")
	if err != nil {
		log.Printf("Erro ao listar ordens (gRPC): %v", err)
		return nil, fmt.Errorf("erro ao listar ordens: %w", err)
	}
	defer rows.Close()

	var orders []*pb.Order
	for rows.Next() {
		var o Order
		var createdAt sql.NullTime
		if err := rows.Scan(&o.ID, &o.Customer_name, &o.Product_name, &o.Quantity, &createdAt); err != nil {
			log.Printf("Erro ao escanear ordem (gRPC): %v", err)
			return nil, fmt.Errorf("erro ao escanear ordem: %w", err)
		}

		if createdAt.Valid {
			o.CreatedAt = createdAt.Time
		} else {
			o.CreatedAt = time.Time{}
		}

		pbOrder := &pb.Order{
			Id:           o.ID,
			CustomerName: o.Customer_name,
			ProductName:  o.Product_name,
			Quantity:     o.Quantity,
			CreatedAt:    timestamppb.New(o.CreatedAt), // Conversão para *timestamp.Timestamp
		}
		orders = append(orders, pbOrder)
	}

	return &pb.ListOrdersResponse{Orders: orders}, nil
}

// createOrderHandler é o manipulador HTTP para criar pedidos via REST.
func createOrderHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var o Order

		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		_, err := db.ExecContext(ctx, "INSERT INTO orders (customer_name, product_name, quantity) VALUES (?, ?, ?)", o.Customer_name, o.Product_name, o.Quantity)
		if err != nil {
			log.Printf("Erro ao inserir ordem: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// listOrdersHandler é o manipulador HTTP para listar pedidos via REST.
func listOrdersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rows, err := db.QueryContext(ctx, "SELECT id, customer_name, product_name, quantity, createdAt FROM orders ORDER BY id DESC")
		if err != nil {
			log.Printf("Erro ao listar ordens: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var createdAt sql.NullTime

		var orders []Order
		for rows.Next() {
			var o Order
			if err := rows.Scan(&o.ID, &o.Customer_name, &o.Product_name, &o.Quantity, &createdAt); err != nil {
				log.Printf("Erro ao escanear ordem: %v", err)
				http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
				return
			}

			if createdAt.Valid {
				o.CreatedAt = createdAt.Time
			} else {
				o.CreatedAt = time.Time{}
			}
			orders = append(orders, o)
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(orders); err != nil {
			log.Printf("Erro ao codificar JSON: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		}
	}
}

func main() {
	// Obtém a URL do banco de dados a partir da variável de ambiente.
	//dbURL := os.Getenv("DB_URL")
	dbURL := "root:root@tcp(127.0.0.1:3307)/orders"
	if dbURL == "" {
		log.Fatal("A variável de ambiente DB_URL não foi definida")
	}

	// Conecta ao banco de dados.
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	// Valida a conexão com o banco de dados.
	if err = db.Ping(); err != nil {
		log.Fatalf("Não foi possível pingar o banco de dados: %v", err)
	}
	log.Println("Conexão com o banco de dados estabelecida com sucesso!")

	// ---------------- INICIALIZAÇÃO DO SERVIDOR GRPC ----------------
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Falha ao ouvir a porta 50051: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterOrderServiceServer(s, &OrderServiceServer{db: db})

		// para interagir com o evans
		reflection.Register(s)

		log.Println("Servidor gRPC iniciado na porta 50051...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Falha ao servir o gRPC: %v", err)
		}
	}()

	// ---------------- INICIALIZAÇÃO DO SERVIDOR GRAPHQL ----------------
	go func() {

		port := os.Getenv("PORT")
		if port == "" {
			port = defaultPortGraphql
		}

		orderDb := database.NewOrder(db)

		srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{OrderDB: orderDb}}))

		srv.AddTransport(transport.Options{})
		srv.AddTransport(transport.GET{})
		srv.AddTransport(transport.POST{})

		srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

		srv.Use(extension.Introspection{})
		srv.Use(extension.AutomaticPersistedQuery{
			Cache: lru.New[string](100),
		})

		http.Handle("/", playground.Handler("GraphQL playground", "/query"))
		http.Handle("/query", srv)

		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}()

	// ---------------- INICIALIZAÇÃO DO SERVIDOR REST ----------------
	router := mux.NewRouter()
	router.HandleFunc("/orders", createOrderHandler(db)).Methods("POST")
	router.HandleFunc("/orders", listOrdersHandler(db)).Methods("GET")
	log.Println("Servidor REST iniciado na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))

}
