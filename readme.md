Definição do desáfio
Agora é a hora de botar a mão na massa. Para este desafio, você precisará criar o usecase de listagem das orders.
Esta listagem precisa ser feita com:
- Endpoint REST (GET /order)
- Service ListOrders com GRPC
- Query ListOrders GraphQL
Não esqueça de criar as migrações necessárias e o arquivo api.http com a request para criar e listar as orders.

Para a criação do banco de dados, utilize o Docker (Dockerfile / docker-compose.yaml), com isso ao rodar o comando docker compose up tudo deverá subir, preparando o banco de dados.
Inclua um README.md com os passos a serem executados no desafio e a porta em que a aplicação deverá responder em cada serviço.

Definição de Portas:
Servidor REST iniciado na porta 8080
Servidor gRPC iniciado na porta 50051
http://localhost:8082/ for GraphQL playground na porta 8082


*** PARA SUBIR O AMBIENTE
docker-compose up

*** No aqruivo api.http temos exemplo de uso do rest / grpc / graphQL

*** Informações Extras

Relação de comandos utilizados para o desafio

-- comando para migrate
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install -tags '<driver_name>' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate create -ext=sql -dir=sql/migrations -seq init
migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3307)/orders" -verbose up
-- para desfazer
migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3307)/orders" -verbose down

-- isso só deve ser executado para tenha ocorrido um erro no processo de migrate
migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3307)/orders" force 1

-- serviço docker
-- tive que resolver um problema pois a porta 3306 já estava sendo usado pelo ambiente windows
docker-compose up -d
docker ps
sudo ss -lntp | grep 3306
docker stop <ID_ou_Nome_do_Container>
docker-compose down
docker-compose up -d

docker rm $(docker ps -aq) -f
docker ps -aq
docker rm -f
docker system prune --volumes -f

-- para entrar no docker
docker-compose exec app bash
.. app name of service


-- para abrir o banco de dados
docker-compose exec mysql bash

mysql -uroot -p orders
password : root

 -- para subir o serviço
 DB_URL="root:root@tcp(127.0.0.1:3307)/orders" go run main.go
 go run main.go [ para subir o serviço ]

 -- gerar o protoc
  protoc --go_out=. --go-grpc_out=. ./order.proto

-- para testar o serviço grpcser
-- fui obrigado a apontar o arquivo proto em função da importação do google no serviço
 evans --host localhost --port 50051 --proto ./proto/order.proto repl  

 -- para graphql
  go run github.com/99designs/gqlgen init
  go run github.com/99designs/gqlgen generate




