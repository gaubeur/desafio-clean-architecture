set -e

host="1"shiftcmd="@"

until nc -z "$host" 3306; do
echo "Banco de dados indisponível - aguardando..."
sleep 1
done

echo "Banco de dados está pronto, executando comando..."
exec $cmd