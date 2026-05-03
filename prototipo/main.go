package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Mudei as constantes para valores menores para testes.
const RATE_LIMIT = 5
const TEMPO_ESPERA = 30 * time.Second

// Structs: os tipos vêm DEPOIS do nome da variável
type Credenciais struct {
	usuario string
	senha   string
	proxy   string
}

type Worker struct {
	mu            sync.Mutex
	lastDescanso  time.Time
	cred          *Credenciais
	rateAtual     int
	isTrabalhando bool
	isOcupado     bool
}

type Pool struct {
	mu         sync.Mutex // Essencial para não dar crash acessando o slice em concorrência
	workers    []*Worker
	numWorkers int
}

func NovaPool() *Pool {
	return &Pool{
		workers:    make([]*Worker, 0),
		numWorkers: 0,
	}
}

func NovoWorker(credenciais string) *Worker {
	// strings.Split precisa do pacote "strings"
	tempCreds := strings.Split(credenciais, ":")

	creds := &Credenciais{
		usuario: tempCreds[0],
		senha:   tempCreds[1],
		proxy:   tempCreds[2],
	}

	return &Worker{
		lastDescanso:  time.Now(),
		cred:          creds,
		rateAtual:     0,
		isTrabalhando: true,
		isOcupado:     false,
	}
}

func criarWorkers(credenciais []string) []*Worker {
	var workers []*Worker

	// Equivalente ao "for v in credenciais"
	for _, v := range credenciais {
		worker := NovoWorker(v)
		workers = append(workers, worker) // append é função, não método
	}

	return workers
}

func instanciarWorkersBulk(workers []*Worker, pool *Pool) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Adicionando vários itens de uma vez no slice
	pool.workers = append(pool.workers, workers...)
	pool.numWorkers += len(workers)
}

func moverWorker(worker *Worker, poolInicio *Pool, poolDestino *Pool) {
	poolInicio.mu.Lock()
	poolDestino.mu.Lock()
	defer poolInicio.mu.Unlock()
	defer poolDestino.mu.Unlock()

	// Go não tem .remove(), a gente recria o slice sem o elemento
	for i, w := range poolInicio.workers {
		if w == worker {
			poolInicio.workers = append(poolInicio.workers[:i], poolInicio.workers[i+1:]...)
			poolInicio.numWorkers--
			break
		}
	}

	poolDestino.workers = append(poolDestino.workers, worker)
	poolDestino.numWorkers++
}

func verificarRateLimit(worker *Worker, poolDescanso *Pool, poolTrabalhando *Pool) {
	// Parênteses no if não são necessários
	if worker.rateAtual >= RATE_LIMIT {
		worker.rateAtual = 0
		worker.lastDescanso = time.Now()
		worker.isTrabalhando = false
		worker.isOcupado = false
		moverWorker(worker, poolTrabalhando, poolDescanso)
		println("Rate limit atingido! Movendo para a pool de descanso...")
	}
}

// A função normal que vamos chamar com "go" depois
func checkarQuemTaMorcegando(poolDescanso *Pool, poolTrabalhando *Pool) {
	for { // while(true) do Go

		var workersParaMover []*Worker

		poolDescanso.mu.Lock()
		for _, worker := range poolDescanso.workers {
			worker.mu.Lock()
			// Subtração de tempo no Go é feita com time.Since
			if time.Since(worker.lastDescanso) >= TEMPO_ESPERA {
				worker.isTrabalhando = true
				workersParaMover = append(workersParaMover, worker)
				println("Tempo de descanso ja deu... movendo para a pool de trabalho")
			}
			worker.mu.Unlock()
		}
		poolDescanso.mu.Unlock()

		// Movemos fora do Lock principal para não travar a pool muito tempo
		for _, worker := range workersParaMover {
			moverWorker(worker, poolDescanso, poolTrabalhando)
		}

		time.Sleep(30 * time.Second) // O sleep precisa ficar DENTRO do loop
	}
}

// Funcao feita pra rodar em background, dando servico para o worker que estiver na pool de trabalho
func testeDeTrabalho(worker *Worker, poolDescanso *Pool, poolTrabalho *Pool, index int) {
	worker.mu.Lock()
	if !worker.isTrabalhando {
		worker.mu.Unlock()
		fmt.Printf("[WORKER %d] Estou descansando! %s - Rate limit atual: %d\n", index, time.Now().Format("15:04:05"), worker.rateAtual)
		return
	}
	worker.isOcupado = true
	worker.mu.Unlock()

	fmt.Printf("[WORKER %d] Trabalhando! %s - Rate limit atual: %d\n", index, time.Now().Format("15:04:05"), worker.rateAtual)
	time.Sleep(10 * time.Second)

	worker.mu.Lock()
	worker.rateAtual++
	worker.isOcupado = false
	worker.mu.Unlock()
	verificarRateLimit(worker, poolDescanso, poolTrabalho)
}

func darTrabalho(poolTrabalho *Pool, poolDescanso *Pool) {
	for {
		// Usando o Lock para evitar condicoes de corrida com outras funcoes que acessam a Pool
		poolTrabalho.mu.Lock()
		for i, worker := range poolTrabalho.workers {
			worker.mu.Lock()
			if worker.isOcupado {
				fmt.Printf("Worker %d esta ocupado.\n", i)
			} else {
				go testeDeTrabalho(worker, poolDescanso, poolTrabalho, i)
			}
			worker.mu.Unlock()
		}
		poolTrabalho.mu.Unlock()
		time.Sleep(4 * time.Second)
	}
}

func main() {
	credenciaisIniciais := []string{"usuario1:senha1:proxy1", "usuario2:senha2:proxy2", "usuario3:senha3:proxy3"}

	workers := criarWorkers(credenciaisIniciais)

	poolDescanso := NovaPool()
	poolTrabalhando := NovaPool()

	instanciarWorkersBulk(workers, poolTrabalhando)

	// Inicia a rotina em background
	go checkarQuemTaMorcegando(poolDescanso, poolTrabalhando)
	go darTrabalho(poolTrabalhando, poolDescanso)

	fmt.Printf("Trabalhadores instanciados: %d\n", poolTrabalhando.numWorkers)

	// Trava a main() para o programa não fechar imediatamente
	select {}
}
